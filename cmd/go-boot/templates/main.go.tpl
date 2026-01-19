package main

import (
	"%s/internal/server"

	"github.com/addls/go-boot/bootstrap"
)

func main() {
	// 使用 WithHTTPRegister 注册 HTTP 路由（在服务器创建后调用）
	// 使用 WithGRPCRegister 注册 gRPC 服务（在服务器创建后调用）
	bootstrap.Run("%s",
		bootstrap.WithHTTPRegister(server.RegisterHTTPServer),
		bootstrap.WithGRPCRegister(server.RegisterGRPCServer),
	)
}
