package queue

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

func NewLoggingMiddleware(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) SetConsumer(ctx context.Context, name, subj string) (err error) {
	err = l.svc.SetConsumer(ctx, name, subj)
	ll := util.LogLevel(err)
	l.log.Log(ctx, ll, fmt.Sprintf("queue.SetConsumer(name=%s, subj=%s): err=%s", name, subj, err))
	return
}

func (l logging) ReceiveMessages(ctx context.Context, queue, subj string, batchSize uint32, consume util.ConsumeFunc[[]*pb.CloudEvent]) (err error) {
	err = l.svc.ReceiveMessages(ctx, queue, subj, batchSize, consume)
	ll := util.LogLevel(err)
	l.log.Log(ctx, ll, fmt.Sprintf("queue.ReceiveMessages(queue=%s, subj=%s, batchSize=%d): err=%s", queue, subj, batchSize, err))
	return
}
