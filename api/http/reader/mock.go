package reader

import (
	"context"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
)

type mock struct {
}

func (m mock) Read(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error) {
	//TODO implement me
	panic("implement me")
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) CreateCallback(ctx context.Context, subId, url string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (m mock) GetCallback(ctx context.Context, subId, url string) (cb Callback, err error) {
	//TODO implement me
	panic("implement me")
}

func (m mock) DeleteCallback(ctx context.Context, subId, url string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (m mock) CountByInterest(ctx context.Context, interestId string) (count int64, err error) {
	count = 42
	return
}
