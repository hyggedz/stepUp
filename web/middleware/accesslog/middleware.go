package accesslog

import (
	"encoding/json"
	"log"
	"stepUp_go/web"
)

type MiddleWareBuilder struct {
	logFunc func(accessLog string)
}

func NewMiddleWareBuilder() *MiddleWareBuilder {
	return &MiddleWareBuilder{
		logFunc: func(accessLog string) {
			log.Println(accessLog)
		},
	}
}

func (m *MiddleWareBuilder) LogFunc(logFunc func(accessLog string)) *MiddleWareBuilder {
	m.logFunc = logFunc
	return m
}

func (m MiddleWareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			//就算panic也会log
			defer func() {
				l := accessLog{
					Host:       ctx.Req.Host,
					Route:      ctx.MatchRoute,
					HTTPMethod: ctx.Req.Method,
					Path:       ctx.Req.URL.Path,
				}
				val, _ := json.Marshal(l)
				m.logFunc(string(val))
			}()
			next(ctx)
		}
	}
}

type accessLog struct {
	Host       string
	Route      string
	HTTPMethod string `json:"http_method"`
	Path       string
}
