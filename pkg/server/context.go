package server

import (
	"github.com/heartbytenet/bblib/collections/generic"
	"github.com/heartbytenet/bblib/containers/optionals"
	"github.com/heartbytenet/go-lerpc/pkg/client"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

type RequestContext struct {
	executor   *Executor
	clientMode client.ClientMode
	conn       optionals.Optional[chan generic.Pair[int, []byte]]
	request    proto.Request
}

func NewRequestContext(executor *Executor, clientMode client.ClientMode, outgoing chan generic.Pair[int, []byte], request proto.Request) *RequestContext {
	return &RequestContext{
		executor:   executor,
		clientMode: clientMode,
		conn:       optionals.FromNillable[chan generic.Pair[int, []byte]](outgoing),
		request:    request,
	}
}

func (ctx *RequestContext) GetExecutor() *Executor {
	return ctx.executor
}

func (ctx *RequestContext) GetClientMode() client.ClientMode {
	return ctx.clientMode
}

func (ctx *RequestContext) GetConn() optionals.Optional[chan generic.Pair[int, []byte]] {
	return ctx.conn
}

func (ctx *RequestContext) GetRequest() proto.Request {
	return ctx.request
}
