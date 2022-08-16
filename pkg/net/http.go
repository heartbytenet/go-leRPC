package net

import (
	"bytes"
	"io"
	"net/http"
)

type HttpClient struct {
	client *http.Client
}

func (c *HttpClient) Init() *HttpClient {
	c.client = &http.Client{}
	return c
}

func (c *HttpClient) Execute(method string, url string, body []byte, header http.Header) (data []byte, err error) {
	var (
		req *http.Request
		res *http.Response
		buf *bytes.Buffer
	)

	buf = nil
	if body != nil {
		buf = bytes.NewBuffer(body)
	}

	req, err = http.NewRequest(method, url, buf)
	if err != nil {
		return
	}

	if header != nil {
		req.Header = header
	}

	res, err = c.client.Do(req)
	if err != nil {
		return
	}

	if res.Body != nil {
		data, err = io.ReadAll(res.Body)
	}

	_ = res.Body.Close()

	return
}
