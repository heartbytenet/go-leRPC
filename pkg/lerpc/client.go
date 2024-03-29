package lerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/heartbytenet/go-lerpc/pkg/net"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"log"
	"strings"
	"sync/atomic"
)

const (
	ClientModeBalanced      = uint32(0)
	ClientModeHttpOnly      = uint32(1)
	ClientModeWebsocketOnly = uint32(2)
)

type Client struct {
	url             string
	mode            uint32
	secure          uint32
	token           string
	clientHttp      *net.HttpClient
	clientWebsocket *net.WebsocketClient
}

func (c *Client) Init(url string, token string, InsecureSkipVerify ...bool) *Client {
	c.url = url
	c.mode = ClientModeHttpOnly
	c.secure = 1
	c.token = token
	if len(InsecureSkipVerify) > 0 {
		c.clientHttp = (&net.HttpClient{Skip: true}).Init()
	} else {
		c.clientHttp = (&net.HttpClient{Skip: false}).Init()
	}
	c.clientWebsocket = (&net.WebsocketClient{}).Init(url, &c.secure, token)
	return c
}

func (c *Client) Start(connections int) (err error) {
	if c.Mode(nil) != ClientModeHttpOnly {
		err = c.clientWebsocket.Start(connections)
		if err != nil {
			return
		}
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

	cmd.Token = c.token

	switch mode {
	case ClientModeHttpOnly:
		{
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

func (c *Client) ExecuteRetries(attempt int, cmd *proto.ExecuteCommand, res *proto.ExecuteResult) (err error) {
	if attempt <= 0 {
		err = errors.New("too many attempts")
		return
	}

	mode := c.Mode(nil)

	switch mode {
	case ClientModeBalanced:
		{
			err = c.ExecuteMode(cmd, res, ClientModeWebsocketOnly)
		}
	case ClientModeHttpOnly:
		{
			err = c.ExecuteMode(cmd, res, ClientModeHttpOnly)
		}
	case ClientModeWebsocketOnly:
		{
			err = c.ExecuteMode(cmd, res, ClientModeWebsocketOnly)
		}
	}

	if err != nil {
		log.Printf("failed at executing command %s::%s, %v. retrying...\n", cmd.Namespace, cmd.Method, err)

		if mode == ClientModeBalanced {
			mode = ClientModeHttpOnly
			c.Mode(&mode)
		}

		return c.ExecuteRetries(attempt-1, cmd, res)
	}

	return nil
}

func (c *Client) Execute(cmd *proto.ExecuteCommand, res *proto.ExecuteResult) (err error) {
	return c.ExecuteRetries(10, cmd, res)
}
