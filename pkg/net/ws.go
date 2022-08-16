package net

import "github.com/gorilla/websocket"

type WebsocketClient struct {
	Conn *websocket.Conn
}

func (c *WebsocketClient) Init() *WebsocketClient {
	return c
}
