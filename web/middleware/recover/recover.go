package recover

import "stepUp_go/web"

type MiddleWareBuilder struct {
	StatusCode int
	RespData   string
	logFunc    func(ctx *web.Context)
}

func (m MiddleWareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespStatusCode = m.StatusCode
					ctx.RespData = []byte(m.RespData)
					m.logFunc(ctx)
				}
			}()
			next(ctx)
		}
	}
}
