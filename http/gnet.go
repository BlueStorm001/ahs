package http

import (
	"ahs_server/module"
	"context"
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type httpServer struct {
	gnet.BuiltinEventEngine
	eng   gnet.Engine
	serve *module.HttpServer
}

type httpConn struct {
	*module.HttpConn
	conn gnet.Conn
}

func (hs *httpServer) Load(addr string) (*module.HttpConn, bool) {
	v, ok := hs.serve.Requester.Load(addr)
	if !ok {
		return nil, false
	}
	return v.(*module.HttpConn), true
}

func (hs *httpServer) OnBoot(eng gnet.Engine) gnet.Action {
	if hs.serve == nil {
		return gnet.Shutdown
	}
	hs.eng = eng
	fmt.Println("listening:", hs.serve.Addr)
	return gnet.None
}

func (hs *httpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	addr := c.RemoteAddr().String()
	hs.serve.Requester.Store(addr, &module.HttpConn{
		RemoteAddr:     addr,
		Handler:        &httpConn{conn: c},
		ConnectionTime: time.Now(),
	})
	fmt.Println("conn", addr)
	return
}

func (hs *httpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	addr := c.RemoteAddr().String()
	http, ok := hs.Load(addr)
	if !ok {
		return gnet.Close
	}
	http.Closed = true
	fmt.Println("close", addr)
	return
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	addr := c.RemoteAddr().String()
	http, ok := hs.Load(addr)
	if !ok {
		return gnet.Close
	}
	http.ReactTime = time.Now()
	var data []byte
	buf, _ := c.Next(-1)
	data = append(data, buf...)
	go hs.serve.Handler(http, data)
	return gnet.None
}

func Serve(serve *module.HttpServer) error {
	http := new(httpServer)
	http.serve = serve
	http.serve.Addr = "tcp://:" + serve.Port
	err := gnet.Run(http, http.serve.Addr, gnet.WithMulticore(true))
	if err != nil {
		fmt.Println("connection error", err)
	}
	return err
}

func Stop(serve *module.HttpServer) error {
	return gnet.Stop(context.Background(), serve.Addr)
}

func (hs *httpConn) Write(data []byte) error {
	_, err := hs.conn.Write(data)
	return err
}

func (hs *httpConn) Close() error {
	hs.conn.Close(func(c gnet.Conn) error {
		return nil
	})
	return nil
}
