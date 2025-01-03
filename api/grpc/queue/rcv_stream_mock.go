package queue

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"time"
)

type rcvStreamMock struct {
	queue string
	limit int
	count int
	delay time.Duration
}

func newRcvStreamMock(limit int, delay time.Duration) Service_ReceiveMessagesClient {
	return &rcvStreamMock{
		limit: limit,
		delay: delay,
	}
}

func (rsm *rcvStreamMock) Recv() (resp *ReceiveMessagesResponse, err error) {
	resp = &ReceiveMessagesResponse{}
	time.Sleep(rsm.delay)
	switch {
	case rsm.queue == "fail":
		err = status.Errorf(codes.Internal, "internal failure")
	case rsm.queue == "missing":
		err = status.Error(codes.NotFound, "queue missing")
	case rsm.count >= rsm.limit:
		err = io.EOF
	default:
		for i := 0; i < 3 && rsm.count < rsm.limit; i++ {
			msg := pb.CloudEvent{
				Id: fmt.Sprintf("msg%d", rsm.count),
			}
			resp.Msgs = append(resp.Msgs, &msg)
			rsm.count++
		}
	}
	return
}

func (rsm *rcvStreamMock) Send(req *ReceiveMessagesRequest) error {
	start := req.GetStart()
	switch {
	case start != nil:
		rsm.queue = start.Queue
	}
	return nil
}

func (rsm *rcvStreamMock) Header() (metadata.MD, error) {
	//TODO implement me
	panic("implement me")
}

func (rsm *rcvStreamMock) Trailer() metadata.MD {
	//TODO implement me
	panic("implement me")
}

func (rsm *rcvStreamMock) CloseSend() error {
	//TODO implement me
	panic("implement me")
}

func (rsm *rcvStreamMock) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (rsm *rcvStreamMock) SendMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (rsm *rcvStreamMock) RecvMsg(m interface{}) error {
	//TODO implement me
	panic("implement me")
}
