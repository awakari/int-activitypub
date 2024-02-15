package service

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/model"
	vocab "github.com/go-ap/activitypub"
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

func (l logging) RequestFollow(ctx context.Context, addr string) (url vocab.IRI, err error) {
	url, err = l.svc.RequestFollow(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.SendActivity(addr=%s): %s, %s", addr, url, err))
	return
}

func (l logging) HandleActivity(ctx context.Context, url vocab.IRI, activity vocab.Activity) (err error) {
	err = l.svc.HandleActivity(ctx, url, activity)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.HandleActivity(url=%s, activity.Type=%s): %s", url, activity.Type, err))
	return
}

func (l logging) Read(ctx context.Context, url vocab.IRI) (a model.Actor, err error) {
	a, err = l.svc.Read(ctx, url)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.Read(url=%s): %+v, %s", url, a, err))
	return
}

func (l logging) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = l.svc.List(ctx, filter, limit, cursor, order)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.List(filter=%+v, limit=%d, cursor=%s, order=%s): %d, %s", filter, limit, cursor, order, len(page), err))
	return
}

func (l logging) Unfollow(ctx context.Context, url vocab.IRI) (err error) {
	err = l.svc.Unfollow(ctx, url)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.Unfollow(url=%s): %s", url, err))
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
