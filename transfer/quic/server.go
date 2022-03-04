package quic

import (
	"ahs_server/module"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"math/big"
	"sync"
	"time"
)

type Server struct {
	Name     []string
	Err      error
	Clients  map[string]*Client
	Handler  func(conn module.HttpConn, data []byte)
	tls      *tls.Config
	listener quic.Listener
	mu       sync.Mutex
}

var config = &quic.Config{HandshakeIdleTimeout: time.Second * 5, MaxIdleTimeout: time.Second * 30, KeepAlive: true}

func NewServer(name ...string) *Server {
	return &Server{Name: name, tls: generateTLSConfig(name...), Clients: make(map[string]*Client)}
}

func (serv *Server) NewServer(name ...string) *Server {
	serv.Name = name
	serv.tls = generateTLSConfig(name...)
	serv.Clients = make(map[string]*Client)
	return serv
}

func (serv *Server) Listens(addr string) error {
	if serv.Handler == nil {
		serv.Err = errors.New("client Handler is not nil")
		return serv.Err
	}
	serv.listener, serv.Err = quic.ListenAddr(addr, serv.tls, config)
	if serv.Err != nil {
		return serv.Err
	}
	for {
		session, err := serv.listener.Accept(context.Background())
		if err != nil {
			fmt.Println(err)
		} else {
			var client *Client
			remoteAddr := session.RemoteAddr().String()
			fmt.Println(remoteAddr, "connected")
			if v, ok := serv.Clients[remoteAddr]; ok {
				client = v
			} else {
				client = &Client{session: session,
					Addr:          remoteAddr,
					bufferPacket:  NewPacker(),
					Handler:       serv.Handler,
					ConnectStatus: true,
				}
				serv.Clients[remoteAddr] = client
			}
			client.MaxIdleTime = time.Now()
			//重置
			reloadFrequency()
			go serv.dealSession(client)
		}
	}
}

func (serv *Server) clearClient(client *Client, err error) {
	client.ConnectStatus = false
	serv.mu.Lock()
	if _, ok := serv.Clients[client.Addr]; ok {
		delete(serv.Clients, client.Addr)
		fmt.Println("cleared", client.Addr, err)
	}
	serv.mu.Unlock()
}

func (serv *Server) dealSession(client *Client) {
	var err error
	client.stream, err = client.session.AcceptStream(context.Background())
	if err != nil {
		serv.clearClient(client, err)
	} else {
		for {
			if !client.ConnectStatus {
				serv.clearClient(client, err)
				break
			}
			if err = client.Read(); err != nil {
				serv.clearClient(client, err)
				break
			}
		}
	}
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig(name ...string) *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   name,
	}
}
