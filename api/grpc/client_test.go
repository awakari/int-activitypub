package grpc

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log/slog"
	"os"
	"testing"
)

var port uint16 = 50051

var log = slog.Default()

func TestMain(m *testing.M) {
	svc := service.NewServiceMock()
	svc = service.NewLogging(svc, log)
	go func() {
		err := Serve(port, svc)
		if err != nil {
			log.Error("", err)
		}
	}()
	code := m.Run()
	os.Exit(code)
}

func TestServiceClient_Create(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req *CreateRequest
		err error
	}{
		"ok": {
			req: &CreateRequest{
				Addr: "user1@server1.social",
			},
		},
		"fail": {
			req: &CreateRequest{
				Addr: "fail",
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			req: &CreateRequest{
				Addr: "missing",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		"invalid": {
			req: &CreateRequest{
				Addr: "invalid",
			},
			err: status.Error(codes.InvalidArgument, "invalid argument"),
		},
		"conflict": {
			req: &CreateRequest{
				Addr: "conflict",
			},
			err: status.Error(codes.AlreadyExists, "already exists"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err := client.Create(context.TODO(), c.req)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestServiceClient_Read(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req  *ReadRequest
		resp *ReadResponse
		err  error
	}{
		"ok": {
			req: &ReadRequest{
				Addr: "user1@server1.social",
			},
			resp: &ReadResponse{
				Actor: &Actor{
					Addr:    "user1@server1.social",
					GroupId: "group1",
					UserId:  "user2",
					Type:    "Person",
					Name:    "John Doe",
					Summary: "yohoho",
				},
			},
		},
		"fail": {
			req: &ReadRequest{
				Addr: "fail",
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			req: &ReadRequest{
				Addr: "missing",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			resp, err := client.Read(context.TODO(), c.req)
			if c.resp != nil {
				assert.EqualExportedValues(t, *c.resp, *resp)
			}
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestServiceClient_ListUrls(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req  *ListUrlsRequest
		resp *ListUrlsResponse
		err  error
	}{
		"ok": {
			req: &ListUrlsRequest{
				Cursor: "user1@server1.social",
			},
			resp: &ListUrlsResponse{
				Page: []string{
					"user1@server1.social",
					"user2@server2.social",
				},
			},
		},
		"fail": {
			req: &ListUrlsRequest{
				Cursor: "fail",
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			resp, err := client.ListUrls(context.TODO(), c.req)
			if c.resp != nil {
				assert.EqualExportedValues(t, *c.resp, *resp)
			}
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestServiceClient_Delete(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req *DeleteRequest
		err error
	}{
		"ok": {
			req: &DeleteRequest{
				Addr: "user1@server1.social",
			},
		},
		"fail": {
			req: &DeleteRequest{
				Addr: "fail",
			},
			err: status.Error(codes.Internal, "internal failure"),
		},
		"missing": {
			req: &DeleteRequest{
				Addr: "missing",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		"invalid": {
			req: &DeleteRequest{
				Addr: "invalid",
			},
			err: status.Error(codes.InvalidArgument, "invalid argument"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err := client.Delete(context.TODO(), c.req)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
