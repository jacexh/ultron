package main

import (
	"net"

	"github.com/wosai/ultron/v2"
	"github.com/wosai/ultron/v2/pkg/genproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":2021")
	if err != nil {
		panic(err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	master := ultron.NewUltronServer()
	genproto.RegisterUltronServiceServer(grpcServer, master)

	if err := grpcServer.Serve(listener); err != nil {
		ultron.Logger.Fatal("ultron server was shutdown", zap.Error(err))
	}
}
