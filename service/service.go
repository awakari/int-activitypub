package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/api/http/pub"
	"github.com/awakari/int-activitypub/api/http/reader"
	"github.com/awakari/int-activitypub/util"
	"github.com/bytedance/sonic"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"net/url"
	"strings"
	"time"

	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/awakari/int-activitypub/storage"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
)

type Service interface {
	RequestFollow(ctx context.Context, addr, groupId, userId, interestId, term string, defaultActor bool) (url string, err error)

	HandleActivity(
		ctx context.Context,
		actorIdLocal, pubKeyId string,
		actor vocab.Actor,
		actorTags util.ObjectTags,
		activity vocab.Activity,
		activityTags util.ActivityTags,
	) (
		post func(),
		err error,
	)

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
	stor             storage.Storage
	ap               activitypub.Service
	hostSelf         string
	conv             converter.Service
	svcPub           pub.Service
	backoffTimeLimit time.Duration
	r                reader.Service
	cbUrlBase        string
}

const lastUpdateThreshold = 1 * time.Hour
const backoffInitDelay = 100 * time.Millisecond

var ErrInvalid = errors.New("invalid argument")
var ErrNoAccept = errors.New("follow request is not accepted yet")
var ErrNoBot = errors.New(fmt.Sprintf("actor or activity contains the %s tag", NoBot))

func NewService(
	stor storage.Storage,
	ap activitypub.Service,
	hostSelf string,
	conv converter.Service,
	svcPub pub.Service,
	backoffTimeLimit time.Duration,
	r reader.Service,
	cbUrlBase string,
) Service {
	return service{
		stor:             stor,
		ap:               ap,
		hostSelf:         hostSelf,
		conv:             conv,
		svcPub:           svcPub,
		backoffTimeLimit: backoffTimeLimit,
		r:                r,
		cbUrlBase:        cbUrlBase,
	}
}

func (svc service) RequestFollow(ctx context.Context, addr, groupId, userId, interestId, term string, defaultActor bool) (addrResolved string, err error) {

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

	if err == nil {
		addrParsed, err = url.Parse(addrResolved)
		if err == nil && addrParsed.Host == svc.hostSelf {
			err = fmt.Errorf("%w: attempt to follow the self hosted actor %s", ErrInvalid, addr)
		}
	}

	var pubKeyId string
	switch defaultActor {
	case true:
		pubKeyId = fmt.Sprintf("https://%s/actor#main-key", svc.hostSelf)
	default:
		pubKeyId = fmt.Sprintf("https://%s/actor/%s#main-key", svc.hostSelf, interestId)
	}

	var target vocab.Actor
	var targetTags util.ObjectTags
	if err == nil {
		target, targetTags, err = svc.ap.FetchActor(ctx, vocab.IRI(addrResolved), pubKeyId)
		if err != nil {
			err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, addrResolved, err)
		}
	}
	if err == nil && ActorHasNoBotTag(targetTags) {
		err = fmt.Errorf("%w: actor %s", ErrNoBot, target.ID)
	}

	var src model.Source
	if err == nil && defaultActor {
		src.ActorId = target.ID.String()
		src.GroupId = groupId
		src.UserId = userId
		src.Type = string(target.Type)
		src.Name = target.Name.String()
		src.Summary = target.Summary.String()
		src.Created = time.Now().UTC()
		src.Last = time.Now().UTC()
		src.SubId = interestId
		src.Term = term
		err = svc.stor.Create(ctx, src)
		if err == nil {
			addrResolved = src.ActorId
		}
	}

	if err == nil {
		activity := vocab.Activity{
			Type:    vocab.FollowType,
			Context: vocab.IRI(model.NsAs),
			Object:  vocab.IRI(addrResolved),
		}
		switch defaultActor {
		case true:
			activity.Actor = vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf))
		default:
			activity.Actor = vocab.IRI(fmt.Sprintf("https://%s/actor/%s", svc.hostSelf, interestId))
		}
		err = svc.ap.SendActivity(ctx, activity, target.Inbox.GetLink(), pubKeyId)
		if err != nil && defaultActor {
			src.Err = err.Error()
			_ = svc.stor.Update(ctx, src)
		}
	}

	return
}

func (svc service) HandleActivity(
	ctx context.Context,
	actorIdLocal, pubKeyId string,
	actor vocab.Actor,
	actorTags util.ObjectTags,
	activity vocab.Activity,
	activityTags util.ActivityTags,
) (
	post func(),
	err error,
) {
	actorId := actor.ID.String()
	switch activity.Type {
	case vocab.FollowType:
		post, err = svc.handleFollowActivity(ctx, actorIdLocal, pubKeyId, actorId, activity)
	case vocab.UndoType:
		err = svc.handleUndoActivity(ctx, actorIdLocal, actorId, activity)
	default:
		err = svc.handleSourceActivity(ctx, actorId, pubKeyId, actor, actorTags, activity, activityTags)
	}
	return
}

func (svc service) handleFollowActivity(ctx context.Context, actorIdLocal, pubKeyId, actorId string, activity vocab.Activity) (post func(), err error) {
	d, _ := sonic.Marshal(activity)
	fmt.Printf("Follow activity payload: %s\n", d)
	cbUrl := svc.makeCallbackUrl(actorId)
	err = svc.r.Subscribe(ctx, actorIdLocal, model.GroupIdDefault, model.UserIdDefault, cbUrl, 0)
	var actor vocab.Actor
	if err == nil {
		actor, _, err = svc.ap.FetchActor(ctx, vocab.IRI(actorId), pubKeyId)
	}
	if err == nil {
		post = func() {
			time.Sleep(10 * time.Second)
			accept := vocab.AcceptNew(vocab.IRI(fmt.Sprintf("https://%s/%s", svc.hostSelf, uuid.NewString())), activity.Object)
			accept.Context = vocab.IRI(model.NsAs)
			accept.Actor = vocab.ID(fmt.Sprintf("https://%s/actor/%s", svc.hostSelf, actorIdLocal))
			accept.Object = activity
			_ = svc.ap.SendActivity(ctx, *accept, actor.Inbox.GetLink(), pubKeyId)
		}
	}
	return
}

func (svc service) handleUndoActivity(ctx context.Context, actorIdLocal, actorId string, activity vocab.Activity) (err error) {
	switch activity.Object.GetType() {
	case vocab.FollowType:
		cbUrl := svc.makeCallbackUrl(actorId)
		err = svc.r.Unsubscribe(ctx, actorIdLocal, model.GroupIdDefault, model.UserIdDefault, cbUrl)
	}
	return
}

func (svc service) makeCallbackUrl(actorId string) (cbUrl string) {
	cbUrl = svc.cbUrlBase + "?" + reader.QueryParamFollower + "=" + url.QueryEscape(actorId)
	return
}

func (svc service) handleSourceActivity(
	ctx context.Context,
	srcId, pubKeyId string,
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
			evt, _ = svc.conv.ConvertActivityToEvent(ctx, actor, activity, activityTags)
			if evt != nil && evt.Data != nil {
				t := time.Now().UTC()
				// don't update the storage on every activity but only when difference is higher than the threshold
				if src.Last.Add(lastUpdateThreshold).Before(t) {
					src.Last = time.Now().UTC()
					err = svc.stor.Update(ctx, src)
				}
				userId := src.UserId
				if userId == "" {
					userId = evt.Source
				}
				err = svc.svcPub.Publish(ctx, evt, src.GroupId, userId)
				if errors.Is(err, pub.ErrNoAck) {
					b := backoff.NewExponentialBackOff()
					b.InitialInterval = backoffInitDelay
					b.MaxElapsedTime = svc.backoffTimeLimit
					err = backoff.Retry(func() error {
						return svc.svcPub.Publish(ctx, evt, src.GroupId, userId)
					}, b)
				}
			}
		default:
			err = fmt.Errorf("%w: actor=%+v, activity.Type=%s", ErrNoAccept, actor, activity.Type)
		}
	case errors.Is(err, storage.ErrNotFound):
		err = svc.unfollow(ctx, actor.ID, pubKeyId)
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
	err = svc.unfollow(ctx, url, fmt.Sprintf("https://%s/actor#main-key", svc.hostSelf))
	err = errors.Join(err, svc.stor.Delete(ctx, url.String(), groupId, userId))
	return
}

func (svc service) unfollow(ctx context.Context, url vocab.IRI, pubKeyId string) (err error) {
	var actor vocab.Actor
	actor, _, err = svc.ap.FetchActor(ctx, url, pubKeyId)
	if err != nil {
		err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, url, err)
	}
	if err == nil {
		actorSelf := vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf))
		activity := vocab.Activity{
			Type:    vocab.UndoType,
			Context: vocab.IRI(model.NsAs),
			Actor:   actorSelf,
			Object: vocab.Activity{
				Type:   vocab.FollowType,
				Actor:  actorSelf,
				Object: url,
			},
		}
		err = svc.ap.SendActivity(ctx, activity, actor.Inbox.GetLink(), pubKeyId)
	}
	return
}
