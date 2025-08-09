package subscriptions

import (
	"context"
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

func (m mock) Unsubscribe(ctx context.Context, _, _, _, url string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (m mock) CountByInterest(ctx context.Context, interestId, _, _ string) (count int64, err error) {
	count = 42
	return
}
