package transfer

import (
	"ahs_server/http"
	"ahs_server/module"
	"ahs_server/parser"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"log"
	"strconv"
	"strings"
	"time"
)

var apis = make(map[string]module.Service)

func Register(key string, api module.Api) {
	apis[key] = module.Service{Api: api}
}

//json
var json = jsoniter.ConfigCompatibleWithStandardLibrary

//web server
type web struct {
	host   module.Host
	server *module.HttpServer
}

//Init  local
func Init(port string) {
	conf, ok := localConfiguration(true)
	if !ok {
		log.Println("config error")
		return
	}
	config = conf

	for _, host := range config.Host {
		if host.Status != 1 {
			continue
		}
		w := &web{host: host}
		w.server = new(module.HttpServer)
		w.server.Port = strconv.Itoa(host.Port)
		w.server.Handler = w.httpHandler
		for _, lb := range host.Lbs {
			if err := w.startLB(lb); err != nil {
				log.Println("lb server error", err)
				return
			}
		}
		go func() {
			if err := http.Serve(w.server); err != nil {
				log.Println("start http server error", err)
			}
		}()
		go w.check()
	}
}

func (w *web) startLB(lb module.Lbs) error {
	if service, ok := apis[lb.Listen]; ok {
		var err error
		go func() {
			err = service.Api.Run(lb.Key, strconv.Itoa(lb.Port))
			if err != nil {
				log.Println("lb server error", err)
			}
		}()
		service.Api.Handler(w.lbHandler)
		//heartbeat
		if lb.Heartbeat != nil {
			if lb.Heartbeat.Content == "" {
				lb.Heartbeat.Content = "ok?"
			}
			if lb.Heartbeat.Interval <= 0 {
				lb.Heartbeat.Interval = 10
			}
			go service.Api.Heartbeat(lb.Heartbeat.Content, lb.Heartbeat.Interval)
		}
		if lb.IdleTime != nil {
			if lb.IdleTime.MaxTimeout <= 0 {
				lb.IdleTime.MaxTimeout = 10
			}
			go service.Api.CheckIdleTime(lb.IdleTime.MaxTimeout)
		}
		return err
	}
	return errors.New("lb server error")
}

func (w *web) lbHandler(conn module.HttpConn, data []byte) {
	var request *module.Request
	if err := json.Unmarshal(data, &request); err != nil {
		return
	}
	w.responseWrite(request)
}

func (w *web) httpHandler(http *module.HttpConn, data []byte) {
	http.ResponseStatus = 0
	request := parser.Request(data)
	request.Header.RemoteAddr = http.RemoteAddr
	if !w.routing(request) {
		w.responseError(request, "5004")
		return
	}
	var err error
	switch w.host.LbsMode {
	case 0:
		err = w.lbFor(request)
	default:
		w.responseError(request, "5007")
	}
	if err != nil {
		w.responseError(request, "5005")
	}
}

func Stop() {
	for _, host := range config.Host {
		log.Println(host)
	}
}

func (w *web) responseError(request *module.Request, code string) {
	request.Body = w.errorMsg(code)
	w.responseWrite(request)
}

func (w *web) responseWrite(request *module.Request) {
	conn, ok := w.get(request)
	if !ok {
		return
	}
	if conn.ResponseStatus > 0 {
		return
	}
	response := parser.CreateResponse(request.Header.Headers)
	response.ContentType = w.host.ContentType
	response.StatusCode = "200"
	if len(request.Body) == 0 {
		request.Body = w.errorMsg("5010")
	}
	response.Body = request.Body
	body := parser.ResponseBody(response)
	if err := conn.Handler.Write(body); err != nil {
		fmt.Println(err)
	}
	conn.ResponseStatus = 200
}

func (w *web) get(request *module.Request) (*module.HttpConn, bool) {
	v, ok := w.server.Requester.Load(request.Header.RemoteAddr)
	if !ok {
		return nil, false
	}
	return v.(*module.HttpConn), true
}

func (w *web) errorMsg(code string) []byte {
	var msg string
	switch code {
	case "5001":
		msg = "no query program"
	case "5002":
		msg = "content error"
	case "5003":
		msg = "send data error"
	case "5004":
		msg = "no access"
	case "5005":
		msg = "data was not delivered"
	case "5006":
		msg = "return format error"
	case "5007":
		msg = "not finished yet"
	case "5009":
		msg = "timeout"
	case "5010":
		msg = "empty"
	}
	var response string
	if w.host.ErrorResponse == "" {
		w.host.ContentType = "text/html; charset=utf-8"
		response = "error(" + code + ") " + msg
	} else {
		response = strings.Replace(w.host.ErrorResponse, "@error", msg, -1)
		response = strings.Replace(response, "@code", code, -1)
	}
	return []byte(response)
}

func (w *web) check() {
	var timeout float64 = 60
	if w.host.Timeout > 0 {
		timeout = float64(w.host.Timeout)
	}
	for {
		time.Sleep(time.Second * 5)
		w.server.Requester.Range(func(key, value interface{}) bool {
			http := value.(*module.HttpConn)
			//清除已关闭连接
			if http.Closed {
				w.server.Requester.Delete(key)
			} else if time.Now().Sub(http.ReactTime).Seconds() > timeout { //超过应答时间
				w.responseError(&module.Request{Header: module.Header{RemoteAddr: http.RemoteAddr}}, "5009")
			}
			return true
		})
	}
}
