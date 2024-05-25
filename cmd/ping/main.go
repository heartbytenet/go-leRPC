package main

import (
	"github.com/heartbytenet/go-lerpc/pkg/server"
)

func main() {
	s := server.NewServer()

	if err := s.Run(); err != nil {
		panic(err)
	}
}
