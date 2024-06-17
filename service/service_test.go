package service

import (
	"context"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/awakari/int-activitypub/service/writer"
	"github.com/awakari/int-activitypub/storage"
	"github.com/awakari/int-activitypub/util"
	vocab "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestService_RequestFollow(t *testing.T) {
	svc := NewService(
		storage.NewStorageMock(),
		activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()),
		"test.social",
		converter.NewLogging(converter.NewService("foo"), slog.Default()),
		writer.NewLogging(writer.NewMock(), slog.Default()),
	)
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		addr string
		url  string
		err  error
	}{
		"ok": {
			addr: "johndoe@host.social",
			url:  "https://host.social/users/johndoe",
		},
		"invalid src addr": {
			addr: "@host.social",
			err:  ErrInvalid,
		},
		"fail resolve webfinger": {
			addr: "fail@host.social",
			err:  activitypub.ErrActorWebFinger,
		},
		"fail to fetch actor": {
			addr: "https://fail.social/users/johndoe",
			url:  "https://fail.social/users/johndoe",
			err:  ErrInvalid,
		},
		"fail to send activity": {
			addr: "https://host.fail/users/johndoe",
			url:  "https://host.fail/users/johndoe",
			err:  activitypub.ErrActivitySend,
		},
		"conflict": {
			addr: "conflict",
			url:  "conflict",
			err:  storage.ErrConflict,
		},
		"storage fails": {
			addr: "fail",
			url:  "fail",
			err:  storage.ErrInternal,
		},
		"nobot1": {
			addr: "https://privacy.social/users/nobot2",
			url:  "https://privacy.social/users/nobot2",
			err:  ErrNoBot,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			u, err := svc.RequestFollow(context.TODO(), c.addr, "group0", "user1", "", "")
			assert.Equal(t, c.url, u)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_HandleActivity(t *testing.T) {
	svc := NewService(
		storage.NewStorageMock(),
		activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()),
		"test.social",
		converter.NewLogging(converter.NewService("foo"), slog.Default()),
		writer.NewLogging(writer.NewMock(), slog.Default()),
	)
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		url      vocab.IRI
		activity vocab.Activity
		err      error
	}{
		"ok": {
			url: "https://host.social/users/existing",
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.HandleActivity(context.TODO(), vocab.Actor{ID: c.url}, util.ObjectTags{}, c.activity, util.ActivityTags{})
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_Read(t *testing.T) {
	svc := NewService(
		storage.NewStorageMock(),
		activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()),
		"test.social",
		converter.NewLogging(converter.NewService("foo"), slog.Default()),
		writer.NewLogging(writer.NewMock(), slog.Default()),
	)
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		url   vocab.IRI
		actor model.Source
		err   error
	}{
		"ok": {
			url:   "https://host.social/users/existing",
			actor: model.Source{ActorId: "user1@server1.social", GroupId: "group1", UserId: "user2", Type: "Person", Name: "John Doe", Summary: "yohoho", Accepted: true},
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
	svc := NewService(
		storage.NewStorageMock(),
		activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()),
		"test.social",
		converter.NewLogging(converter.NewService("foo"), slog.Default()),
		writer.NewLogging(writer.NewMock(), slog.Default()),
	)
	svc = NewLogging(svc, slog.Default())
	cases := map[string]struct {
		filter model.Filter
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
	svc := NewService(
		storage.NewStorageMock(),
		activitypub.NewServiceLogging(activitypub.NewServiceMock(), slog.Default()),
		"test.social",
		converter.NewLogging(converter.NewService("foo"), slog.Default()),
		writer.NewLogging(writer.NewMock(), slog.Default()),
	)
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
			err := svc.Unfollow(context.TODO(), c.url, "group0", "user1")
			assert.ErrorIs(t, err, c.err)
		})
	}
}
