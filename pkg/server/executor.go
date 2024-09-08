package server

import (
	"log"
	"log/slog"
	"time"

	"github.com/heartbytenet/go-lerpc/pkg/client"

	"github.com/heartbytenet/bblib/collections/generic"
	"github.com/heartbytenet/bblib/containers/optionals"
	"github.com/heartbytenet/bblib/containers/sync"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

var (
	ErrorHandlerNotFound = "handler not found"
	ErrorAuthFailed      = "handler auth failed"
)

type Executor struct {
	queue      *sync.Locked[[]generic.Pair[*RequestContext, *proto.Promise[proto.Result]]]
	queueLimit int
	handlers   *sync.Locked[[]Handler]

	downloadHandlers *sync.Locked[[]DownloadHandler]
}

func NewExecutor(queueLimit int) (executor *Executor) {
	executor = &Executor{
		queue:            sync.NewLocked(make([]generic.Pair[*RequestContext, *proto.Promise[proto.Result]], 0)),
		queueLimit:       queueLimit,
		handlers:         sync.NewLocked(make([]Handler, 0)),
		downloadHandlers: sync.NewLocked(make([]DownloadHandler, 0)),
	}

	return executor
}

func (executor *Executor) AddHandler(handler Handler) {
	executor.handlers.Map(func(data []Handler) []Handler {
		return append(data, handler)
	})
}

func (executor *Executor) AddDownloadHandler(handler DownloadHandler) {
	executor.downloadHandlers.Map(func(data []DownloadHandler) []DownloadHandler {
		return append(data, handler)
	})
}

func (executor *Executor) GetHandler(namespace string, method string) (result optionals.Optional[Handler]) {
	result = optionals.None[Handler]()

	executor.handlers.Apply(func(data []Handler) {
		for _, handler := range data {
			if !handler.Match(namespace, method) {
				continue
			}

			result = optionals.Some(handler)
			break
		}
	})

	return
}

func (executor *Executor) GetDownloadHandlerAlive(key string) (result optionals.Optional[DownloadHandler]) {
	result = optionals.None[DownloadHandler]()

	executor.downloadHandlers.Apply(func(data []DownloadHandler) {
		for _, handler := range data {
			if !handler.Match(key) {
				continue
			}

			if !handler.GetIsAlive() {
				continue
			}

			result = optionals.Some(handler)
			break
		}
	})

	return
}

func (executor *Executor) Start(loop time.Duration) (err error) {
	go executor.LoopExecute(loop)
	go executor.LoopClearHandlers()

	return
}

func (executor *Executor) LoopExecute(duration time.Duration) {
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

func (executor *Executor) LoopClearHandlers() {
	var (
		ticker *time.Ticker
	)

	ticker = time.NewTicker(time.Minute)

	for {
		<-ticker.C

		slog.Info("clearing handlers")
		executor.downloadHandlers.Map(func(curr []DownloadHandler) (next []DownloadHandler) {
			next = make([]DownloadHandler, 0)

			for _, handler := range curr {
				if !handler.GetIsAlive() {
					continue
				}

				next = append(next, handler)
			}

			return
		})
	}
}

func (executor *Executor) CreateQueueEntry(
	clientMode client.ClientMode,
	outgoing chan generic.Pair[int, []byte],
	request proto.Request,
) generic.Pair[*RequestContext, *proto.Promise[proto.Result]] {
	return generic.NewPair(
		NewRequestContext(executor, clientMode, outgoing, request),
		proto.NewPromise[proto.Result](),
	)
}

func (executor *Executor) PushRequest(clientMode client.ClientMode, outgoing chan generic.Pair[int, []byte], request proto.Request) (entry *proto.Promise[proto.Result], flag bool) {
	executor.queue.Map(func(data []generic.Pair[*RequestContext, *proto.Promise[proto.Result]]) []generic.Pair[*RequestContext, *proto.Promise[proto.Result]] {
		if len(data) >= executor.queueLimit {
			flag = false
			return data
		}

		value := executor.CreateQueueEntry(clientMode, outgoing, request)

		entry = value.B()
		flag = true
		return append(data, value)
	})

	return
}

func (executor *Executor) ExecuteOne() (err error) {
	entry := optionals.None[generic.Pair[*RequestContext, *proto.Promise[proto.Result]]]()

	executor.queue.Map(func(data []generic.Pair[*RequestContext, *proto.Promise[proto.Result]]) []generic.Pair[*RequestContext, *proto.Promise[proto.Result]] {
		if len(data) < 1 {
			return data
		}

		entry = optionals.Some(data[0])
		return data[1:]
	})

	entry.IfPresent(func(value generic.Pair[*RequestContext, *proto.Promise[proto.Result]]) {
		var (
			ctx     *RequestContext
			promise *proto.Promise[proto.Result]
			result  proto.Result
		)

		ctx, promise = value.A(), value.B()

		result, err = executor.ExecuteRequest(ctx)
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

func (executor *Executor) ExecuteRequest(ctx *RequestContext) (result proto.Result, err error) {
	request := ctx.GetRequest()

	executor.GetHandler(request.GetNamespace(), request.GetMethod()).
		IfPresentElse(
			func(handler Handler) {
				if !handler.Auth(ctx, request.Token) {
					result = proto.NewResult().
						WithCode(proto.ResultCodeError).
						WithMessage(ErrorAuthFailed)
					return
				}

				result = handler.Execute(ctx, request).WithKey(request.GetKey())
			},
			func() {
				result = proto.NewResult().
					WithCode(proto.ResultCodeError).
					WithMessage(ErrorHandlerNotFound)
			},
		)

	return
}
