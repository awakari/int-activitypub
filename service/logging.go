package service

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/util"
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

func (l logging) RequestFollow(ctx context.Context, addr, groupId, userId, subId, term string) (url string, err error) {
	url, err = l.svc.RequestFollow(ctx, addr, groupId, userId, subId, term)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.RequestFollow(addr=%s, groupId=%s, userId=%s, subId=%s, term=%s): %s, %s", addr, groupId, userId, subId, term, url, err))
	return
}

func (l logging) HandleActivity(ctx context.Context, actorIdLocal, pubKeyId string, actor vocab.Actor, actorTags util.ObjectTags, activity vocab.Activity, activityTags util.ActivityTags) (post func(), err error) {
	post, err = l.svc.HandleActivity(ctx, actorIdLocal, pubKeyId, actor, actorTags, activity, activityTags)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.HandleActivity(actorIdLocal=%s, actor.Id=%s, actor.Tags=%d, activity.Type=%s, activity.Tags=%d): err=%s", actorIdLocal, actor.ID, len(actorTags.Tag), activity.Type, len(activityTags.Tag), err))
	return
}

func (l logging) Read(ctx context.Context, url vocab.IRI) (a model.Source, err error) {
	a, err = l.svc.Read(ctx, url)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.Read(url=%s): %+v, %s", url, a, err))
	return
}

func (l logging) List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = l.svc.List(ctx, filter, limit, cursor, order)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.List(filter=%+v, limit=%d, cursor=%s, order=%s): %d, %s", filter, limit, cursor, order, len(page), err))
	return
}

func (l logging) Unfollow(ctx context.Context, url vocab.IRI, groupId, userId string) (err error) {
	err = l.svc.Unfollow(ctx, url, groupId, userId)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("service.Unfollow(url=%s, groupId=%s, userId=%s): %s", url, groupId, userId, err))
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
