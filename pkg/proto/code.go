package proto

type ResultCode uint8

const (
	ResultCodeNone ResultCode = iota
	ResultCodeSuccess
	ResultCodeWarning
	ResultCodeError
)
