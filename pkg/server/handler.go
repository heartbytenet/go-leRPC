package server

import (
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

type Handler interface {
	Match(namespace string, method string) bool
	Auth(ctx *RequestContext, token string) bool
	Execute(ctx *RequestContext, request proto.Request) (result proto.Result)
}

type HandlerBase struct {
	fnMatch   HandlerMatchFunction
	fnAuth    HandlerAuthFunction
	fnExecute HandlerExecuteFunction
}

type HandlerMatchFunction func(namespace string, method string) bool
type HandlerAuthFunction func(ctx *RequestContext, token string) bool
type HandlerExecuteFunction func(ctx *RequestContext, request proto.Request) (result proto.Result)

func NewHandler(
	fnMatch HandlerMatchFunction,
	fnAuth HandlerAuthFunction,
	fnExecute HandlerExecuteFunction,
) Handler {
	return &HandlerBase{
		fnMatch,
		fnAuth,
		fnExecute,
	}
}

func NewHandlerWith(namespace string, method string, fnAuth HandlerAuthFunction, fnExecute HandlerExecuteFunction) Handler {
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
		fnAuth,
		fnExecute)
}

func NewHandlerWithToken(namespace string, method string, token string, fnExecute HandlerExecuteFunction) Handler {
	return NewHandlerWith(
		namespace,
		method,
		func(ctx *RequestContext, t string) bool {
			return t == token
		},
		fnExecute)
}

func AuthNone() func(ctx *RequestContext, token string) bool {
	return func(ctx *RequestContext, token string) bool {
		return true
	}
}

func (handler *HandlerBase) Match(namespace string, method string) bool {
	if handler.fnMatch == nil {
		return false
	}

	return handler.fnMatch(namespace, method)
}

func (handler *HandlerBase) Auth(ctx *RequestContext, token string) bool {
	if handler.fnAuth == nil {
		return true
	}

	return handler.fnAuth(ctx, token)
}

func (handler *HandlerBase) Execute(ctx *RequestContext, request proto.Request) (result proto.Result) {
	return handler.fnExecute(ctx, request)
}
