package web

import (
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
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		router: NewRouter(),
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

	h.serve(ctx)
}

func (h *HTTPServer) serve(ctx *Context) {
	info, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || info.n.handler == nil {
		ctx.Resp.WriteHeader(404)
		_, _ = ctx.Resp.Write([]byte("NOT FOUND"))
		return
	}

	ctx.PathParams = info.pathParams
	info.n.handler(ctx)
}

// AddRoute 核心API

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

type HandleFunc func(ctx *Context)
