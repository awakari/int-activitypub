package service

import (
	"context"
	"github.com/awakari/int-activitypub/api/http/activitypub"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/storage"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestService_RequestFollow(t *testing.T) {
	svc := NewService(storage.NewStorageMock(), activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()), "test.social")
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		addr string
		url  vocab.IRI
		err  error
	}{
		"ok": {
			addr: "johndoe@host.social",
			url:  "https://host.social/users/johndoe",
		},
		"invalid src addr1": {
			addr: "johndoe",
			err:  ErrInvalid,
		},
		"invalid src addr2": {
			addr: "@host.social",
			err:  ErrInvalid,
		},
		"fail resolve webfinger": {
			addr: "fail@host.social",
			err:  ErrInvalid,
		},
		"fail to fetch actor": {
			addr: "johndoe@fail.social",
			url:  "https://fail.social/users/johndoe",
			err:  ErrInvalid,
		},
		"fail to send activity": {
			addr: "johndoe@host.fail",
			url:  "https://host.fail/users/johndoe",
			err:  activitypub.ErrActivitySend,
		},
		"conflict": {
			addr: "existing@host.social",
			url:  "https://host.social/users/existing",
			err:  storage.ErrConflict,
		},
		"storage fails": {
			addr: "storfail@host.social",
			url:  "https://host.social/users/storfail",
			err:  storage.ErrInternal,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			u, err := svc.RequestFollow(context.TODO(), c.addr)
			assert.Equal(t, c.url, u)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_HandleActivity(t *testing.T) {
	svc := NewService(storage.NewStorageMock(), activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()), "test.social")
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		url      vocab.IRI
		activity vocab.Activity
		err      error
	}{
		"ok": {
			url: "ok",
		},
		"accept fail": {
			url: "fail",
			activity: vocab.Activity{
				Type: vocab.AcceptType,
			},
			err: storage.ErrInternal,
		},
		"accept conflict": {
			url: "conflict",
			activity: vocab.Activity{
				Type: vocab.AcceptType,
			},
			err: storage.ErrConflict,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.HandleActivity(context.TODO(), c.url, c.activity)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_Read(t *testing.T) {
	svc := NewService(storage.NewStorageMock(), activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()), "test.social")
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		url   vocab.IRI
		actor model.Actor
		err   error
	}{
		"ok": {
			url:   "https://host.social/users/existing",
			actor: model.Actor{Addr: "user1@server1.social", GroupId: "group1", UserId: "user2", Type: "Person", Name: "John Doe", Summary: "yohoho"},
		},
		"fail": {
			url: "https://host.social/users/storfail",
			err: storage.ErrInternal,
		},
		"missing": {
			err: storage.ErrNotFound,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			actor, err := svc.Read(context.TODO(), c.url)
			assert.Equal(t, c.actor, actor)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_List(t *testing.T) {
	svc := NewService(storage.NewStorageMock(), activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()), "test.social")
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		filter model.ActorFilter
		limit  uint32
		cursor string
		order  model.Order
		page   []string
		err    error
	}{
		"ok": {
			page: []string{
				"user1@server1.social",
				"user2@server2.social",
			},
		},
		"fail": {
			cursor: "fail",
			err:    storage.ErrInternal,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			page, err := svc.List(context.TODO(), c.filter, c.limit, c.cursor, c.order)
			assert.Equal(t, c.page, page)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_Unfollow(t *testing.T) {
	svc := NewService(storage.NewStorageMock(), activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()), "test.social")
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		url vocab.IRI
		err error
	}{
		"ok": {},
		"fails to fetch actor": {
			url: "https://fail.social/users/johndoe",
			err: ErrInvalid,
		},
		"fails to send activity": {
			url: "https://host.fail/users/johndoe",
			err: activitypub.ErrActivitySend,
		},
		"missing": {
			url: "missing",
			err: storage.ErrNotFound,
		},
		"fail": {
			url: "fail",
			err: storage.ErrInternal,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.Unfollow(context.TODO(), c.url)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
