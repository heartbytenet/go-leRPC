package main

import (
	"fmt"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"github.com/heartbytenet/go-lerpc/pkg/server"
	"math/rand"
	"time"
)

func main() {
	s := server.NewServer()

	s.AddHandler(server.NewHandlerWith(
		"download",
		"file.prepare",
		server.AuthNone(),
		func(ctx *server.RequestContext, request proto.Request) (result proto.Result) {
			key := fmt.Sprintf("%x", rand.Int())

			ctx.GetExecutor().AddDownloadHandler(server.NewDownloadHandlerFile(
				key,
				time.Second*5,
				5,
				"image/png",
				"./download.png"))

			return proto.NewResult().
				WithCode(proto.ResultCodeSuccess).
				SetData("key", key)
		}))

	if err := s.Run(); err != nil {
		panic(err)
	}
}
