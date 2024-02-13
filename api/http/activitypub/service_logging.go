package activitypub

import (
	"context"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"log/slog"
)

type svcLogging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return svcLogging{
		svc: svc,
		log: log,
	}
}

func (l svcLogging) ResolveActorLink(ctx context.Context, host, name string) (self vocab.IRI, err error) {
	self, err = l.svc.ResolveActorLink(ctx, host, name)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("http.ResolveActorLink(host=%s, name=%s): %s, %s", host, name, self, err))
	return
}

func (l svcLogging) FetchActor(ctx context.Context, addr vocab.IRI) (actor vocab.Actor, err error) {
	actor, err = l.svc.FetchActor(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("http.FetchActor(addr=%s): %+v, %s", addr, actor, err))
	return
}

func (l svcLogging) RequestFollow(ctx context.Context, host string, addr, inbox vocab.IRI) (err error) {
	err = l.svc.RequestFollow(ctx, host, addr, inbox)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("http.RequestFollow(host=%s, addr=%s, inbox=%s): %s", host, addr, inbox, err))
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
