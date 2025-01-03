package queue

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type Service interface {
	SetConsumer(ctx context.Context, name, subj string) (err error)
	ReceiveMessages(ctx context.Context, queue, subj string, batchSize uint32, consume util.ConsumeFunc[[]*pb.CloudEvent]) (err error)
}

type service struct {
	client ServiceClient
}

var ErrInternal = errors.New("queue: internal failure")

var ErrQueueMissing = errors.New("missing queue")

func NewService(client ServiceClient) Service {
	return service{
		client: client,
	}
}

func (svc service) SetConsumer(ctx context.Context, name, subj string) (err error) {
	req := SetQueueRequest{
		Name: name,
		Subj: subj,
	}
	_, err = svc.client.SetQueue(ctx, &req)
	err = decodeError(ctx, err)
	return
}

func (svc service) ReceiveMessages(ctx context.Context, queue, subj string, batchSize uint32, consume util.ConsumeFunc[[]*pb.CloudEvent]) (err error) {
	var stream Service_ReceiveMessagesClient
	stream, err = svc.client.ReceiveMessages(ctx)
	if err != nil {
		err = decodeError(ctx, err)
	}
	var req *ReceiveMessagesRequest
	if err == nil {
		req = &ReceiveMessagesRequest{
			Command: &ReceiveMessagesRequest_Start{
				Start: &ReceiveMessagesCommandStart{
					Queue:     queue,
					BatchSize: batchSize,
					Subj:      subj,
				},
			},
		}
		err = stream.Send(req)
	}
	if err == nil {
		var resp *ReceiveMessagesResponse
		for {
			resp, err = stream.Recv()
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				err = decodeError(ctx, err)
				break
			}
			if resp != nil {
				err = errors.Join(err, consume(resp.Msgs))
				req = &ReceiveMessagesRequest{
					Command: &ReceiveMessagesRequest_Ack{
						Ack: &ReceiveMessagesCommandAck{
							Count: uint32(len(resp.Msgs)),
						},
					},
				}
				err = errors.Join(err, stream.Send(req))
			}
			if err == nil {
				select {
				case <-ctx.Done():
					err = ctx.Err()
				default:
					continue
				}
			}
			if err != nil {
				break
			}
		}
	}
	return
}

func decodeError(ctx context.Context, src error) (dst error) {
	switch {
	case src == io.EOF:
		dst = src // return as it is
	case status.Code(src) == codes.OK:
		dst = nil
	case status.Code(src) == codes.NotFound:
		dst = ErrQueueMissing
	default:
		dst = fmt.Errorf("%w: %s", ErrInternal, src)
	}
	return
}
