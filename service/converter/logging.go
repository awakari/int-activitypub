package converter

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	vocab "github.com/go-ap/activitypub"
	"log/slog"
	"time"
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

func (l logging) ConvertActivityToEvent(ctx context.Context, actor vocab.Actor, activity vocab.Activity, tags util.ActivityTags) (evt *pb.CloudEvent, err error) {
	evt, err = l.svc.ConvertActivityToEvent(ctx, actor, activity, tags)
	switch evt {
	case nil:
		l.log.Log(ctx, logLevel(err), fmt.Sprintf("converter.ConvertActivityToEvent(actor=%s, activity=%s, tags=%d): <nil>, %s", actor.ID, activity.ID, len(tags.Tag), err))
	default:
		l.log.Log(ctx, logLevel(err), fmt.Sprintf("converter.ConvertActivityToEvent(actor=%s, activity=%s, tags=%d): %s, %s", actor.ID, activity.ID, len(tags.Tag), evt.Id, err))
	}
	return
}

func (l logging) ConvertEventToActivity(ctx context.Context, evt *pb.CloudEvent, interestId string, follower *vocab.Actor, t *time.Time) (a vocab.Activity, err error) {
	a, err = l.svc.ConvertEventToActivity(ctx, evt, interestId, follower, t)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("converter.ConvertEventToActivity(evtId=%s, interestId=%s, follower=%v): err=%s", evt.Id, interestId, follower, err))
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
