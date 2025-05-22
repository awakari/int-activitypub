package reader

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"log/slog"
	"time"
)

type serviceLogging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return serviceLogging{
		svc: svc,
		log: log,
	}
}

func (sl serviceLogging) Subscribe(ctx context.Context, interestId, groupId, userId, url string, interval time.Duration) (err error) {
	err = sl.svc.Subscribe(ctx, interestId, groupId, userId, url, interval)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.Subscribe(%s, %s, %s): err=%s", interestId, url, interval, err))
	return
}

func (sl serviceLogging) Subscription(ctx context.Context, interestId, groupId, userId, url string) (cb Subscription, err error) {
	cb, err = sl.svc.Subscription(ctx, interestId, groupId, userId, url)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.Subscription(%s, %s): %+v, err=%s", interestId, url, cb, err))
	return
}

func (sl serviceLogging) Unsubscribe(ctx context.Context, interestId, groupId, userId, url string) (err error) {
	err = sl.svc.Unsubscribe(ctx, interestId, groupId, userId, url)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.Unsubscribe(%s, %s): err=%s", interestId, url, err))
	return
}

func (sl serviceLogging) CountByInterest(ctx context.Context, interestId, groupId, userId string) (count int64, err error) {
	count, err = sl.svc.CountByInterest(ctx, interestId, groupId, userId)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.CountByInterest(%s): %d, err=%s", interestId, count, err))
	return
}

func (sl serviceLogging) Feed(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error) {
	last, err = sl.svc.Feed(ctx, interestId, limit)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.Feed(%s, %d): %d, err=%s", interestId, limit, len(last), err))
	return
}
