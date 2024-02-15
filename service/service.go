package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/api/http/activitypub"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/storage"
	vocab "github.com/go-ap/activitypub"
	"strings"
)

type Service interface {
	RequestFollow(ctx context.Context, addr string) (url vocab.IRI, err error)
	HandleActivity(ctx context.Context, url vocab.IRI, activity vocab.Activity) (err error)
	Read(ctx context.Context, url vocab.IRI) (a model.Actor, err error)
	List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error)
	Unfollow(ctx context.Context, url vocab.IRI) (err error)
}

var ErrInvalid = errors.New("invalid argument")

type service struct {
	stor           storage.Storage
	svcActivityPub activitypub.Service
	hostSelf       string
}

const acctSep = "@"

func NewService(stor storage.Storage, svcActivityPub activitypub.Service, hostSelf string) Service {
	return service{
		stor:           stor,
		svcActivityPub: svcActivityPub,
		hostSelf:       hostSelf,
	}
}

func (svc service) RequestFollow(ctx context.Context, addr string) (url vocab.IRI, err error) {
	acct := strings.SplitN(addr, acctSep, 3)
	if len(acct) != 2 {
		err = fmt.Errorf("%w address to follow: %s, should be <name>@<host>", ErrInvalid, addr)
	}
	var host, name string
	if err == nil {
		name, host = acct[0], acct[1]
		if name == "" || host == "" {
			err = fmt.Errorf("%w address to follow: %s, should be <name>@<host>", ErrInvalid, addr)
		}
	}
	if err == nil {
		url, err = svc.svcActivityPub.ResolveActorLink(ctx, host, name)
		if err != nil {
			err = fmt.Errorf("%w: failed to resolve the actor %s@%s, cause: %s", ErrInvalid, name, host, err)
		}
	}
	if err == nil {
		_, err = svc.stor.Read(ctx, url.String())
		switch {
		case err == nil:
			err = fmt.Errorf("%w: %s", storage.ErrConflict, url)
		case errors.Is(err, storage.ErrNotFound):
			err = nil
		}
	}
	var actor vocab.Actor
	if err == nil {
		actor, err = svc.svcActivityPub.FetchActor(ctx, url)
		if err != nil {
			err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, url, err)
		}
	}
	if err == nil {
		activity := vocab.Activity{
			Type:    vocab.FollowType,
			Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
			Actor:   vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf)),
			Object:  vocab.IRI(addr),
		}
		err = svc.svcActivityPub.SendActivity(ctx, activity, actor.Inbox.GetLink())
	}
	return
}

func (svc service) HandleActivity(ctx context.Context, url vocab.IRI, activity vocab.Activity) (err error) {
	switch activity.Type {
	case vocab.AcceptType:
		err = svc.stor.Create(ctx, url.String())
	default:
		// TODO convert to event and submit it
	}
	return
}

func (svc service) Read(ctx context.Context, url vocab.IRI) (a model.Actor, err error) {
	a, err = svc.stor.Read(ctx, url.String())
	return
}

func (svc service) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = svc.stor.List(ctx, filter, limit, cursor, order)
	return
}

func (svc service) Unfollow(ctx context.Context, url vocab.IRI) (err error) {
	var actor vocab.Actor
	if err == nil {
		actor, err = svc.svcActivityPub.FetchActor(ctx, url)
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
		err = svc.svcActivityPub.SendActivity(ctx, activity, actor.Inbox.GetLink())
	}
	if err == nil {
		err = svc.stor.Delete(ctx, url.String())
	}
	return
}
