package activitypub

import (
	"context"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"log/slog"
)

type logging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) ResolveActorLink(ctx context.Context, host, name string) (self vocab.IRI, err error) {
	self, err = l.svc.ResolveActorLink(ctx, host, name)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("http.ResolveActorLink(host=%s, name=%s): %s, %s", host, name, self, err))
	return
}

func (l logging) FetchActor(ctx context.Context, addr vocab.IRI) (actor vocab.Actor, err error) {
	actor, err = l.svc.FetchActor(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("http.FetchActor(addr=%s): {Id;%+v, Inbox:%+v}, %s", addr, actor.ID, actor.Inbox, err))
	return
}

func (l logging) SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI) (err error) {
	err = l.svc.SendActivity(ctx, a, inbox)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("http.SendActivity(a=%+v, inbox=%s): %s", a, inbox, err))
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
