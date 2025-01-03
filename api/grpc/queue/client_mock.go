package queue

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type clientMock struct {
	rcvLimit int
	rcvDelay time.Duration
}

func newClientMock(rcvLimit int, rcvDelay time.Duration) ServiceClient {
	return clientMock{
		rcvLimit: rcvLimit,
		rcvDelay: rcvDelay,
	}
}

func (cm clientMock) SetQueue(ctx context.Context, in *SetQueueRequest, opts ...grpc.CallOption) (resp *emptypb.Empty, err error) {
	resp = &emptypb.Empty{}
	switch in.Name {
	case "fail":
		err = ErrInternal
	}
	return
}

func (cm clientMock) ReceiveMessages(ctx context.Context, opts ...grpc.CallOption) (Service_ReceiveMessagesClient, error) {
	return newRcvStreamMock(cm.rcvLimit, cm.rcvDelay), nil
}
