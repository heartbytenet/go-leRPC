package server

import (
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

type Handler interface {
	Match(namespace string, method string) bool
	Execute(ctx *RequestContext, request proto.Request) (result proto.Result)
}

type HandlerBase struct {
	fnMatch   HandlerMatchFunction
	fnExecute HandlerExecuteFunction
}

type HandlerMatchFunction func(namespace string, method string) bool
type HandlerExecuteFunction func(ctx *RequestContext, request proto.Request) (result proto.Result)

func NewHandler(fnMatch HandlerMatchFunction, fnExecute HandlerExecuteFunction) Handler {
	return &HandlerBase{
		fnMatch,
		fnExecute,
	}
}

func NewHandlerWith(namespace string, method string, fnExecute HandlerExecuteFunction) Handler {
	return NewHandler(
		func(n string, m string) bool {
			if n != namespace {
				return false
			}

			if m != method {
				return false
			}

			return true
		},
		fnExecute)
}

func (handler *HandlerBase) Match(namespace string, method string) bool {
	return handler.fnMatch(namespace, method)
}

func (handler *HandlerBase) Execute(ctx *RequestContext, request proto.Request) (result proto.Result) {
	return handler.fnExecute(ctx, request)
}
