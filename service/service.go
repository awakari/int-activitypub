package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/awakari/int-activitypub/service/writer"
	"github.com/awakari/int-activitypub/storage"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
)

type Service interface {
	RequestFollow(ctx context.Context, addr, groupId, userId string) (err error)
	HandleActivity(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (err error)
	Read(ctx context.Context, url vocab.IRI) (src model.Source, err error)
	List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error)
	Unfollow(ctx context.Context, url vocab.IRI, groupId, userId string) (err error)
}

var ErrInvalid = errors.New("invalid argument")

type service struct {
	stor     storage.Storage
	ap       activitypub.Service
	hostSelf string
	conv     converter.Service
	w        writer.Service
}

const acctSep = "@"

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

func (svc service) RequestFollow(ctx context.Context, addr, groupId, userId string) (err error) {
	_, err = svc.stor.Read(ctx, addr)
	switch {
	case err == nil:
		err = fmt.Errorf("%w: %s", storage.ErrConflict, addr)
	case errors.Is(err, storage.ErrNotFound):
		err = nil
	}
	var actor vocab.Actor
	if err == nil {
		actor, err = svc.ap.FetchActor(ctx, vocab.IRI(addr))
		if err != nil {
			err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, addr, err)
		}
	}
	if err == nil {
		activity := vocab.Activity{
			Type:    vocab.FollowType,
			Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
			Actor:   vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf)),
			Object:  vocab.IRI(addr),
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
		}
		err = svc.stor.Create(ctx, src)
	}
	return
}

func (svc service) HandleActivity(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (err error) {
	var src model.Source
	src, err = svc.stor.Read(ctx, actor.ID.String())
	if err == nil {
		switch activity.Type {
		case vocab.AcceptType:
			src.Accepted = true
			err = svc.stor.Update(ctx, src)
		default:
			var evt *pb.CloudEvent
			evt, _ = svc.conv.Convert(ctx, actor, activity)
			if evt != nil && evt.Data != nil {
				err = svc.w.Write(ctx, evt, src.GroupId, src.UserId)
			}
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
