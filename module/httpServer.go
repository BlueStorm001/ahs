package module

import (
	"sync"
	"time"
)

type HttpServer struct {
	Addr      string
	Port      string
	Requester sync.Map
	Handler   func(http *HttpConn, data []byte)
	Mux       sync.Mutex
}

type Handler interface {
	Write(data []byte) error
	Close() error
}

type HttpConn struct {
	RemoteAddr     string
	Handler        Handler
	ConnectionTime time.Time
	ReactTime      time.Time
	Closed         bool
	ResponseStatus int32
}
