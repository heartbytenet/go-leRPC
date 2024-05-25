package server

import (
	"fmt"
	"log"

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

	engine *gin.Engine
}

func NewServer() *Server {
	return &Server{
		settings: NewSettingsDefault(),
		executor: NewExecutor(),

		engine: gin.Default(),
	}
}

func (server *Server) AddHandler(handler Handler) {
	server.executor.AddHandler(handler)
}

func (server *Server) Addr() string {
	return fmt.Sprintf(":%d", server.settings.Port)
}

func (server *Server) Run() (err error) {
	err = server.executor.Start()
	if err != nil {
		log.Fatalln("failed at starting executor:", err)
	}

	server.engine.POST("/execute", server.HandleExecute)

	err = server.engine.Run(server.Addr())
	if err != nil {
		log.Fatalln("failed at running server:", err)
	}

	return
}

func (server *Server) ErrorResult(err error) (result proto.Result) {
	result = proto.NewResult().
		WithCode(proto.ResultCodeError)

	if debug.DEBUG {
		return result.WithMessage(err.Error())
	} else {
		return result.WithMessage(FallbackErrorMessage)
	}
}

func (server *Server) HandleExecute(ctx *gin.Context) {
	var (
		request proto.Request
		result  proto.Result
		promise proto.Promise[proto.Result]
		err     error
	)

	err = ctx.BindJSON(&request)
	if err != nil {
		result = server.ErrorResult(err)
		ctx.JSON(200, result)
		return
	}

	promise = server.executor.PushRequest(request)
	result = promise.Await()
	ctx.JSON(200, result)
}
