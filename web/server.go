package web

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

var _ Server = &HTTPServer{}

// 服务器抽象
// ServeHTTP方法是http包和web框架的接入点
// 我们需要组合handler
// 因为handler是带有ServeHTTP方法的接口
type Server interface {
	http.Handler

	//注册路由
	addRoute(method string, path string, handler HandleFunc)

	//控制生命周期
	Start(addr string) error
}

type HTTPServer struct {
	router
	mdl []Middleware

	log func(msg string, args ...any)
}

type HTTPServerOption func(server *HTTPServer)

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: NewRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.mdl = mdls
	}
}

func (h *HTTPServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return http.Serve(l, h)
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//注册上下文
	ctx := &Context{
		Req:  r,
		Resp: w,
	}
	//注册路由

	root := h.serve
	//AOP
	for i := len(h.mdl) - 1; i >= 0; i-- {
		root = h.mdl[i](root)
	}

	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.n.handler == nil {
		ctx.Resp.WriteHeader(404)
		_, _ = ctx.Resp.Write([]byte("NOT FOUND"))
		return
	}

	ctx.PathParams = info.pathParams
	ctx.MatchRoute = info.n.route
	info.n.handler(ctx)
}

func (h *HTTPServer) Use(mdl ...Middleware) {
	if h.mdl == nil {
		h.mdl = mdl
		return
	}
	h.mdl = append(h.mdl, mdl...)
}

func (h *HTTPServer) Get(path string, handler HandleFunc) {
	h.addRoute(http.MethodGet, path, handler)
}

func (h *HTTPServer) Post(path string, handler HandleFunc) {
	h.addRoute(http.MethodPost, path, handler)
}

func (h *HTTPServer) Delete(path string, handler HandleFunc) {
	h.addRoute(http.MethodDelete, path, handler)
}
func (h *HTTPServer) Options(path string, handler HandleFunc) {
	h.addRoute(http.MethodOptions, path, handler)
}

func (s *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode > 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	_, err := ctx.Resp.Write(ctx.RespData)
	if err != nil {
		log.Fatalln("回写响应失败", err)
	}
}

type HandleFunc func(ctx *Context)
