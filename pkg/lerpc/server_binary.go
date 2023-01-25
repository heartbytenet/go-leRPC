package lerpc

import (
	"github.com/gofiber/fiber/v2"
	"github.com/heartbytenet/go-lerpc/pkg/proto"
	"log"
	"os"
	"sync"
	"time"
)

const (
	BinaryTypeNone  BinaryType = iota
	BinaryTypeFile             = iota
	BinaryTypeBytes            = iota
)

type BinaryType int

type Hook func(*Binary)

type Binary struct {
	Server *Server

	Timeout int64

	binariesData map[string]*BinaryEntry
	binariesLock sync.Mutex

	hooksData map[string]Hook
	hooksLock sync.Mutex
}

type BinaryEntry struct {
	dataType BinaryType
	dataBody []byte
	dataPath string

	contentType  string
	timeCreation int64
}

func (entry *BinaryEntry) Init(contentType string) *BinaryEntry {
	entry.dataType = BinaryTypeNone

	entry.contentType = contentType
	entry.timeCreation = time.Now().UnixMilli()

	return entry
}

func (entry *BinaryEntry) FromFile(path string) *BinaryEntry {
	entry.dataType = BinaryTypeFile
	entry.dataPath = path

	return entry
}

func (entry *BinaryEntry) FromBytes(data []byte) *BinaryEntry {
	entry.dataType = BinaryTypeBytes
	entry.dataBody = data[:]

	return entry
}

func (entry *BinaryEntry) Data() (contentType string, timeCreation int64) {
	contentType = entry.contentType
	timeCreation = entry.timeCreation

	return
}

func (entry *BinaryEntry) Bytes() (data []byte, err error) {
	switch entry.dataType {
	case BinaryTypeBytes:
		data = entry.dataBody[:]
		return

	case BinaryTypeFile:
		data, err = os.ReadFile(entry.dataPath)
		if err != nil {
			return
		}

		return

	default:
		return
	}
}

func (binary *Binary) Init(server *Server) *Binary {
	binary.Server = server

	binary.Timeout = 1000 * 60

	binary.binariesData = map[string]*BinaryEntry{}
	binary.binariesLock = sync.Mutex{}

	binary.hooksData = map[string]Hook{}
	binary.hooksLock = sync.Mutex{}

	return binary
}

func (binary *Binary) Start() (err error) {
	go binary.routineClean()

	return
}

func (binary *Binary) Clean() {
	var (
		ts   int64
		keys []string
	)

	ts = time.Now().UnixMilli()
	keys = make([]string, 0)

	binary.binariesLock.Lock()
	defer binary.binariesLock.Unlock()

	for key, entry := range binary.binariesData {
		if (ts - entry.timeCreation) >= binary.Timeout {
			keys = append(keys, key)
		}
	}

	for _, key := range keys {
		delete(binary.binariesData, key)
	}
}

func (binary *Binary) routineClean() {
	var (
		ticker *time.Ticker
	)

	ticker = time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			break
		}

		binary.Clean()
	}
}

func (binary *Binary) RegisterEntry(key string, entry *BinaryEntry) {
	binary.binariesLock.Lock()
	defer binary.binariesLock.Unlock()

	binary.binariesData[key] = entry
}

func (binary *Binary) GetEntry(key string) (entry *BinaryEntry) {
	var (
		flag bool
	)

	binary.binariesLock.Lock()
	defer binary.binariesLock.Unlock()

	entry, flag = binary.binariesData[key]
	if !flag {
		entry = nil
		return
	}

	return
}

func (binary *Binary) RegisterHook(key string, hook Hook) {
	binary.hooksLock.Lock()
	defer binary.hooksLock.Unlock()

	binary.hooksData[key] = hook
}

func (binary *Binary) ExecuteHook(key string) {
	var (
		hook Hook
		flag bool
	)

	binary.hooksLock.Lock()
	defer binary.hooksLock.Unlock()

	hook, flag = binary.hooksData[key]
	if flag {
		hook(binary)
	}
}

func (binary *Binary) handleRoute(ctx *fiber.Ctx) (err error) {
	var (
		res      *proto.ExecuteResult
		data     []byte
		_type    string
		entryKey string
		entryVal *BinaryEntry
	)

	res = &proto.ExecuteResult{}

	entryKey = ctx.Query("key", "")
	if entryKey == "" {
		return ctx.JSON(res.ToError("missing entry key"))
	}

	binary.ExecuteHook(entryKey)

	entryVal = binary.GetEntry(entryKey)
	if entryVal == nil {
		return ctx.JSON(res.ToError("entry not found"))
	}

	_type, _ = entryVal.Data()
	ctx.Set("Content-Type", _type)

	data, err = entryVal.Bytes()
	if err != nil {
		log.Println("failed at getting binary entry bytes", err)
		return ctx.JSON(res.ToError("invalid entry body"))
	}

	return ctx.Send(data)
}
