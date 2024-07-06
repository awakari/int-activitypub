package service

import (
	"context"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/storage"
	"github.com/awakari/int-activitypub/util"
	vocab "github.com/go-ap/activitypub"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) RequestFollow(ctx context.Context, addr, groupId, userId, subId, term string) (url string, err error) {
	switch addr {
	case "activitypub_fail":
		err = activitypub.ErrActivitySend
	case "invalid":
		err = ErrInvalid
	case "conflict":
		err = storage.ErrConflict
	case "fail":
		err = storage.ErrInternal
	default:
		url = addr
	}
	return
}

func (m mock) HandleActivity(ctx context.Context, actorIdLocal string, actor vocab.Actor, actorTags util.ObjectTags, activity vocab.Activity, tags util.ActivityTags) (err error) {
	switch actor.ID {
	case "fail":
		err = storage.ErrInternal
	case "missing":
		err = storage.ErrNotFound
	}
	return
}

func (m mock) Read(ctx context.Context, url vocab.IRI) (a model.Source, err error) {
	switch url {
	case "fail":
		err = storage.ErrInternal
	case "missing":
		err = storage.ErrNotFound
	default:
		a.ActorId = "user1@server1.social"
		a.UserId = "user2"
		a.GroupId = "group1"
		a.Name = "John Doe"
		a.Type = "Person"
		a.Summary = "yohoho"
	}
	return
}

func (m mock) List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
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

func (m mock) Unfollow(ctx context.Context, url vocab.IRI, groupId, userId string) (err error) {
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
