package queue

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestService_SetQueue(t *testing.T) {
	svc := NewService(newClientMock(0, 0))
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	svc = NewLoggingMiddleware(svc, log)
	cases := map[string]error{
		"ok":   nil,
		"fail": ErrInternal,
	}
	for k, expectedErr := range cases {
		t.Run(k, func(t *testing.T) {
			err := svc.SetConsumer(context.TODO(), k, k)
			assert.ErrorIs(t, err, expectedErr)
		})
	}
}

func TestService_ReceiveMessages(t *testing.T) {
	cases := map[string]struct {
		countMin int
		countMax int
		delay    time.Duration
		err      error
	}{
		"ok": {
			countMin: 1000,
			countMax: 1000,
			delay:    0,
		},
		"timeout": {
			countMin: 10,
			countMax: 100,
			delay:    100 * time.Millisecond,
			err:      context.DeadlineExceeded,
		},
		"fail": {
			err: ErrInternal,
		},
		"missing": {
			err: ErrQueueMissing,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			svc := NewService(newClientMock(1000, c.delay))
			log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
			svc = NewLoggingMiddleware(svc, log)
			ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
			defer cancel()
			var msgs []*pb.CloudEvent
			consume := func(msgBatch []*pb.CloudEvent) (err error) {
				msgs = append(msgs, msgBatch...)
				return
			}
			err := svc.ReceiveMessages(ctx, k, k, 10, consume)
			assert.ErrorIs(t, err, c.err)
			assert.LessOrEqual(t, c.countMin, len(msgs))
			assert.GreaterOrEqual(t, c.countMax, len(msgs))
		})
	}
}
