//go:build e2e

package prometheus

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	web "stepUp_go/web"
	"testing"
	"time"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := MiddleWareBuilder{
		Namespace: "geekbang",
		Subsystem: "web",
		Name:      "http_response",
	}

	server := web.NewHTTPServer(web.ServerWithMiddleware(builder.Build()))

	server.Get("/user", func(ctx *web.Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Duration(val) * time.Millisecond)
		ctx.RespJSON(200, User{Name: "Tom"})
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()

	server.Start(":8081")
}

type User struct {
	Name string
}