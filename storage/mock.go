package storage

import (
	"context"
	"github.com/awakari/int-activitypub/model"
)

type mock struct {
}

func NewStorageMock() Storage {
	return mock{}
}

func (s mock) Close() error {
	return nil
}

func (s mock) Create(ctx context.Context, addr string) (err error) {
	switch addr {
	case "fail":
		err = ErrInternal
	case "conflict":
		err = ErrConflict
	}
	return
}

func (s mock) Read(ctx context.Context, addr string) (a model.Actor, err error) {
	switch addr {
	case "https://host.social/users/storfail":
		err = ErrInternal
	case "https://host.social/users/existing":
		a.Addr = "user1@server1.social"
		a.UserId = "user2"
		a.GroupId = "group1"
		a.Name = "John Doe"
		a.Type = "Person"
		a.Summary = "yohoho"
	default:
		err = ErrNotFound
	}
	return
}

func (s mock) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	switch cursor {
	case "fail":
		err = ErrInternal
	default:
		page = []string{
			"user1@server1.social",
			"user2@server2.social",
		}
	}
	return
}

func (s mock) Delete(ctx context.Context, addr string) (err error) {
	switch addr {
	case "fail":
		err = ErrInternal
	case "missing":
		err = ErrNotFound
	}
	return
}
