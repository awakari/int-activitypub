package subscriptions

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
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
	sl.log.Log(ctx, ll, fmt.Sprintf("subscriptions.Subscribe(%s, %s, %s): err=%s", interestId, url, interval, err))
	return
}

func (sl serviceLogging) Unsubscribe(ctx context.Context, interestId, groupId, userId, url string) (err error) {
	err = sl.svc.Unsubscribe(ctx, interestId, groupId, userId, url)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("subscriptions.Unsubscribe(%s, %s): err=%s", interestId, url, err))
	return
}

func (sl serviceLogging) CountByInterest(ctx context.Context, interestId, groupId, userId string) (count int64, err error) {
	count, err = sl.svc.CountByInterest(ctx, interestId, groupId, userId)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("subscriptions.CountByInterest(%s): %d, err=%s", interestId, count, err))
	return
}
