package net

import (
	"encoding/json"
	"fmt"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type WebsocketClient struct {
	endpoint     string
	secure       *uint32
	token        string
	connections  []*WebsocketConnection
	promises     map[string]*WebsocketPromise
	promisesLock sync.Mutex
}

func (c *WebsocketClient) Init(endpoint string, secure *uint32, token string) *WebsocketClient {
	c.endpoint = endpoint
	c.secure = secure
	c.token = token
	c.connections = make([]*WebsocketConnection, 0)
	c.promises = map[string]*WebsocketPromise{}
	c.promisesLock = sync.Mutex{}
	return c
}

func (c *WebsocketClient) Secure() uint32 {
	return atomic.LoadUint32(c.secure)
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

	c.promisesLock.Lock()
	c.promises[cmd.ID] = promise
	c.promisesLock.Unlock()

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
	c.promisesLock.Lock()
	promise, flag := c.promises[res.ID]
	c.promisesLock.Unlock()
	if !flag {
		log.Println("promise not found with id", res.ID)
		return
	}
	*promise.result = res
	promise.callback <- 42

	c.promisesLock.Lock()
	ts := time.Now().UnixMilli()
	rm := make([]string, 0)
	for id, promise := range c.promises {
		if (ts - promise.creation) >= (60 * 1000) {
			rm = append(rm, id)
		}
	}
	for _, id := range rm {
		delete(c.promises, id)
	}
	c.promisesLock.Unlock()
}
