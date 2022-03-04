package quic

import (
	"ahs_server/module"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"sync"
	"time"
)

type Client struct {
	Name          string
	ConnectStatus bool
	Addr          string
	Handler       func(conn module.HttpConn, data []byte)
	MaxIdleTime   time.Time
	bufferPacket  *Packer
	tls           *tls.Config
	session       quic.Session
	stream        quic.Stream
	mu            sync.Mutex
	check         bool
	frequency     int
}

//func (c *Client) Read() error {
//	var buf [4]byte
//	if _, err := c.stream.Read(buf[:4]); err != nil {
//		return err
//	}
//	var pkgLen uint32
//	pkgLen = binary.BigEndian.Uint32(buf[:4])
//	fmt.Println("pkgLen", pkgLen)
//	var data = make([]byte, pkgLen)
//	if _, err := c.stream.Read(data); err != nil {
//		return err
//	}
//	fmt.Println("data", len(data))
//	c.MaxIdleTime = time.Now()
//	go c.receive(data)
//	return nil
//}

func (c *Client) Read() error {
	buf := make([]byte, 500)
	n, err := c.stream.Read(buf)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	c.MaxIdleTime = time.Now()
	c.bufferPacket.datagram = buf[:n]
	c.bufferPacket.Packet(c.receive)
	return nil
}

func (c *Client) receive(buffer []byte) {
	if len(buffer) == 3 && buffer[0] == 'o' && buffer[1] == 'k' && buffer[2] == '?' {
		fmt.Println(c.Addr, "ok?")
		return
	}
	if len(buffer) == 3 && buffer[0] == '?' && buffer[1] == 'n' && buffer[2] == 'o' {
		fmt.Println(c.Addr, "?no")
		c.ConnectStatus = false
		return
	}
	conn := module.HttpConn{Handler: c, RemoteAddr: c.Addr}
	c.Handler(conn, buffer)
}

func (c *Client) Write(message []byte) (err error) {
	if !c.ConnectStatus {
		return errors.New("connection failed")
	}
	c.mu.Lock()
	if _, err = c.stream.Write(Packer{}.Write(message)); err != nil {
		c.ConnectStatus = false
	}
	c.mu.Unlock()
	return
}

//
//func (c *Client) Write(message []byte) (err error) {
//	if !c.ConnectStatus {
//		return errors.New("connection failed")
//	}
//	fmt.Println("Write", len(message))
//
//	c.mu.Lock()
//	defer c.mu.Unlock()
//	var readLen int
//	pkgLen := uint32(len(message))
//	var buf [4]byte
//	binary.BigEndian.PutUint32(buf[:4], pkgLen)
//	if readLen, err = c.stream.Write(buf[:4]); readLen != 4 && err != nil {
//		if readLen == 0 {
//			return errors.New("数据长度发生异常")
//		}
//		return
//	}
//	// 发送消息
//	readLen, err = c.stream.Write(message)
//	return
//}

func (c *Client) Close() error {
	c.ConnectStatus = false
	if err := c.Write([]byte("?no")); err == nil {
		time.Sleep(time.Second)
	}
	return c.stream.Close()
}
