package writer

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

type mock struct {
}

func NewMock() Service {
	return mock{}
}

func (m mock) Close() error {
	return nil
}

func (m mock) Write(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error) {
	switch userId {
	case "fail":
		err = ErrWrite
	}
	return
}
