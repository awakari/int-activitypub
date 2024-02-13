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

var ErrInternal = errors.New("internal failure")
var ErrInvalid = errors.New("invalid argument")
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("already exists")

type service struct {
	stor           storage.Storage
	svcActivityPub activitypub.Service
}

const acctSep = "@"

func NewService(stor storage.Storage, svcActivityPub activitypub.Service) Service {
	return service{
		stor:           stor,
		svcActivityPub: svcActivityPub,
	}
}

func (svc service) RequestFollow(ctx context.Context, addr string) (url vocab.IRI, err error) {
	acct := strings.SplitN(addr, acctSep, 3)
	if len(acct) != 2 {
		err = fmt.Errorf("%s address to follow: %s, should be <name>@<host>", ErrInvalid, addr)
	}
	var host, name string
	if err == nil {
		name, host = acct[0], acct[1]
		if name == "" || host == "" {
			err = fmt.Errorf("%s address to follow: %s, should be <name>@<host>", ErrInvalid, addr)
		}
	}
	if err == nil {
		url, err = svc.svcActivityPub.ResolveActorLink(ctx, host, name)
		if err != nil {
			err = fmt.Errorf("%w: failed to resolve the actor %s@%s, cause: %s", ErrInvalid, name, host, err)
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
		err = svc.svcActivityPub.RequestFollow(ctx, host, url, actor.Inbox.GetLink())
	}
	return
}

func (svc service) HandleActivity(ctx context.Context, url vocab.IRI, activity vocab.Activity) (err error) {
	switch activity.Type {
	case vocab.AcceptType:
		err = svc.stor.Create(ctx, url.String())
	default:

	}
	return
}

func (svc service) Read(ctx context.Context, url vocab.IRI) (a model.Actor, err error) {
	//TODO implement me
	panic("implement me")
}

func (svc service) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	//TODO implement me
	panic("implement me")
}

func (svc service) Unfollow(ctx context.Context, url vocab.IRI) (err error) {
	//TODO implement me
	panic("implement me")
}
