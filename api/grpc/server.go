package grpc

import (
	"fmt"
	"github.com/awakari/int-activitypub/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
)

func Serve(port uint16, svc service.Service) (err error) {
	srv := grpc.NewServer()
	c := NewController(svc)
	RegisterServiceServer(srv, c)
	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		err = srv.Serve(conn)
	}
	return
}
