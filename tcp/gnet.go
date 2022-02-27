package tcp

import (
	"fmt"
	"github.com/panjf2000/gnet/v2"
	"log"
	"strconv"
	"sync"
	"time"
)

type httpServer struct {
	gnet.BuiltinEventEngine
	eng      gnet.Engine
	httpPool sync.Map
}

type httpConn struct {
	conn gnet.Conn
}

func (hs *httpServer) OnBoot(eng gnet.Engine) gnet.Action {
	hs.eng = eng
	log.Printf("echo server with multi-core=%t is listening on %s\n")
	return gnet.None
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	//buf, _ := c.Next(-1)
	//c.Write(buf)
	addr := c.RemoteAddr().String()
	//fmt.Println("react", addr, string(body))
	if _, ok := hs.httpPool.Load(addr); ok {
		go func() {
			time.Sleep(time.Second * 2)
			b := response("200 OK", "", "HELLO:"+addr)
			c.Write(b)
		}()
	}

	return gnet.None
}

func (hs *httpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	addr := c.RemoteAddr().String()
	fmt.Println("conn", addr)
	conn := &httpConn{conn: c}
	hs.httpPool.Store(addr, conn)
	return
}

func (hs *httpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	addr := c.RemoteAddr().String()
	fmt.Println("close", addr)
	if _, ok := hs.httpPool.Load(addr); ok {
		hs.httpPool.Delete(addr)
	}
	return
}

func Serve(port string) error {
	http := new(httpServer)
	//go func() {
	//	for {
	//		time.Sleep(time.Second)
	//		fmt.Println(http)
	//	}
	//}()
	// Start serving!
	return gnet.Run(http, "tcp://:"+port, gnet.WithMulticore(true), gnet.WithTCPKeepAlive(time.Second*60))
}

func response(status, head, body string) []byte {
	var b []byte
	b = append(b, "HTTP/1.1"...)
	b = append(b, ' ')
	b = append(b, status...)
	b = append(b, '\r', '\n')
	b = append(b, "Server: MBK 1.01\r\n"...)
	b = append(b, "Date: "...)
	b = time.Now().AppendFormat(b, "Mon, 02 Jan 2006 15:04:05 GMT")
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, "Content-Length: "...)
		b = strconv.AppendInt(b, int64(len(body)), 10)
		b = append(b, '\r', '\n')
	}
	b = append(b, head...)
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, body...)
	}
	return b
}
