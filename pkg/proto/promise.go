package proto

import (
	"context"
)

type Promise[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	value  T
	err    error
}

func NewPromise[T any]() *Promise[T] {
	ctx, cancel := context.WithCancel(context.Background())

	return &Promise[T]{
		ctx:    ctx,
		cancel: cancel,
		err:    nil,
	}
}

func (promise *Promise[T]) Complete(value T) {
	promise.value = value
	promise.cancel()
}

func (promise *Promise[T]) Failed(err error) {
	promise.err = err
	promise.cancel()
}

func (promise *Promise[T]) Await() (T, error) {
	select {
	case <-promise.ctx.Done():
		return promise.value, promise.err
	}
}

func (promise *Promise[T]) AwaitUnwrap() T {
	value, err := promise.Await()
	if err != nil {
		panic(err)
	}

	return value
}
