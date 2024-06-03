package mastodon

import "context"

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error) {
	return 42, nil
}

func (m mock) ConsumeLiveStreamPublic() (err error) {
	return
}
