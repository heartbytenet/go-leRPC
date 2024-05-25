package server

import (
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

type Handler interface {
	Match(namespace string, method string) bool
	Execute(request proto.Request) (result proto.Result)
}

type HandlerBase struct {
	fnMatch   func(namespace string, method string) bool
	fnExecute func(request proto.Request) (result proto.Result)
}

func NewHandler(fnMatch func(namespace string, method string) bool, fnExecute func(request proto.Request) (result proto.Result)) Handler {
	return &HandlerBase{
		fnMatch,
		fnExecute,
	}
}

func (handler *HandlerBase) Match(namespace string, method string) bool {
	return handler.fnMatch(namespace, method)
}

func (handler *HandlerBase) Execute(request proto.Request) (result proto.Result) {
	return handler.fnExecute(request)
}
