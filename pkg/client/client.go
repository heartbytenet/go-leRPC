package client

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"net/http"
)

type ClientMode int

const (
	ClientModeHttp ClientMode = iota
	ClientModeHttps
	ClientModeWs
	ClientModeWss
)

type Client struct {
	mode   ClientMode
	remote string
	token  string

	httpClient *http.Client
}

func NewClient(mode ClientMode, remote string, token string) *Client {
	return &Client{
		mode:   mode,
		remote: remote,
		token:  token,

		httpClient: &http.Client{},
	}
}

func (client *Client) GetMode() ClientMode {
	return client.mode
}

func (client *Client) GetRemote() string {
	return client.remote
}

func (client *Client) GetToken() string {
	return client.token
}

func (client *Client) GetUrl(mode ClientMode) string {
	switch mode {
	case ClientModeHttp:
		return fmt.Sprintf("http://%s/execute", client.remote)

	case ClientModeHttps:
		return fmt.Sprintf("https://%s/execute", client.remote)

	case ClientModeWs:
		return fmt.Sprintf("ws://%s/connect", client.remote)

	case ClientModeWss:
		return fmt.Sprintf("wss://%s/connect", client.remote)

	default:
		panic("invalid mode")
	}
}

func (client *Client) Execute(request proto.Request) (promise *proto.Promise[proto.Result], err error) {
	mode := client.GetMode()

	switch mode {
	case ClientModeHttp, ClientModeHttps:
		{
			promise = client.ExecuteHttp(mode, request)
			return
		}

	default:
		panic("not implemented")
	}
}

func (client *Client) ExecuteSync(request proto.Request) (result proto.Result, err error) {
	var (
		promise *proto.Promise[proto.Result]
	)

	promise, err = client.Execute(request)
	if err != nil {
		return
	}

	result, err = promise.Await()
	if err != nil {
		return
	}

	return
}

func (client *Client) ExecuteHttp(mode ClientMode, request proto.Request) (promise *proto.Promise[proto.Result]) {
	promise = proto.NewPromise[proto.Result]()

	go func(request proto.Request) {
		var (
			result proto.Result
			req    *http.Request
			res    *http.Response
			data   []byte
			err    error
		)

		data, err = json.Marshal(request)
		if err != nil {
			promise.Failed(err)
			return
		}

		req, err = http.NewRequest(
			"POST",
			client.GetUrl(mode),
			bytes.NewReader(data))
		if err != nil {
			promise.Failed(err)
			return
		}

		res, err = client.httpClient.Do(req)
		if err != nil {
			promise.Failed(err)
			return
		}

		if res.Body == nil {
			err = fmt.Errorf("response body is empty")
			promise.Failed(err)
			return
		}

		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			promise.Failed(err)
			return
		}

		promise.Complete(result)
		return
	}(request)

	return
}
