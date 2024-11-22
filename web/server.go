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
	AddRoute(method string, path string, handler HandleFunc)

	//控制生命周期
	Start(addr string) error
}

type HTTPServer struct{}

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

// serve 注册路由
func (h *HTTPServer) serve(ctx *Context) {

}

// AddRoute 核心API
func (h *HTTPServer) AddRoute(method string, path string, handler HandleFunc) {
	panic("implement me")
}

func (h *HTTPServer) Get(path string, handler HandleFunc) {
	h.AddRoute(http.MethodGet, path, handler)
}

func (h *HTTPServer) Post(path string, handler HandleFunc) {
	h.AddRoute(http.MethodPost, path, handler)
}

func (h *HTTPServer) Delete(path string, handler HandleFunc) {
	h.AddRoute(http.MethodDelete, path, handler)
}
func (h *HTTPServer) Options(path string, handler HandleFunc) {
	h.AddRoute(http.MethodOptions, path, handler)
}

type HandleFunc func(ctx *Context)
