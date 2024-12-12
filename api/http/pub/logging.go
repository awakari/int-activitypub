package pub

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
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

func (l logging) Publish(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error) {
	err = l.svc.Publish(ctx, evt, groupId, userId)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("pub.Publish(%s, %s, %s): err=%s", evt.Id, groupId, userId, err))
	return
}
