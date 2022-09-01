package net

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"net/http"
	"sync"
)

type WebsocketConnection struct {
	client    *WebsocketClient
	conn      *websocket.Conn
	ID        string
	promiseID uint64

	sync.Mutex
}

func (c *WebsocketConnection) Init(client *WebsocketClient) *WebsocketConnection {
	c.client = client
	c.conn = nil
	c.promiseID = 0
	return c
}

func (c *WebsocketConnection) PromiseID() (ID string) {
	c.Lock()
	ID = fmt.Sprintf("%d", c.promiseID)
	c.promiseID++
	c.Unlock()
	return
}

func (c *WebsocketConnection) Dial() (err error) {
	var (
		data []byte
		hdr  http.Header
		res  *http.Response
	)

	c.Lock()
	defer c.Unlock()

	hdr = http.Header{}
	hdr.Add("TK", c.client.token)

	c.conn, res, err = websocket.DefaultDialer.Dial(c.client.endpoint, hdr)
	if err != nil {
		c.conn = nil
		return
	}

	_ = res

	_, data, err = c.conn.ReadMessage()
	if err != nil {
		c.conn = nil
		return
	}

	c.ID = string(data)

	return
}

func (c *WebsocketConnection) Close() (err error) {
	c.Lock()
	defer c.Unlock()

	err = c.conn.Close()
	if err != nil {
		return
	}
	return
}

func (c *WebsocketConnection) message() (data []byte, err error) {
	_, data, err = c.conn.ReadMessage()
	return
}

func (c *WebsocketConnection) execute(data []byte) (err error) {
	var _res proto.ExecuteResult

	c.Lock()
	defer c.Unlock()

	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return
	}

	data, err = c.message()
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &_res)
	if err != nil {
		return
	}

	c.client.complete(_res)

	return
}
