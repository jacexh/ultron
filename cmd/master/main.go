package main

import (
	"time"

	"github.com/wosai/ultron/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func main() {
	runner := ultron.NewMasterRunner()
	runner.Launch(
		grpc.KeepaliveEnforcementPolicy(
			keepalive.EnforcementPolicy{
				MinTime:             5 * time.Second,
				PermitWithoutStream: true,
			}),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Timeout: 2 * time.Second,
			}),
	)
}
