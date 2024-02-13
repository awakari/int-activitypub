package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/storage"
	vocab "github.com/go-ap/activitypub"
	"strings"
)

type Service interface {
	RequestFollow(ctx context.Context, addr string) (err error)
	Read(ctx context.Context, addr string) (a model.Actor, err error)
	List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error)
	Unfollow(ctx context.Context, addr string) (err error)
}

var ErrInternal = errors.New("internal failure")
var ErrInvalid = errors.New("invalid argument")
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("already exists")

type service struct {
	stor           storage.Storage
	svcActivityPub http.Service
}

const acctSep = "@"

func NewService(stor storage.Storage, svcActivityPub http.Service) Service {
	return service{
		stor:           stor,
		svcActivityPub: svcActivityPub,
	}
}

func (svc service) RequestFollow(ctx context.Context, addr string) (err error) {
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
	var obj vocab.IRI
	if err == nil {
		obj, err = svc.svcActivityPub.ResolveActorLink(ctx, host, name)
		if err != nil {
			err = fmt.Errorf("%w: failed to resolve the actor %s@%s, cause: %s", ErrInvalid, name, host, err)
		}
	}
	var actor vocab.Actor
	if err == nil {
		actor, err = svc.svcActivityPub.FetchActor(ctx, obj)
		if err != nil {
			err = fmt.Errorf("%w: failed to fetch actor: %s, cause: %s", ErrInvalid, obj, err)
		}
	}
	if err == nil {
		err = svc.svcActivityPub.RequestFollow(ctx, host, obj, actor.Inbox.GetLink())
	}
	return
}

func (svc service) Read(ctx context.Context, addr string) (a model.Actor, err error) {
	//TODO implement me
	panic("implement me")
}

func (svc service) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	//TODO implement me
	panic("implement me")
}

func (svc service) Unfollow(ctx context.Context, addr string) (err error) {
	//TODO implement me
	panic("implement me")
}
