package parser

import (
	"ahs_server/toolkit"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	StatusCode     string
	AcceptEncoding string
	ContentType    string
	Headers        []string
	Body           []byte
}

// ResponseBody
//1xx：指示信息--表示请求已接收，继续处理
//2xx：成功--表示请求已被成功接收、理解、接受
//3xx：重定向--要完成请求必须进行更进一步的操作
//4xx：客户端错误--请求有语法错误或请求无法实现
//5xx：服务器端错误--服务器未能实现合法的请求
func ResponseBody(response *Response) []byte {
	var b []byte
	b = append(b, "HTTP/1.1"...)
	b = append(b, ' ')
	b = append(b, response.StatusCode...)
	b = append(b, ' ')
	b = append(b, "OK"...)
	b = append(b, '\r', '\n')
	b = append(b, "Server: AgileHttpServer 0.01\r\n"...)
	b = append(b, "Date: "...)
	b = time.Now().AppendFormat(b, "Mon, 02 Jan 2006 15:04:05 GMT")
	b = append(b, '\r', '\n')
	l := int64(len(response.Body))
	b = append(b, "Content-Type: "...)
	if response.AcceptEncoding == "gzip" && l > 1024*2 {
		response.Body = toolkit.GzipCompressBytes(response.Body)
		response.Headers = append(response.Headers, "Content-Encoding: gzip")
		l = int64(len(response.Body))
	}
	b = append(b, response.ContentType...)
	b = append(b, '\r', '\n')
	b = append(b, "Content-Length: "...)
	b = strconv.AppendInt(b, l, 10)
	b = append(b, '\r', '\n')

	for _, head := range response.Headers {
		b = append(b, head...)
		b = append(b, '\r', '\n')
	}
	b = append(b, '\r', '\n')
	b = append(b, response.Body...)
	return b
}

func CreateResponse(header []string) *Response {
	rh := &Response{}
	for _, s := range header {
		if strings.Index(s, "Accept-Encoding") > -1 && strings.LastIndex(s, "gzip") > -1 {
			rh.AcceptEncoding = "gzip"
		}
		if strings.Index(s, "Content-Type") > -1 {
			rh.ContentType = s
		}
	}
	return rh
}
