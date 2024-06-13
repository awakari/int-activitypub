package activitypub

import (
	"context"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"github.com/writeas/go-nodeinfo"
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
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("activitypub.ResolveActorLink(host=%s, name=%s): %s, %s", host, name, self, err))
	return
}

func (l logging) FetchActor(ctx context.Context, addr vocab.IRI) (actor vocab.Actor, err error) {
	actor, err = l.svc.FetchActor(ctx, addr)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("activitypub.FetchActor(addr=%s): {Id;%+v, Inbox:%+v}, %s", addr, actor.ID, actor.Inbox, err))
	return
}

func (l logging) SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI) (err error) {
	err = l.svc.SendActivity(ctx, a, inbox)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("activitypub.SendActivity(a=%+v, inbox=%s): %s", a, inbox, err))
	return
}

func (l logging) IsOpenRegistration() (isOpen bool, err error) {
	isOpen, err = l.svc.IsOpenRegistration()
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("activitypub.IsOpenRegistration(): %t, %s", isOpen, err))
	return
}

func (l logging) Usage() (u nodeinfo.Usage, err error) {
	u, err = l.svc.Usage()
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("activitypub.Usage(): %+v, %s", u, err))
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
