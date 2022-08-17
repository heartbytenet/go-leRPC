package lerpc

import (
	"encoding/json"
	"fmt"
	"github.com/heartbytenet/go-lerpc/pkg/net"
	"log"
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
	token           string
	clientHttp      *net.HttpClient
	clientWebsocket *net.WebsocketClient
}

func (c *Client) Init(url string, token string) *Client {
	c.url = url
	c.mode = ClientModeHttpOnly
	c.token = token
	c.clientHttp = (&net.HttpClient{}).Init()
	c.clientWebsocket = (&net.WebsocketClient{}).Init()
	return c
}

func (c *Client) Mode(val *uint32) uint32 {
	if val != nil {
		atomic.StoreUint32(&c.mode, *val)
		return *val
	}
	return atomic.LoadUint32(&c.mode)
}

func (c *Client) Execute(cmd *ExecuteCommand, res *ExecuteResult) (err error) {
	var (
		data []byte
	)

	cmd.Token = c.token

	switch c.Mode(nil) {
	case ClientModeBalanced:
		{
			log.Fatalln("leRPC client mode is unimplemented")
		}
	case ClientModeHttpOnly:
		{
			data, err = json.Marshal(cmd)
			if err != nil {
				return
			}
			data, err = c.clientHttp.Execute(
				"POST", fmt.Sprintf("https://%s/execute", c.url), data, map[string][]string{
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
			log.Fatalln("leRPC client mode is unimplemented")
		}
	}
	return
}
