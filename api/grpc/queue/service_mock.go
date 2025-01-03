package queue

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"time"
)

type serviceMock struct {
	msgs []*pb.CloudEvent
}

func NewServiceMock(msgs []*pb.CloudEvent) Service {
	return serviceMock{
		msgs: msgs,
	}
}

func (sm serviceMock) SetConsumer(ctx context.Context, name, subj string) (err error) {
	switch name {
	case "fail":
		err = ErrInternal
	}
	return
}

func (sm serviceMock) ReceiveMessages(ctx context.Context, queue, subj string, batchSize uint32, consume util.ConsumeFunc[[]*pb.CloudEvent]) (err error) {
	switch {
	case queue == "fail":
		err = ErrInternal
	case queue == "queue_missing":
		err = ErrQueueMissing
	default:
		time.Sleep(1 * time.Microsecond)
		var msgs []*pb.CloudEvent
		for i := 0; i < int(batchSize); i++ {
			msg := pb.CloudEvent{
				Id:          fmt.Sprintf("msg%d", i),
				Source:      fmt.Sprintf("source%d", i),
				SpecVersion: fmt.Sprintf("specversion%d", i),
				Type:        fmt.Sprintf("type%d", i),
				Attributes:  map[string]*pb.CloudEventAttributeValue{},
				Data: &pb.CloudEvent_TextData{
					TextData: "yohoho",
				},
			}
			msgs = append(msgs, &msg)
		}
		_ = consume(msgs)
	}
	return
}
