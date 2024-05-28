package main

import (
	"time"

	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"github.com/heartbytenet/go-lerpc/pkg/server"
)

func main() {
	s := server.NewServer()

	s.AddHandler(server.NewHandlerWith(
		"base",
		"ping",
		func(request proto.Request) (result proto.Result) {
			return proto.NewResult().
				WithCode(proto.ResultCodeSuccess).
				SetData("ts", time.Now().UnixMilli())
		},
	))

	if err := s.Run(); err != nil {
		panic(err)
	}
}
