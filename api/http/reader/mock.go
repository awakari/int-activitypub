package reader

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"time"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) Subscribe(ctx context.Context, _, _, _, url string, interval time.Duration) (err error) {
	//TODO implement me
	panic("implement me")
}

func (m mock) Subscription(ctx context.Context, _, _, _, url string) (cb Subscription, err error) {
	//TODO implement me
	panic("implement me")
}

func (m mock) Unsubscribe(ctx context.Context, _, _, _, url string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (m mock) CountByInterest(ctx context.Context, interestId, _, _ string) (count int64, err error) {
	count = 42
	return
}

func (m mock) Feed(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error) {
	//TODO implement me
	panic("implement me")
}
