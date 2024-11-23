//go:bulid e2e
package web

import "testing"

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/order/detail", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, this is order detail"))
	})
	s.Start(":8083")
}
