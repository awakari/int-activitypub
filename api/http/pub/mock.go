package pub

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

type mock struct {
}

func NewMock() Service {
	return mock{}
}

func (m mock) Publish(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error) {
	return
}
