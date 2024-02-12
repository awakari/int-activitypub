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

func NewLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) ResolveActor(ctx context.Context, host, name string) (self vocab.IRI, err error) {
	self, err = l.svc.ResolveActor(ctx, host, name)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("activitypub.ResolveActor(host=%s, name=%s): %s, %s", host, name, self, err))
	return
}

func (l logging) ResolveInbox(ctx context.Context, addr vocab.IRI) (inbox vocab.IRI, err error) {
	inbox, err = l.svc.ResolveInbox(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("activitypub.ResolveInbox(addr=%s): %s, %s", addr, inbox, err))
	return
}

func (l logging) RequestFollow(ctx context.Context, host string, addr, inbox vocab.IRI) (err error) {
	err = l.svc.RequestFollow(ctx, host, addr, inbox)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("activitypub.ResolveActor(host=%s, addr=%s, inbox=%s): %s", host, addr, inbox, err))
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
