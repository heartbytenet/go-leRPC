package proto

import (
	"github.com/heartbytenet/bblib/containers/optionals"
	"github.com/heartbytenet/bblib/reflection"
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

func GetParamValue[T any](request Request, key string) (value optionals.Optional[T]) {
	return optionals.FlatMap[any, T](
		request.GetParam(key),
		func(v any) optionals.Optional[T] {
			var (
				obj  T
				flag bool
			)

			obj, flag = v.(T)
			if !flag {
				return optionals.None[T]()
			}

			return optionals.Some(obj)
		})
}

func GetParamConvert[T any](request Request, key string) (value optionals.Optional[T]) {
	return optionals.FlatMap[any, T](
		request.GetParam(key),
		func(v any) optionals.Optional[T] {
			var (
				obj T
				err error
			)

			err = reflection.Convert(v, &obj)
			if err != nil {
				return optionals.None[T]()
			}

			return optionals.Some(obj)
		})
}

func (request Request) GetToken() string {
	return request.Token
}

func (request Request) GetKey() string {
	return request.Key
}

func (request Request) GetNamespace() string {
	return request.Namespace
}

func (request Request) GetMethod() string {
	return request.Method
}
