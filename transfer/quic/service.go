package quic

import (
	"ahs_server/module"
	"ahs_server/transfer"
	"errors"
	"fmt"
	"time"
)

func init() {
	transfer.Register("quic", service{})
}

type service struct {
	module.Api
}

var serve = &Server{}

func (s service) Run(name, port string) error {
	serve.NewServer(name)
	return serve.Listens("localhost:" + port)
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
	fmt.Println("clients", len(serve.Clients))
	for _, c := range serve.Clients {
		if !c.ConnectStatus {
			continue
		}
		if c.frequency == 0 {
			client = c
			break
		}
		if min == 0 || c.frequency < min {
			min = c.frequency
			client = c
		}
		if c.frequency > min {
			break
		}
	}
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
		for _, client := range serve.Clients {
			if err := client.Write([]byte(data)); err != nil {
				client.ConnectStatus = false
			}
		}
	}
}

func (s service) CheckIdleTime(maxTimeout int) {
	max := float64(maxTimeout)
	for {
		time.Sleep(time.Second * 3)
		now := time.Now()
		for _, client := range serve.Clients {
			if now.Sub(client.MaxIdleTime).Seconds() > max {
				client.ConnectStatus = false
				continue
			}
			if client.frequency > 999999999 {
				reloadFrequency()
			}
		}
	}
}

func reloadFrequency() {
	for _, client := range serve.Clients {
		client.frequency = 0
	}
}
