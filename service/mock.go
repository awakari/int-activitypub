package service

import (
	"context"
	"github.com/awakari/int-activitypub/model"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) RequestFollow(ctx context.Context, addr string) (err error) {
	switch addr {
	case "fail":
		err = ErrInternal
	case "missing":
		err = ErrNotFound
	case "invalid":
		err = ErrInvalid
	case "conflict":
		err = ErrConflict
	}
	return
}

func (m mock) Read(ctx context.Context, addr string) (a model.Actor, err error) {
	switch addr {
	case "fail":
		err = ErrInternal
	case "missing":
		err = ErrNotFound
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
		err = ErrInternal
	default:
		page = []string{
			"user1@server1.social",
			"user2@server2.social",
		}
	}
	return
}

func (m mock) Unfollow(ctx context.Context, addr string) (err error) {
	switch addr {
	case "fail":
		err = ErrInternal
	case "missing":
		err = ErrNotFound
	case "invalid":
		err = ErrInvalid
	}
	return
}
