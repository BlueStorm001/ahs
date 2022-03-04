package tcp

import (
	"ahs_server/module"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type httpServer struct {
	gnet.BuiltinEventEngine
	eng   gnet.Engine
	serve *Server
}

type Client struct {
	RemoteAddr    string
	conn          gnet.Conn
	ConnectStatus bool
	MaxIdleTime   time.Time
	frequency     int
}

func (hs *httpServer) Load(addr string) (*Client, bool) {
	v, ok := hs.serve.Clients.Load(addr)
	if !ok {
		return nil, false
	}
	return v.(*Client), true
}

func (hs *httpServer) Store(addr string, client *Client) {
	hs.serve.Clients.Store(addr, client)
}

func (hs *httpServer) Delete(addr string) {
	hs.serve.Clients.Delete(addr)
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
	hs.Store(addr, &Client{
		conn:          c,
		ConnectStatus: true,
	})
	fmt.Println("conn", addr)
	return
}

func (hs *httpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	addr := c.RemoteAddr().String()
	hs.Delete(addr)
	fmt.Println("close", addr)
	return
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	addr := c.RemoteAddr().String()
	client, ok := hs.Load(addr)
	if !ok {
		return gnet.Close
	}
	client.conn = c
	if err := hs.Read(client); err != nil {
		client.ConnectStatus = false
		return gnet.Close
	}
	return gnet.None
}

func (hs *httpServer) Read(client *Client) error {
	var buf [4]byte
	if _, err := client.conn.Read(buf[:4]); err != nil {
		return err
	}
	var pkgLen uint32
	pkgLen = binary.BigEndian.Uint32(buf[:4])
	var data = make([]byte, pkgLen)
	if _, err := client.conn.Read(data); err != nil {
		return err
	}
	client.MaxIdleTime = time.Now()
	if len(data) == 3 && data[0] == 'o' && data[1] == 'k' && data[2] == '?' {
		fmt.Println(client.RemoteAddr, "ok?")
		return nil
	}
	if len(data) == 3 && data[0] == '?' && data[1] == 'n' && data[2] == 'o' {
		fmt.Println(client.RemoteAddr, "?no")
		return errors.New("connection error")
	}
	conn := module.HttpConn{Handler: client, RemoteAddr: client.RemoteAddr}
	go hs.serve.Handler(conn, data)
	return nil
}

func Serve(serve *Server) error {
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

func (client *Client) Write(data []byte) error {
	if !client.ConnectStatus {
		return errors.New("connection failed")
	}
	_, err := client.conn.Write(data)
	return err
}

func (client *Client) Close() error {
	client.conn.Close(func(c gnet.Conn) error {
		return nil
	})
	return nil
}
