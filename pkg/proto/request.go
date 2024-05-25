package proto

import (
	"github.com/heartbytenet/bblib/containers/optionals"
)

type Request struct {
	Token     string         `json:"t,omitempty"`
	Key       string         `json:"k,omitempty"`
	Namespace string         `json:"n,omitempty"`
	Method    string         `json:"m,omitempty"`
	Params    map[string]any `json:"p,omitempty"`
}

func NewRequest() Request {
	return Request{}
}

func (request Request) WithToken(value string) Request {
	request.Token = value

	return request
}

func (request Request) WithKey(value string) Request {
	request.Key = value

	return request
}

func (request Request) WithNamespace(value string) Request {
	request.Namespace = value

	return request
}

func (request Request) WithMethod(value string) Request {
	request.Method = value

	return request
}

func (request Request) WithParams(value map[string]any) Request {
	request.Params = value

	return request
}

func (request Request) SetParam(key string, value any) Request {
	if request.Params == nil {
		request.Params = map[string]any{}
	}

	request.Params[key] = value

	return request
}

func (request Request) GetParam(key string) optionals.Optional[any] {
	value, flag := request.Params[key]
	if !flag {
		return optionals.None[any]()
	}

	return optionals.Some(value)
}
