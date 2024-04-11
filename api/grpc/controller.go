package grpc

import (
	"context"
	"errors"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/mastodon"
	"github.com/awakari/int-activitypub/storage"
	vocab "github.com/go-ap/activitypub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controller struct {
	svc    service.Service
	search mastodon.Service
}

func NewController(svc service.Service, search mastodon.Service) ServiceServer {
	return controller{
		svc:    svc,
		search: search,
	}
}

func (c controller) Create(ctx context.Context, req *CreateRequest) (resp *CreateResponse, err error) {
	resp = &CreateResponse{}
	resp.Url, err = c.svc.RequestFollow(ctx, req.Addr, req.GroupId, req.UserId)
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

func (c controller) SearchAndAdd(ctx context.Context, req *SearchAndAddRequest) (resp *SearchAndAddResponse, err error) {
	resp = &SearchAndAddResponse{}
	resp.N, err = c.search.SearchAndAdd(ctx, req.SubId, req.Q, req.Limit)
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
	case errors.Is(src, context.DeadlineExceeded):
		dst = status.Error(codes.DeadlineExceeded, src.Error())
	case errors.Is(src, context.Canceled):
		dst = status.Error(codes.Canceled, src.Error())
	default:
		dst = status.Error(codes.Unknown, src.Error())
	}
	return
}
