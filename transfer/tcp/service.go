package tcp

import (
	"ahs_server/module"
	"ahs_server/transfer"
	"errors"
	"log"
	"sync"
	"time"
)

func init() {
	transfer.Register("tcp", service{})
}

type service struct {
	module.Api
}

type Server struct {
	Addr    string
	Port    string
	Clients sync.Map
	Handler func(conn module.HttpConn, data []byte)
}

var serve = &Server{}

func (s service) Run(name, port string) (err error) {
	serve.Port = port
	if err = Serve(serve); err != nil {
		log.Println("start tcp server error", err)
	}
	return err
}

func (s service) Handler(receive func(conn module.HttpConn, data []byte)) {
	serve.Handler = receive
}

func (s service) Shutdown() error {
	return nil
}

func (s service) Send(data []byte) (err error) {
	var min int
	var client *Client
	serve.Clients.Range(func(k, v interface{}) bool {
		m := v.(*Client)
		if m.ConnectStatus {
			if m.frequency == 0 {
				client = m
				return false
			}
			if min == 0 || m.frequency < min {
				min = m.frequency
				client = m
			}
			if m.frequency > min {
				return false
			}
		}
		return true
	})
	if client == nil {
		return errors.New("no data")
	}
	err = client.Write(data)
	if err == nil {
		client.frequency++
	}
	return
}

func (s service) Heartbeat(data string, interval int) {
	for {
		time.Sleep(time.Second * time.Duration(interval))
		serve.Clients.Range(func(k, v interface{}) bool {
			v.(*Client).Write([]byte(data))
			return true
		})
	}
}

func (s service) CheckIdleTime(maxTimeout int) {
	max := float64(maxTimeout)
	for {
		time.Sleep(time.Second * 3)
		now := time.Now()
		serve.Clients.Range(func(k, v interface{}) bool {
			client := v.(*Client)
			if now.Sub(client.MaxIdleTime).Seconds() > max {
				client.ConnectStatus = false
				return true
			}
			if client.frequency > 999999999 {
				reloadFrequency()
			}
			return true
		})
	}
}

func reloadFrequency() {
	serve.Clients.Range(func(k, v interface{}) bool {
		v.(*Client).frequency = 0
		return true
	})
}
