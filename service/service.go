package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/api/http/reader"
	"github.com/awakari/int-activitypub/util"
	"net/url"
	"strings"
	"time"

	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/awakari/int-activitypub/service/writer"
	"github.com/awakari/int-activitypub/storage"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
)

type Service interface {
	RequestFollow(ctx context.Context, addr, groupId, userId, subId, term string) (url string, err error)

	HandleActivity(
		ctx context.Context,
		actorIdLocal string,
		actor vocab.Actor,
		actorTags util.ObjectTags,
		activity vocab.Activity,
		activityTags util.ActivityTags,
	) (err error)

	Read(ctx context.Context, url vocab.IRI) (src model.Source, err error)

	List(
		ctx context.Context,
		filter model.Filter,
		limit uint32,
		cursor string,
		order model.Order,
	) (
		page []string,
		err error,
	)

	Unfollow(ctx context.Context, url vocab.IRI, groupId, userId string) (err error)
}

type service struct {
	stor      storage.Storage
	ap        activitypub.Service
	hostSelf  string
	conv      converter.Service
	w         writer.Service
	r         reader.Service
	cbUrlBase string
}

const lastUpdateThreshold = 1 * time.Hour

var ErrInvalid = errors.New("invalid argument")
var ErrNoAccept = errors.New("follow request is not accepted yet")
var ErrNoBot = errors.New(fmt.Sprintf("actor or activity contains the %s tag", NoBot))

func NewService(
	stor storage.Storage,
	ap activitypub.Service,
	hostSelf string,
	conv converter.Service,
	w writer.Service,
	r reader.Service,
	cbUrlBase string,
) Service {
	return service{
		stor:      stor,
		ap:        ap,
		hostSelf:  hostSelf,
		conv:      conv,
		w:         w,
		r:         r,
		cbUrlBase: cbUrlBase,
	}
}

func (svc service) RequestFollow(ctx context.Context, addr, groupId, userId, subId, term string) (addrResolved string, err error) {

	var addrParsed *url.URL
	addrParsed, err = url.Parse(addr)
	if err == nil {
		switch {
		case addrParsed.Scheme == "" && strings.Contains(addr, model.AcctSep):
			if len(addr) > 0 && strings.HasPrefix(addr, model.AcctSep) {
				addr = addr[1:]
			}
			acct := strings.SplitN(addr, model.AcctSep, 3)
			if len(acct) == 2 {
				name, host := acct[0], acct[1]
				var actorSelfLink vocab.IRI
				actorSelfLink, err = svc.ap.ResolveActorLink(ctx, host, name)
				if err == nil {
					addrResolved = actorSelfLink.String()
				}
			} else {
				err = fmt.Errorf("%w: invalid WebFinger handle: %s", ErrInvalid, addr)
			}
		default:
			addrResolved = addr
		}
	}

	var actor vocab.Actor
	var actorTags util.ObjectTags
	if err == nil {
		actor, actorTags, err = svc.ap.FetchActor(ctx, vocab.IRI(addrResolved))
		if err != nil {
			err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, addrResolved, err)
		}
	}
	if err == nil && ActorHasNoBotTag(actorTags) {
		err = fmt.Errorf("%w: actor %s", ErrNoBot, actor.ID)
	}

	var src model.Source
	if err == nil {
		src.ActorId = actor.ID.String()
		src.GroupId = groupId
		src.UserId = userId
		src.Type = string(actor.Type)
		src.Name = actor.Name.String()
		src.Summary = actor.Summary.String()
		src.Created = time.Now().UTC()
		src.Last = time.Now().UTC()
		src.SubId = subId
		src.Term = term
		err = svc.stor.Create(ctx, src)
		if err == nil {
			addrResolved = src.ActorId
		}
	}

	if err == nil {
		activity := vocab.Activity{
			Type:    vocab.FollowType,
			Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
			Actor:   vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf)),
			Object:  vocab.IRI(addrResolved),
		}
		err = svc.ap.SendActivity(ctx, activity, actor.Inbox.GetID())
		if err != nil {
			src.Err = err.Error()
			_ = svc.stor.Update(ctx, src)
		}
	}

	return
}

func (svc service) HandleActivity(
	ctx context.Context,
	actorIdLocal string,
	actor vocab.Actor,
	actorTags util.ObjectTags,
	activity vocab.Activity,
	activityTags util.ActivityTags,
) (
	err error,
) {
	actorId := actor.ID.String()
	switch activity.Type {
	case vocab.FollowType:
		err = svc.handleFollowActivity(ctx, actorIdLocal, actorId, activity)
	case vocab.UndoType:
		err = svc.handleUndoActivity(ctx, actorIdLocal, actorId, activity)
	default:
		err = svc.handleSourceActivity(ctx, actorId, actor, actorTags, activity, activityTags)
	}
	return
}

func (svc service) handleFollowActivity(ctx context.Context, actorIdLocal, actorId string, activity vocab.Activity) (err error) {
	cbUrl := svc.makeCallbackUrl(actorId)
	err = svc.r.CreateCallback(ctx, actorIdLocal, cbUrl)
	var actor vocab.Actor
	if err == nil {
		actor, _, err = svc.ap.FetchActor(ctx, vocab.IRI(actorId))
	}
	if err == nil {
		accept := vocab.AcceptNew(vocab.ID(actorIdLocal), activity.Object)
		err = svc.ap.SendActivity(ctx, *accept, actor.Inbox.GetID())
	}
	return
}

func (svc service) handleUndoActivity(ctx context.Context, actorIdLocal, actorId string, activity vocab.Activity) (err error) {
	switch activity.Object.GetType() {
	case vocab.FollowType:
		cbUrl := svc.makeCallbackUrl(actorId)
		err = svc.r.DeleteCallback(ctx, actorIdLocal, cbUrl)
	}
	return
}

func (svc service) makeCallbackUrl(actorId string) (cbUrl string) {
	cbUrl = svc.cbUrlBase + "?" + reader.QueryParamFollower + "=" + url.QueryEscape(actorId)
	return
}

func (svc service) handleSourceActivity(
	ctx context.Context,
	srcId string,
	actor vocab.Actor,
	actorTags util.ObjectTags,
	activity vocab.Activity,
	activityTags util.ActivityTags,
) (err error) {
	var src model.Source
	src, err = svc.stor.Read(ctx, srcId)
	switch {
	case err == nil:
		switch {
		case activity.Type == vocab.AcceptType:
			src.Accepted = true
			err = svc.stor.Update(ctx, src)
		case activity.Type == vocab.RejectType:
			src.Rejected = true
			err = svc.stor.Update(ctx, src)
		case ActorHasNoBotTag(actorTags):
			err = svc.stor.Delete(ctx, srcId, src.GroupId, src.UserId)
		case src.Accepted:
			var evt *pb.CloudEvent
			evt, _ = svc.conv.Convert(ctx, actor, activity, activityTags)
			if evt != nil && evt.Data != nil {
				t := time.Now().UTC()
				// don't update the storage on every activity but only when difference is higher than the threshold
				if src.Last.Add(lastUpdateThreshold).Before(t) {
					src.Last = time.Now().UTC()
					err = svc.stor.Update(ctx, src)
				}
				userId := src.UserId
				if userId == "" {
					userId = srcId
				}
				err = svc.w.Write(ctx, evt, src.GroupId, userId)
			}
		default:
			err = fmt.Errorf("%w: actor=%+v, activity.Type=%s", ErrNoAccept, actor, activity.Type)
		}
	case errors.Is(err, storage.ErrNotFound):
		err = svc.unfollow(ctx, actor.ID)
	}
	return
}

func (svc service) Read(ctx context.Context, url vocab.IRI) (a model.Source, err error) {
	a, err = svc.stor.Read(ctx, url.String())
	return
}

func (svc service) List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = svc.stor.List(ctx, filter, limit, cursor, order)
	return
}

func (svc service) Unfollow(ctx context.Context, url vocab.IRI, groupId, userId string) (err error) {
	err = svc.unfollow(ctx, url)
	if err == nil {
		err = svc.stor.Delete(ctx, url.String(), groupId, userId)
	}
	return
}

func (svc service) unfollow(ctx context.Context, url vocab.IRI) (err error) {
	var actor vocab.Actor
	actor, _, err = svc.ap.FetchActor(ctx, url)
	if err != nil {
		err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, url, err)
	}
	if err == nil {
		actorSelf := vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf))
		activity := vocab.Activity{
			Type:    vocab.UndoType,
			Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
			Actor:   actorSelf,
			Object: vocab.Activity{
				Type:   vocab.FollowType,
				Actor:  actorSelf,
				Object: url,
			},
		}
		err = svc.ap.SendActivity(ctx, activity, actor.Inbox.GetLink())
	}
	return
}
