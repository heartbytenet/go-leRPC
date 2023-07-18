package lerpc

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"log"
	"sync"
	"time"
)

type Server struct {
	Settings *ServerSettings

	fiberApp *fiber.App
	Binary   *Binary

	handlers map[string][]func(cmd *proto.ExecuteCommand, res *proto.ExecuteResult) // Todo: move this to handler struct
	clientsL sync.Mutex
	clients  []string

	functionOnReceive func(cmd *proto.ExecuteCommand)
}

type ServerSettings struct {
	Port         int
	CommandsFile string
	Tls          bool
	TlsCert      tls.Certificate
}

// Default settings for server
func (s *ServerSettings) Default() *ServerSettings {
	s.Port = 8000
	return s
}

// Init the rpc server
func (s *Server) Init(settings ...*ServerSettings) *Server {
	if len(settings) < 1 {
		s.Settings = (&ServerSettings{}).Default()
	} else {
		s.Settings = settings[0]
	}

	s.fiberApp = fiber.New(fiber.Config{
		Prefork:               false,
		DisableStartupMessage: true,
	})

	s.Binary = (&Binary{}).Init(s)

	s.handlers = map[string][]func(cmd *proto.ExecuteCommand, res *proto.ExecuteResult){}
	s.clientsL = sync.Mutex{}
	s.clients = make([]string, 0)

	return s
}

// Start the rpc server
func (s *Server) Start() (err error) {
	err = s.Binary.Start()
	if err != nil {
		return
	}

	s.route()

	s.fiberApp.Hooks().OnListen(s.hookListen)

	s.fiberApp.Use(cors.New(cors.Config{
		Next:             nil,
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "",
		AllowCredentials: false,
		ExposeHeaders:    "",
		MaxAge:           0,
	}))

	if s.Settings.Tls {
		err = s.fiberApp.ListenTLSWithCertificate(fmt.Sprintf(":%d", s.Settings.Port), s.Settings.TlsCert)
		if err != nil {
			return
		}
	} else {
		err = s.fiberApp.Listen(fmt.Sprintf(":%d", s.Settings.Port))
		if err != nil {
			return
		}
	}

	return
}

// Stop the rpc server
func (s *Server) Stop() (err error) {
	err = s.fiberApp.Shutdown()
	if err != nil {
		return
	}

	return
}

// HookOnListen
// registers hook function to execute on listening
func (s *Server) HookOnListen(fn func() error) {
	s.fiberApp.Hooks().OnListen(func(data fiber.ListenData) (err error) {
		err = fn()
		if err != nil {
			return
		}

		return
	})
}

// HookOnShutdown
// registers hook function to execute on shutdown
func (s *Server) HookOnShutdown(fn func() error) {
	s.fiberApp.Hooks().OnShutdown(fn)
}

// HookOnReceive
// registers hook function to execute after receiving command and before its execution
func (s *Server) HookOnReceive(fn func(cmd *proto.ExecuteCommand)) {
	s.functionOnReceive = fn
}

// RegisterHandler
// attaches a command handler on top of the list of existing ones
func (s *Server) RegisterHandler(path string, handler func(cmd *proto.ExecuteCommand, res *proto.ExecuteResult)) {
	var flag bool

	_, flag = s.handlers[path]
	if !flag {
		s.handlers[path] = make([]func(cmd *proto.ExecuteCommand, res *proto.ExecuteResult), 0)
	}

	s.handlers[path] = append(s.handlers[path], handler)
}

func (s *Server) RegisterBinary(key string, data []byte, contentType ...string) {
	var (
		_type string
		entry *BinaryEntry
	)

	if len(contentType) > 0 {
		_type = contentType[0]
	} else {
		_type = ""
	}

	entry = (&BinaryEntry{}).
		Init(_type).
		FromBytes(data)

	s.Binary.RegisterEntry(key, entry)
}

func (s *Server) clientNew() (ID string) {
	var flag bool

	s.clientsL.Lock()
	defer s.clientsL.Unlock()

	ts := func() string { return fmt.Sprintf("%x", time.Now().UnixNano()) }

	ID = ts()
	for {
		flag = false
		for _, v := range s.clients {
			if v == ID {
				flag = true
				break
			}
		}
		if !flag {
			break
		}
		ID = ts()
	}

	return
}

func (s *Server) clientDel(ID string) {
	clients := make([]string, 0)

	s.clientsL.Lock()
	defer s.clientsL.Unlock()

	for _, val := range s.clients {
		if val != ID {
			clients = append(clients, val)
		}
	}

	s.clients = clients
}

func (s *Server) hookListen(data fiber.ListenData) (err error) {
	log.Printf("server listening on :%d\n", s.Settings.Port)
	return
}

func (s *Server) route() {
	s.fiberApp.Get("/", s.routeIndex)

	s.fiberApp.Get("/binary", s.Binary.handleRoute)
	s.fiberApp.Post("/binary", s.Binary.handleRoute)

	s.fiberApp.Post("/execute", s.routeExecute)

	s.fiberApp.Use("/connect", s.routeConnectUpgrade)
	s.fiberApp.Get("/connect", websocket.New(s.routeConnect))
}

func (s *Server) routeIndex(ctx *fiber.Ctx) (err error) {
	return ctx.JSON(
		(&proto.ExecuteResult{}).
			ToPayload(map[string]interface{}{}))
}

func (s *Server) routeExecute(ctx *fiber.Ctx) (err error) {
	var (
		cmd proto.ExecuteCommand
		res proto.ExecuteResult
	)

	err = ctx.BodyParser(&cmd)
	if err != nil {
		res.Success = false
		res.Payload = nil
		res.Error = "failed at parsing request body"
		return ctx.JSON(res)
	}

	s.Execute(&cmd, &res)

	return ctx.JSON(res)
}

func (s *Server) routeConnectUpgrade(ctx *fiber.Ctx) (err error) {
	if websocket.IsWebSocketUpgrade(ctx) {
		ctx.Locals("ID", s.clientNew())
		return ctx.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (s *Server) routeConnect(conn *websocket.Conn) {
	var (
		data     []byte
		clientID string
		flag     bool
		err      error
	)

	stop := func(conn *websocket.Conn) {
		var err error

		err = conn.Close()
		if err != nil {
			log.Println("failed to close websocket connection", err)
		}
	}

	if conn.Locals("ID") == nil {
		log.Println("failed at getting client ID")
		stop(conn)
		return
	}

	clientID, flag = conn.Locals("ID").(string)
	if !flag {
		log.Println("failed at asserting client ID type")
		stop(conn)
		return
	}

	err = conn.WriteMessage(1, []byte(clientID))
	if err != nil {
		log.Println("failed at sending client ID", err)
		stop(conn)
		return
	}

	defer s.clientDel(clientID)

	for {
		_, data, err = conn.ReadMessage()
		if err != nil {
			log.Println("failed at reading websocket message", err)
			break
		}

		cmd := proto.ExecuteCommand{}
		res := proto.ExecuteResult{}

		err = json.Unmarshal(data, &cmd)
		if err != nil {
			log.Println("failed at unmarshalling websocket message", err)
			break
		}

		s.Execute(&cmd, &res)

		data, err = json.Marshal(res)
		if err != nil {
			log.Println("failed at marshalling websocket message", res)
			break
		}

		err = conn.WriteMessage(1, data)
		if err != nil {
			log.Println("failed at writing websocket message", err)
			break
		}
	}

	stop(conn)
	return
}

func (s *Server) Execute(cmd *proto.ExecuteCommand, res *proto.ExecuteResult) {
	var (
		callback chan byte
		handlers []func(cmd *proto.ExecuteCommand, res *proto.ExecuteResult)
		flag     bool
	)

	res.ID = cmd.ID

	// Todo: command checks

	if s.functionOnReceive != nil {
		s.functionOnReceive(cmd)
	}

	handlerKey := fmt.Sprintf("%s::%s", cmd.Namespace, cmd.Method)
	handlers, flag = s.handlers[handlerKey]
	if !flag {
		res.ToError(fmt.Sprintf("handler not found %s", handlerKey))
		return
	}

	if len(handlers) < 1 {
		res.ToError(fmt.Sprintf("handler not found %s", handlerKey))
		return
	}

	callback = make(chan byte, 1)

	go func() {
		for _, handler := range handlers {
			handler(cmd, res)
		}

		callback <- 42
	}()

	<-callback
}

func (s *Server) Exec(cmd *proto.ExecuteCommand) *proto.ExecuteResult {
	var res proto.ExecuteResult
	s.Execute(cmd, &res)
	return &res
}
