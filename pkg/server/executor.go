package server

import (
	"log"
	"time"

	"github.com/heartbytenet/bblib/collections/generic"
	"github.com/heartbytenet/bblib/containers/optionals"
	"github.com/heartbytenet/bblib/containers/sync"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

var (
	ErrorHandlerNotFound = "handler not found"
)

type Executor struct {
	queue      *sync.Mutex[[]generic.Pair[proto.Request, *proto.Promise[proto.Result]]]
	queueLimit int
	handlers   *sync.Mutex[[]Handler]
}

func NewExecutor(queueLimit int) (executor *Executor) {
	executor = &Executor{
		queue:      sync.NewMutex(make([]generic.Pair[proto.Request, *proto.Promise[proto.Result]], 0)),
		queueLimit: queueLimit,
		handlers:   sync.NewMutex(make([]Handler, 0)),
	}

	return executor
}

func (executor *Executor) AddHandler(handler Handler) {
	executor.handlers.Map(func(data []Handler) []Handler {
		return append(data, handler)
	})
}

func (executor *Executor) GetHandler(namespace string, method string) (result optionals.Optional[Handler]) {
	result = optionals.None[Handler]()

	executor.handlers.Apply(func(data []Handler) {
		for _, handler := range data {
			if handler.Match(namespace, method) {
				result = optionals.Some(handler)
				break
			}
		}
	})

	return
}

func (executor *Executor) Start(loop time.Duration) (err error) {
	go executor.Loop(loop)

	return
}

func (executor *Executor) Loop(duration time.Duration) {
	var (
		ticker *time.Ticker
		err    error
	)

	ticker = time.NewTicker(duration)

	for {
		<-ticker.C
		err = executor.ExecuteOne()
		if err != nil {
			log.Println("failed at executing request:", err)
			continue
		}
	}
}

func (executor *Executor) CreateQueueEntry(request proto.Request) generic.Pair[proto.Request, *proto.Promise[proto.Result]] {
	return generic.NewPair(
		request,
		proto.NewPromise[proto.Result](),
	)
}

func (executor *Executor) PushRequest(request proto.Request) (entry *proto.Promise[proto.Result], flag bool) {
	executor.queue.Map(func(data []generic.Pair[proto.Request, *proto.Promise[proto.Result]]) []generic.Pair[proto.Request, *proto.Promise[proto.Result]] {
		if len(data) >= executor.queueLimit {
			flag = false
			return data
		}

		value := executor.CreateQueueEntry(request)

		entry = value.B()
		flag = true
		return append(data, value)
	})

	return
}

func (executor *Executor) ExecuteOne() (err error) {
	entry := optionals.None[generic.Pair[proto.Request, *proto.Promise[proto.Result]]]()

	executor.queue.Map(func(data []generic.Pair[proto.Request, *proto.Promise[proto.Result]]) []generic.Pair[proto.Request, *proto.Promise[proto.Result]] {
		if len(data) < 1 {
			return data
		}

		entry = optionals.Some(data[0])
		return data[1:]
	})

	entry.IfPresent(func(value generic.Pair[proto.Request, *proto.Promise[proto.Result]]) {
		var (
			request proto.Request
			promise *proto.Promise[proto.Result]
			result  proto.Result
		)

		request = value.A()
		promise = value.B()

		result, err = executor.ExecuteRequest(request)
		if err != nil {
			return
		}

		promise.Complete(result)
	})
	if err != nil {
		return
	}

	return
}

func (executor *Executor) ExecuteRequest(request proto.Request) (result proto.Result, err error) {
	executor.GetHandler(request.GetNamespace(), request.GetMethod()).
		IfPresentElse(
			func(handler Handler) {
				result = handler.Execute(request).
					WithKey(request.GetKey())
			},
			func() {
				result = proto.NewResult().
					WithCode(proto.ResultCodeError).
					WithMessage(ErrorHandlerNotFound)
			},
		)
	if err != nil {
		return
	}

	return
}
