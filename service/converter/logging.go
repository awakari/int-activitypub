package converter

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
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

func (l logging) Convert(ctx context.Context, actor vocab.Actor, activity vocab.Activity) (evt *pb.CloudEvent, err error) {
	evt, err = l.svc.Convert(ctx, actor, activity)
	l.log.Log(ctx, logLevel(err), fmt.Sprintf("converter.Convert(actor=%s, activity=%s): %s, %s", actor.ID, activity.ID, evt.Id, err))
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
