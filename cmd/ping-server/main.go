package main

import (
	"time"

	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"github.com/heartbytenet/go-lerpc/pkg/server"
)

func main() {
	s := server.NewServer()

	s.AddHandler(server.NewHandlerWithToken(
		"base",
		"ping",
		"secret_token",
		func(_ *server.RequestContext, request proto.Request) (result proto.Result) {
			return proto.NewResult().
				WithCode(proto.ResultCodeSuccess).
				SetData("ts", time.Now().UnixMilli()).
				SetData("value", proto.GetParamConvert[float64](request, "value").GetDefault(0)+1)
		},
	))

	if err := s.Run(); err != nil {
		panic(err)
	}
}
