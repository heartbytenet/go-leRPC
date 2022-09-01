package net

import (
	"encoding/json"
	"fmt"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"log"
	"math/rand"
	"time"
)

type WebsocketClient struct {
	endpoint    string
	token       string
	connections []*WebsocketConnection
	promises    map[string]*WebsocketPromise
}

func (c *WebsocketClient) Init(endpoint string, token string) *WebsocketClient {
	c.endpoint = fmt.Sprintf("wss://%s/connect", endpoint)
	c.token = token
	c.connections = make([]*WebsocketConnection, 0)
	c.promises = map[string]*WebsocketPromise{}
	return c
}

func (c *WebsocketClient) Start(connections int) (err error) {
	var connection *WebsocketConnection

	if connections < 1 {
		connections = 1
	}

	for i := 0; i < connections; i++ {
		connection = (&WebsocketConnection{}).Init(c)
		err = connection.Dial()
		if err != nil {
			return
		} else {
			log.Println("client connected", connection.ID)
		}
		c.connections = append(c.connections, connection)
	}

	return
}

func (c *WebsocketClient) Execute(cmd *proto.ExecuteCommand, res *proto.ExecuteResult) (callback chan byte, err error) {
	var (
		promise *WebsocketPromise
		data    []byte
		conn    *WebsocketConnection
	)

	conn = c.connections[rand.Intn(len(c.connections))]

	cmd.ID = fmt.Sprintf("%v%s%s", time.Now().UnixNano(), conn.ID, cmd.ID)

	data, err = json.Marshal(cmd)
	if err != nil {
		return
	}

	promise = (&WebsocketPromise{}).Init(res)

	c.promises[cmd.ID] = promise

	err = conn.execute(data)
	if err != nil {
		go func() {
			var err error

			err = conn.Dial()
			if err != nil {
				log.Println("failed to dial websocket connection", err)
			}
		}()
		return
	}

	callback = promise.callback
	return
}

func (c *WebsocketClient) complete(res proto.ExecuteResult) {
	promise, flag := c.promises[res.ID]
	if !flag {
		log.Println("promise not found with id", res.ID)
		return
	}
	*promise.result = res
	promise.callback <- 42
}
