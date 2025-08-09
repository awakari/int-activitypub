package reader

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"log/slog"
)

type serviceLogging struct {
	svc Service
	log *slog.Logger
}

func NewLogging(svc Service, log *slog.Logger) Service {
	return serviceLogging{
		svc: svc,
		log: log,
	}
}

func (sl serviceLogging) Feed(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error) {
	last, err = sl.svc.Feed(ctx, interestId, limit)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.Feed(%s, %d): %d, err=%s", interestId, limit, len(last), err))
	return
}
