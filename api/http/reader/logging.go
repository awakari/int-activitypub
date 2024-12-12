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

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return serviceLogging{
		svc: svc,
		log: log,
	}
}

func (sl serviceLogging) CreateCallback(ctx context.Context, subId, url string) (err error) {
	err = sl.svc.CreateCallback(ctx, subId, url)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.CreateCallback(%s, %s): err=%s", subId, url, err))
	return
}

func (sl serviceLogging) GetCallback(ctx context.Context, subId, url string) (cb Callback, err error) {
	cb, err = sl.svc.GetCallback(ctx, subId, url)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.GetCallback(%s, %s): %+v, err=%s", subId, url, cb, err))
	return
}

func (sl serviceLogging) DeleteCallback(ctx context.Context, subId, url string) (err error) {
	err = sl.svc.DeleteCallback(ctx, subId, url)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.DeleteCallback(%s, %s): err=%s", subId, url, err))
	return
}

func (sl serviceLogging) CountByInterest(ctx context.Context, interestId string) (count int64, err error) {
	count, err = sl.svc.CountByInterest(ctx, interestId)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.CountByInterest(%s): %d, err=%s", interestId, count, err))
	return
}

func (sl serviceLogging) Read(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error) {
	last, err = sl.svc.Read(ctx, interestId, limit)
	ll := util.LogLevel(err)
	sl.log.Log(ctx, ll, fmt.Sprintf("reader.Read(%s, %d): %d, err=%s", interestId, limit, len(last), err))
	return
}
