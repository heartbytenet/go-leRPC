package net

import (
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"time"
)

type WebsocketPromise struct {
	result   *proto.ExecuteResult
	callback chan byte
	creation int64
}

func (p *WebsocketPromise) Init(result *proto.ExecuteResult) *WebsocketPromise {
	p.result = result
	p.callback = make(chan byte, 1)
	p.creation = time.Now().UnixMilli()
	return p
}
