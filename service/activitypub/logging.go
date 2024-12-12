package activitypub

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
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
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("activitypub.ResolveActorLink(host=%s, name=%s): %s, %s", host, name, self, err))
	return
}

func (l logging) FetchActor(ctx context.Context, addr vocab.IRI, pubKeyId string) (actor vocab.Actor, tags util.ObjectTags, err error) {
	actor, tags, err = l.svc.FetchActor(ctx, addr, pubKeyId)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("activitypub.FetchActor(addr=%s, pubKeyId=%s): {Id;%+v, Inbox:%+v, Tags:%+v}, %s", addr, pubKeyId, actor.ID, actor.Inbox, tags, err))
	return
}

func (l logging) SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI, pubKeyId string) (err error) {
	err = l.svc.SendActivity(ctx, a, inbox, pubKeyId)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("activitypub.SendActivity(a=%v, inbox=%s, pubKeyId=%s): %s", a, inbox, pubKeyId, err))
	return
}

func (l logging) IsOpenRegistration() (isOpen bool, err error) {
	isOpen, err = l.svc.IsOpenRegistration()
	l.log.Log(context.TODO(), util.LogLevel(err), fmt.Sprintf("activitypub.IsOpenRegistration(): %t, %s", isOpen, err))
	return
}

func (l logging) Usage() (u nodeinfo.Usage, err error) {
	u, err = l.svc.Usage()
	l.log.Log(context.TODO(), util.LogLevel(err), fmt.Sprintf("activitypub.Usage(): %+v, %s", u, err))
	return
}
