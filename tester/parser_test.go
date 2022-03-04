package tester

import (
	"ahs_server/parser"
	"testing"
)

var body []byte

func init() {
	body = append(body, "POST /?act=1 HTTP/1.1"...)
	body = append(body, '\r', '\n')
	body = append(body, "Content-Type: application/json"...)
	body = append(body, '\r', '\n')
	body = append(body, "Accept: */*"...)
	body = append(body, '\r', '\n')
	body = append(body, "Accept-Encoding:  deflate, br"...)
	body = append(body, '\r', '\n')
	body = append(body, '\r', '\n')
	body = append(body, `{
   "requestHeader": {
       "action": "exchangeRate",
       "user": "ApiUser005",
       "password": "Lilxyg@mm%@142536"
   },
   "requestBody": {
       "BaseCode": "CNY"
   }
}`...)
}
func TestParser_Request(t *testing.T) {
	parser.Request(body)
}
func TestParser_Handler(t *testing.T) {

}
func Benchmark_Parser_Request(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			parser.Request(body)
		}
	})
}

func Benchmark_Parser_Handler(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

		}
	})
}
