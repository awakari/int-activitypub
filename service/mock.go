package service

import (
	"context"
	"github.com/awakari/int-activitypub/api/http/activitypub"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/storage"
	vocab "github.com/go-ap/activitypub"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) RequestFollow(ctx context.Context, addr string) (url vocab.IRI, err error) {
	url = vocab.IRI(addr)
	switch addr {
	case "activitypub_fail":
		err = activitypub.ErrActivitySend
	case "invalid":
		err = ErrInvalid
	case "conflict":
		err = storage.ErrConflict
	case "fail":
		err = storage.ErrInternal
	}
	return
}

func (m mock) HandleActivity(ctx context.Context, url vocab.IRI, activity vocab.Activity) (err error) {
	switch url {
	case "storage_fail":
		err = storage.ErrInternal
	case "conflict":
		err = storage.ErrConflict
	}
	return
}

func (m mock) Read(ctx context.Context, url vocab.IRI) (a model.Actor, err error) {
	switch url {
	case "fail":
		err = storage.ErrInternal
	case "missing":
		err = storage.ErrNotFound
	default:
		a.Addr = "user1@server1.social"
		a.UserId = "user2"
		a.GroupId = "group1"
		a.Name = "John Doe"
		a.Type = "Person"
		a.Summary = "yohoho"
	}
	return
}

func (m mock) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	switch cursor {
	case "fail":
		err = storage.ErrInternal
	default:
		page = []string{
			"user1@server1.social",
			"user2@server2.social",
		}
	}
	return
}

func (m mock) Unfollow(ctx context.Context, url vocab.IRI) (err error) {
	switch url {
	case "fail":
		err = storage.ErrInternal
	case "missing":
		err = storage.ErrNotFound
	case "invalid":
		err = ErrInvalid
	case "activitypub_fail":
		err = activitypub.ErrActivitySend
	}
	return
}
