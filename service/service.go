package service

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/google/uuid"
)

type Service interface {
	RequestFollow(ctx context.Context, addr, groupId, userId, subId, term string) (url string, err error)
	HandleActivity(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (err error)
	Read(ctx context.Context, url vocab.IRI) (src model.Source, err error)
	List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error)
	Unfollow(ctx context.Context, url vocab.IRI, groupId, userId string) (err error)
}

var ErrInvalid = errors.New("invalid argument")
var ErrNoAccept = errors.New("follow request is not accepted yet")

type service struct {
	stor     storage.Storage
	ap       activitypub.Service
	hostSelf string
	conv     converter.Service
	w        writer.Service
}

const acctSep = "@"
const lastUpdateThreshold = 1 * time.Hour

func NewService(
	stor storage.Storage,
	ap activitypub.Service,
	hostSelf string,
	conv converter.Service,
	w writer.Service,
) Service {
	return service{
		stor:     stor,
		ap:       ap,
		hostSelf: hostSelf,
		conv:     conv,
		w:        w,
	}
}

func (svc service) RequestFollow(ctx context.Context, addr, groupId, userId, subId, term string) (addrResolved string, err error) {
	//
	var addrParsed *url.URL
	addrParsed, err = url.Parse(addr)
	if err == nil {
		switch {
		case addrParsed.Scheme == "" && strings.Contains(addr, acctSep):
			if len(addr) > 0 && strings.HasPrefix(addr, acctSep) {
				addr = addr[1:]
			}
			acct := strings.SplitN(addr, acctSep, 3)
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
	//
	var actor vocab.Actor
	if err == nil {
		actor, err = svc.ap.FetchActor(ctx, vocab.IRI(addrResolved))
		if err != nil {
			err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, addrResolved, err)
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
	}
	if err == nil {
		src := model.Source{
			ActorId: actor.ID.String(),
			GroupId: groupId,
			UserId:  userId,
			Type:    string(actor.Type),
			Name:    actor.Name.String(),
			Summary: actor.Summary.String(),
			Created: time.Now().UTC(),
			Last:    time.Now().UTC(),
			SubId:   subId,
			Term:    term,
		}
		err = svc.stor.Create(ctx, src)
		if err == nil {
			addrResolved = src.ActorId
		}
	}
	if err == nil {
		var reason string
		switch term {
		case "":
			reason = "an Awakari user manually requested this"
		default:
			reason = "Awakari discovered your public posts matching a user query"
		}
		actorSelf := vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf))
		activityFollowRequested := vocab.Activity{
			Type:    vocab.CreateType,
			Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
			Actor:   actorSelf,
			To: []vocab.Item{
				actor.GetLink(),
			},
			AttributedTo: actor.GetLink(),
			Object: vocab.Note{
				ID: vocab.ID(actorSelf.String() + "/" + uuid.NewString()),
				To: []vocab.Item{
					actor.GetLink(),
				},
				Type:      vocab.NoteType,
				Published: time.Now().UTC(),
				Content: vocab.DefaultNaturalLanguageValue(
					"<p>Hi " + actor.Name.String() + "!</p>" +
						"<p><a href=\"https://awakari.com\">Awakari</a> bot requests to follow you because " +
						reason + ".</p>" +
						"Contact: <a href=\"mailto:awakari@awakari.com\">awakari@awakari.com</a><br/>" +
						"Donate: <a href=\"https://t.me/donateawk/48\">https://t.me/donateawk/48</a><br/>" +
						"Privacy: <a href=\"https://awakari.com/privacy.html\">https://awakari.com/privacy.html</a><br/>" +
						"Source: <a href=\"https://github.com/awakari/int-activitypub\">https://github.com/awakari/int-activitypub</a><br/>" +
						"Terms: <a href=\"https://awakari.com/tos.html\">https://awakari.com/tos.html</a></p>",
				),
			},
		}
		err = svc.ap.SendActivity(ctx, activityFollowRequested, actor.Inbox.GetID())
	}
	return
}

func (svc service) HandleActivity(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (err error) {
	var src model.Source
	srcId := actor.ID.String()
	src, err = svc.stor.Read(ctx, srcId)
	if err == nil {
		switch {
		case activity.Type == vocab.AcceptType:
			actorSelf := vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf))
			var reasonFollowed string
			switch src.Term {
			case "":
				reasonFollowed = "an Awakari user manually requested this"
			default:
				reasonFollowed = "Awakari discovered your public posts matching a user query"
			}
			activityFollowAccepted := vocab.Activity{
				Type:    vocab.CreateType,
				Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
				Actor:   actorSelf,
				To: []vocab.Item{
					actor.GetLink(),
				},
				AttributedTo: actor.GetLink(),
				Object: vocab.Note{
					ID: vocab.ID(actorSelf.String() + "/" + uuid.NewString()),
					To: []vocab.Item{
						actor.GetLink(),
					},
					Type:      vocab.NoteType,
					Published: time.Now().UTC(),
					Content: vocab.DefaultNaturalLanguageValue(
						"<p>Hi " + actor.Name.String() + "!</p>" +
							"<p><a href=\"https://awakari.com\">Awakari</a> bot followed you because " +
							reasonFollowed + ". The follow request has been <b>accepted</b>.</p>" +
							"Note this acceptance means your <i>explicit consent</i> to process your public (only) posts.</p>" +
							"<p>If you don't agree with the following, please remove the bot from your followers. " +
							"Additionally, you can disable automatic follow request acceptance.</p>" +
							"Contact: <a href=\"mailto:awakari@awakari.com\">awakari@awakari.com</a><br/>" +
							"Donate: <a href=\"https://t.me/donateawk/48\">https://t.me/donateawk/48</a><br/>" +
							"Privacy: <a href=\"https://awakari.com/privacy.html\">https://awakari.com/privacy.html</a><br/>" +
							"Source: <a href=\"https://github.com/awakari/int-activitypub\">https://github.com/awakari/int-activitypub</a><br/>" +
							"Terms: <a href=\"https://awakari.com/tos.html\">https://awakari.com/tos.html</a></p>",
					),
				},
			}
			err = svc.ap.SendActivity(ctx, activityFollowAccepted, actor.Inbox.GetLink())
			if err == nil {
				src.Accepted = true
				err = svc.stor.Update(ctx, src)
			}
		case src.Accepted:
			var evt *pb.CloudEvent
			evt, _ = svc.conv.Convert(ctx, actor, activity)
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
	var actor vocab.Actor
	if err == nil {
		actor, err = svc.ap.FetchActor(ctx, url)
		if err != nil {
			err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, url, err)
		}
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
	if err == nil {
		err = svc.stor.Delete(ctx, url.String(), groupId, userId)
	}
	return
}
