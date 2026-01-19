package server

import (
	v1 "%s/api/network/v1"
	"%s/internal/service"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func RegisterGRPCServer(grpcSrv *grpc.Server) {
	v1.RegisterPingServer(grpcSrv.Server, service.NewPingService())
}
