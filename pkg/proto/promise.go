package proto

type Promise[T any] struct {
	callback chan T
}

func NewPromise[T any]() Promise[T] {
	return Promise[T]{
		callback: make(chan T, 1),
	}
}

func (promise Promise[T]) Complete(value T) {
	promise.callback <- value
}

func (promise Promise[T]) Await() T {
	return <-promise.callback
}
