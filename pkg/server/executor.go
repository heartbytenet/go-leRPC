package server

import (
	"log"
	"time"

	"github.com/heartbytenet/bblib/collections/generic"
	"github.com/heartbytenet/bblib/containers/optionals"
	"github.com/heartbytenet/bblib/containers/sync"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

type Executor struct {
	queue *sync.Mutex[[]generic.Pair[proto.Request, proto.Promise[proto.Result]]]
}

func NewExecutor() (executor *Executor) {
	executor = &Executor{
		queue: sync.NewMutex(make([]generic.Pair[proto.Request, proto.Promise[proto.Result]], 0)),
	}

	return executor
}

func (executor *Executor) Start() (err error) {
	go executor.Loop()

	return
}

func (executor *Executor) Loop() {
	var (
		ticker *time.Ticker
		err    error
	)

	ticker = time.NewTicker(time.Millisecond * 50)

	for {
		<-ticker.C

		err = executor.ExecuteOne()
		if err != nil {
			log.Println("failed at executing request:", err)
			continue
		}
	}
}

func (executor *Executor) CreateQueueEntry(request proto.Request) generic.Pair[proto.Request, proto.Promise[proto.Result]] {
	return generic.NewPair(
		request,
		proto.NewPromise[proto.Result](),
	)
}

func (executor *Executor) PushRequest(request proto.Request) proto.Promise[proto.Result] {
	entry := executor.CreateQueueEntry(request)

	executor.queue.Map(func(data []generic.Pair[proto.Request, proto.Promise[proto.Result]]) []generic.Pair[proto.Request, proto.Promise[proto.Result]] {
		return append(data, entry)
	})

	return entry.B()
}

func (executor *Executor) ExecuteOne() (err error) {
	entry := optionals.None[generic.Pair[proto.Request, proto.Promise[proto.Result]]]()

	executor.queue.Map(func(data []generic.Pair[proto.Request, proto.Promise[proto.Result]]) []generic.Pair[proto.Request, proto.Promise[proto.Result]] {
		if len(data) < 1 {
			return data
		}

		entry = optionals.Some(data[0])
		return data[1:]
	})

	entry.IfPresent(func(value generic.Pair[proto.Request, proto.Promise[proto.Result]]) {
		var (
			request proto.Request
			promise proto.Promise[proto.Result]
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
	result = proto.NewResult().
		WithCode(proto.ResultCodeSuccess).
		SetData("value", 37) // Totally random number

	// Todo: implement handlers

	return
}
