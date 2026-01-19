package main

import (
	"%s/internal/server"
	"github.com/addls/go-boot/bootstrap"
)

func main() {
	bootstrap.Run("%s",
		bootstrap.WithHTTPRegister(server.RegisterHTTPServer),
		bootstrap.WithGRPCRegister(server.RegisterGRPCServer),
	)
}
