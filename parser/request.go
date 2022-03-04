package parser

import (
	"ahs_server/module"
)

type request struct {
	*module.Request
}

func Request(data []byte) *module.Request {
	r := request{Request: new(module.Request)}
	r.Header = module.Header{}
	var (
		line int8
		step int8
		span bool
		body []byte
	)
	for i := 0; i < len(data); i++ {
		b := data[i]
		if span {
			body = append(body, b)
			continue
		}
		switch b {
		case ' ':
			if line == 0 {
				step++
				r.stepBody(line, step, body)
				body = body[:0]
				continue
			}
		case '\r':
			continue
		case '\n':
			if len(body) == 0 {
				span = true
				continue
			}
			step = 0
			r.stepBody(line, step, body)
			body = body[:0]
			line++
			continue
		}
		body = append(body, b)
	}
	if r.Request.Header.Method == "POST" {
		r.Request.Body = body
	}
	return r.Request
}

func (r request) stepBody(line, step int8, data []byte) {
	if len(data) == 0 {
		return
	}
	value := string(data)
	switch line {
	case 0:
		switch step {
		case 0:
			r.Header.Proto = value
		case 1:
			r.Header.Method = value
		case 2:
			r.Header.Path = value
		}
	default: //header
		r.Header.Headers = append(r.Header.Headers, value)
	}
}
