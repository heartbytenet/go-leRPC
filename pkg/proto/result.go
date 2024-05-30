package proto

import (
	"fmt"
	"github.com/heartbytenet/bblib/containers/optionals"
)

type Result struct {
	Key     string         `json:"k,omitempty"`
	Code    ResultCode     `json:"c"`
	Data    map[string]any `json:"d"`
	Message string         `json:"m,omitempty"`
}

func NewResult() Result {
	return Result{}
}

func (result Result) WithKey(value string) Result {
	result.Key = value

	return result
}

func (result Result) WithCode(value ResultCode) Result {
	result.Code = value

	return result
}

func (result Result) WithData(value map[string]any) Result {
	result.Data = value

	return result
}

func (result Result) WithMessage(value string) Result {
	result.Message = value

	return result
}

func (result Result) SetData(key string, value any) Result {
	if result.Data == nil {
		result.Data = map[string]any{}
	}

	result.Data[key] = value

	return result
}

func (result Result) GetKey() string {
	return result.Key
}

func (result Result) GetCode() ResultCode {
	return result.Code
}

func (result Result) GetData(key string) optionals.Optional[any] {
	value, flag := result.Data[key]
	if !flag {
		return optionals.None[any]()
	}

	return optionals.Some(value)
}

func (result Result) GetDataAll() map[string]any {
	return result.Data
}

func (result Result) GetMessage() string {
	return result.Message
}

func (result Result) Check() (err error) {
	switch result.GetCode() {
	case ResultCodeError:
		return fmt.Errorf("request failed: %s", result.GetMessage())

	default:
		return nil
	}
}

func (result Result) Unwrap() {
	var (
		err error
	)

	err = result.Check()
	if err != nil {
		panic(err)
	}
}
