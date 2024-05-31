package server

import (
	"io"
	"os"
	"time"
)

type DownloadHandler interface {
	Match(key string) bool
	Lifetime() time.Duration
	Limit() int
	GetIsAlive() bool
	GetContentType() string
	Pull() (io.Reader, error)
	Remove() error
}

type DownloadHandlerFile struct {
	key      string
	lifetime time.Duration
	limit    int
	count    int
	isAlive  bool

	contentType string
	path        string
}

func NewDownloadHandlerFile(
	key string,
	lifetime time.Duration,
	limit int,
	contentType string,
	path string,
) *DownloadHandlerFile {
	return &DownloadHandlerFile{
		key:      key,
		lifetime: lifetime,
		limit:    limit,
		count:    0,
		isAlive:  true,

		contentType: contentType,
		path:        path,
	}
}

func (handler *DownloadHandlerFile) Match(key string) bool {
	return handler.key == key
}

func (handler *DownloadHandlerFile) Lifetime() time.Duration {
	return handler.lifetime
}

func (handler *DownloadHandlerFile) Limit() int {
	return handler.limit
}

func (handler *DownloadHandlerFile) GetIsAlive() bool {
	if handler.count == handler.Limit() {
		handler.isAlive = false
	}

	return handler.isAlive
}

func (handler *DownloadHandlerFile) GetContentType() string {
	return handler.contentType
}

func (handler *DownloadHandlerFile) Pull() (reader io.Reader, err error) {
	var (
		file *os.File
	)

	handler.count++

	file, err = os.Open(handler.path)
	if err != nil {
		return
	}

	reader = file
	return
}

func (handler *DownloadHandlerFile) Remove() (err error) {
	err = os.Remove(handler.path)
	if err != nil {
		return
	}

	return
}
