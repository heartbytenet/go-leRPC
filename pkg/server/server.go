package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/heartbytenet/bblib/collections/generic"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/heartbytenet/bblib/debug"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
)

var (
	FallbackErrorMessage = "error"
)

func init() {
	if debug.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

type Server struct {
	settings Settings

	executor *Executor

	engine   *gin.Engine
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	settings := NewSettingsDefault()
	executor := NewExecutor(settings.ExecutorLimit)

	return &Server{
		settings: settings,
		executor: executor,

		engine:   gin.Default(),
		upgrader: websocket.Upgrader{},
	}
}

func (server *Server) AddHandler(handler Handler) {
	server.executor.AddHandler(handler)
}

func (server *Server) Addr() string {
	return fmt.Sprintf(":%d", server.settings.Port)
}

func (server *Server) Run() (err error) {
	err = server.executor.Start(time.Millisecond * server.settings.ExecutorDelay)
	if err != nil {
		log.Fatalln("failed at starting executor:", err)
	}

	server.engine.GET("/connect", server.HandleConnect)
	server.engine.POST("/execute", server.HandleExecute)

	err = server.engine.Run(server.Addr())
	if err != nil {
		log.Fatalln("failed at running server:", err)
	}

	return
}

func (server *Server) ErrorResult(err string) (result proto.Result) {
	result = proto.NewResult().
		WithCode(proto.ResultCodeError)

	if debug.DEBUG {
		return result.WithMessage(err)
	} else {
		return result.WithMessage(FallbackErrorMessage)
	}
}

func (server *Server) HandleExecute(ctx *gin.Context) {
	var (
		request proto.Request
		promise *proto.Promise[proto.Result]
		result  proto.Result
		flag    bool
		err     error
	)

	err = ctx.BindJSON(&request)
	if err != nil {
		ctx.JSON(500, server.ErrorResult(err.Error()))
		return
	}

	promise, flag = server.executor.PushRequest(request)
	if !flag {
		ctx.JSON(500, server.ErrorResult("executor queue is full"))
		return
	}

	result, err = promise.Await()
	if err != nil {
		ctx.JSON(500, server.ErrorResult(err.Error()))
		return
	}

	ctx.JSON(200, result)
	return
}

func (server *Server) HandleConnect(ctx *gin.Context) {
	var (
		writer  gin.ResponseWriter
		request *http.Request
		conn    *websocket.Conn
		err     error
	)

	writer = ctx.Writer
	request = ctx.Request

	conn, err = server.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		ctx.JSON(500, server.ErrorResult(err.Error()))
		return
	}

	go server.HandleConnection(conn)
}

func (server *Server) HandleConnection(conn *websocket.Conn) {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		request  proto.Request
		outgoing chan generic.Pair[int, []byte]
		kind     int
		data     []byte
		err      error
	)

	ctx, cancel = context.WithCancel(context.Background())
	outgoing = make(chan generic.Pair[int, []byte], 1024)

	defer conn.Close()
	defer cancel()

	go func() {
		var (
			err error
		)

		defer close(outgoing)

		for {
			select {
			case <-ctx.Done():
				{
					return
				}

			case msg := <-outgoing:
				{
					err = conn.WriteMessage(msg.A(), msg.B())
					if err != nil {
						cancel()
						return
					}

					break
				}
			}
		}
	}()

	for {
		kind, data, err = conn.ReadMessage()
		if err != nil {
			break
		}
		_ = kind

		err = json.Unmarshal(data, &request)
		if err != nil {
			break
		}

		go func(request proto.Request) {
			var (
				promise *proto.Promise[proto.Result]
				result  proto.Result
				data    []byte
				flag    bool
				err     error
			)

			promise, flag = server.executor.PushRequest(request)
			if !flag {
				return
			}

			result, err = promise.Await()
			if err != nil {
				return
			}

			data, err = json.Marshal(result)
			if err != nil {
				return
			}

			select {
			case <-ctx.Done():
				return

			default:
				outgoing <- generic.NewPair(websocket.TextMessage, data)
			}
		}(request)
	}
}
