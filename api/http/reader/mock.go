package reader

import "context"

type mock struct {
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
