package lerpc

import (
	"encoding/json"
	"fmt"
	"github.com/heartbytenet/go-lerpc/pkg/net"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"strings"
	"sync/atomic"
)

const (
	ClientModeBalanced      = 0
	ClientModeHttpOnly      = 1
	ClientModeWebsocketOnly = 2
)

type Client struct {
	url             string
	mode            uint32
	secure          uint32
	token           string
	clientHttp      *net.HttpClient
	clientWebsocket *net.WebsocketClient
}

func (c *Client) Init(url string, token string) *Client {
	c.url = url
	c.mode = ClientModeBalanced
	c.secure = 1
	c.token = token
	c.clientHttp = (&net.HttpClient{}).Init()
	c.clientWebsocket = (&net.WebsocketClient{}).Init(url, &c.secure, token)
	return c
}

func (c *Client) Start(connections int) (err error) {
	err = c.clientWebsocket.Start(connections)
	if err != nil {
		return
	}
	return
}

func (c *Client) Mode(val *uint32) uint32 {
	if val != nil {
		atomic.StoreUint32(&c.mode, *val)
		return *val
	}
	return atomic.LoadUint32(&c.mode)
}

func (c *Client) Secure(val *uint32) uint32 {
	if val != nil {
		atomic.StoreUint32(&c.secure, *val)
		return *val
	}
	return atomic.LoadUint32(&c.secure)
}

func (c *Client) ExecuteMode(cmd *proto.ExecuteCommand, res *proto.ExecuteResult, mode uint32) (err error) {
	var (
		callback chan byte
		data     []byte
	)

	switch mode {
	case ClientModeHttpOnly:
		{
			cmd.Token = c.token

			data, err = json.Marshal(cmd)
			if err != nil {
				return
			}
			data, err = c.clientHttp.Execute(
				"POST", fmt.Sprintf("http%s://%s/execute",
					strings.Repeat("s", int(c.Secure(nil))), c.url), data, map[string][]string{
					"Content-Type": {"application/json"},
				})
			if err != nil {
				return
			}
			err = json.Unmarshal(data, res)
			if err != nil {
				return
			}
		}
	case ClientModeWebsocketOnly:
		{
			callback, err = c.clientWebsocket.Execute(cmd, res)
			if err != nil {
				return
			}
			<-callback
		}
	}

	return
}

func (c *Client) Execute(cmd *proto.ExecuteCommand, res *proto.ExecuteResult) (err error) {
	switch c.Mode(nil) {
	case ClientModeBalanced:
		{
			err = c.ExecuteMode(cmd, res, ClientModeWebsocketOnly)
			if err == nil {
				break
			}
			err = c.ExecuteMode(cmd, res, ClientModeHttpOnly)
			if err != nil {
				return
			}
		}
	case ClientModeHttpOnly:
		{
			return c.ExecuteMode(cmd, res, ClientModeHttpOnly)
		}
	case ClientModeWebsocketOnly:
		{
			return c.ExecuteMode(cmd, res, ClientModeWebsocketOnly)
		}
	}

	return
}
