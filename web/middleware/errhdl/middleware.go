package errhdl

import "stepUp_go/web"

type MiddleWareBuilder struct {
	resp map[int][]byte
}

func NewMiddleWareBuilder() *MiddleWareBuilder {
	return &MiddleWareBuilder{
		resp: make(map[int][]byte, 64),
	}
}

func (m *MiddleWareBuilder) RegisterErr(code int, data []byte) *MiddleWareBuilder {
	m.resp[code] = data
	return m
}

func (m MiddleWareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			next(ctx)
			resp, ok := m.resp[ctx.RespStatusCode]
			if ok {
				m.resp[ctx.RespStatusCode] = resp
			}
		}
	}
}
