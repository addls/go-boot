package server

import (
	v1 "%s/api/network/v1"
	"%s/internal/service"

	"github.com/go-kratos/kratos/v2/transport/http"
)

func RegisterHTTPServer(httpSrv *http.Server) {
	v1.RegisterPingHTTPServer(httpSrv, service.NewPingService())
}
