package service

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/model"
	"log/slog"
)

type logging struct {
	svc Service
	log *slog.Logger
}

func NewLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) Follow(ctx context.Context, addr string) (err error) {
	err = l.svc.Follow(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.Follow(addr=%s): %s", addr, err))
	return
}

func (l logging) Read(ctx context.Context, addr string) (a model.Actor, err error) {
	a, err = l.svc.Read(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.Read(addr=%s): %+v, %s", addr, a, err))
	return
}

func (l logging) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = l.svc.List(ctx, filter, limit, cursor, order)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.List(filter=%+v, limit=%d, cursor=%s, order=%s): %d, %s", filter, limit, cursor, order, len(page), err))
	return
}

func (l logging) Unfollow(ctx context.Context, addr string) (err error) {
	err = l.svc.Unfollow(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.Unfollow(addr=%s): %s", addr, err))
	return
}

func logLevel(err error) (lvl slog.Level) {
	switch err {
	case nil:
		lvl = slog.LevelDebug
	default:
		lvl = slog.LevelError
	}
	return
}
