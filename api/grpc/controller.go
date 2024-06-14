package grpc

import (
	"context"
	"errors"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/storage"
	vocab "github.com/go-ap/activitypub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type controller struct {
	svc service.Service
}

func NewController(svc service.Service) ServiceServer {
	return controller{
		svc: svc,
	}
}

func (c controller) Create(ctx context.Context, req *CreateRequest) (resp *CreateResponse, err error) {
	resp = &CreateResponse{}
	resp.Url, err = c.svc.RequestFollow(ctx, req.Addr, req.GroupId, req.UserId, req.SubId, req.Term)
	err = encodeError(err)
	return
}

func (c controller) Read(ctx context.Context, req *ReadRequest) (resp *ReadResponse, err error) {
	resp = &ReadResponse{}
	src, err := c.svc.Read(ctx, vocab.IRI(req.Url))
	switch err {
	case nil:
		resp.Src = encodeSource(src)
	default:
		err = encodeError(err)
	}
	return
}

func (c controller) Delete(ctx context.Context, req *DeleteRequest) (resp *DeleteResponse, err error) {
	resp = &DeleteResponse{}
	err = c.svc.Unfollow(ctx, vocab.IRI(req.Url), req.GroupId, req.UserId)
	switch err {
	case nil:
	default:
		err = encodeError(err)
	}
	return
}

func (c controller) ListUrls(ctx context.Context, req *ListUrlsRequest) (resp *ListUrlsResponse, err error) {
	resp = &ListUrlsResponse{}
	var filter model.Filter
	reqFilter := req.Filter
	if reqFilter != nil {
		filter.Pattern = reqFilter.Pattern
		filter.GroupId = reqFilter.GroupId
		filter.UserId = reqFilter.UserId
		filter.SubId = reqFilter.SubId
	}
	var order model.Order
	switch req.Order {
	case Order_DESC:
		order = model.OrderDesc
	default:
		order = model.OrderAsc
	}
	page, err := c.svc.List(ctx, filter, req.Limit, req.Cursor, order)
	switch err {
	case nil:
		for _, addr := range page {
			resp.Page = append(resp.Page, addr)
		}
	default:
		err = encodeError(err)
	}
	return
}

func encodeSource(src model.Source) (dst *Source) {
	dst = &Source{
		ActorId:  src.ActorId,
		GroupId:  src.GroupId,
		UserId:   src.UserId,
		Type:     src.Type,
		Name:     src.Name,
		Summary:  src.Summary,
		Accepted: src.Accepted,
		SubId:    src.SubId,
		Term:     src.Term,
	}
	if !src.Created.IsZero() {
		dst.Created = timestamppb.New(src.Created)
	}
	if !src.Last.IsZero() {
		dst.Last = timestamppb.New(src.Last)
	}
	return
}

func encodeError(src error) (dst error) {
	switch {
	case src == nil:
	case errors.Is(src, storage.ErrConflict):
		dst = status.Error(codes.AlreadyExists, src.Error())
	case errors.Is(src, storage.ErrNotFound):
		dst = status.Error(codes.NotFound, src.Error())
	case errors.Is(src, storage.ErrInternal), errors.Is(src, activitypub.ErrActivitySend):
		dst = status.Error(codes.Internal, src.Error())
	case errors.Is(src, service.ErrInvalid), errors.Is(src, storage.ErrInternal):
		dst = status.Error(codes.InvalidArgument, src.Error())
	case errors.Is(src, service.ErrNoBot):
		dst = status.Error(codes.PermissionDenied, src.Error())
	case errors.Is(src, context.DeadlineExceeded):
		dst = status.Error(codes.DeadlineExceeded, src.Error())
	case errors.Is(src, context.Canceled):
		dst = status.Error(codes.Canceled, src.Error())
	default:
		dst = status.Error(codes.Unknown, src.Error())
	}
	return
}
