package grpc

import (
	"context"
	"github.com/awakari/int-activitypub/service"
)

type controller struct {
	svc service.Service
}

func NewController(svc service.Service) ServiceServer {
	return controller{
		svc: svc,
	}
}

func (c controller) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c controller) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c controller) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (c controller) ListUrls(ctx context.Context, req *ListUrlsRequest) (*ListUrlsResponse, error) {
	//TODO implement me
	panic("implement me")
}
