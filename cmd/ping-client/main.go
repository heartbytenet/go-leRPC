package main

import (
	"github.com/heartbytenet/go-lerpc/pkg/client"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"github.com/schollz/progressbar/v3"
)

func main() {
	c := client.NewClient(client.ClientModeHttp, "localhost:3000", "secret_token")

	t := int64(1000 * 1000)
	bar := progressbar.Default(t)

	for i := int64(0); i < t; i++ {
		promise, err := c.Execute(proto.NewRequest().
			WithToken("secret_token").
			WithNamespace("base").
			WithMethod("ping").
			SetParam("value", 1337))
		if err != nil {
			panic(err)
		}

		value := promise.AwaitUnwrap()
		_ = value

		_ = bar.Add(1)
	}
}
