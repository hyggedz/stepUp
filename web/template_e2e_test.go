//go:build e2e

package web

import (
	"github.com/stretchr/testify/require"
	"html/template"
	"log"
	"testing"
)

func TestGoTemplateEngine_Render(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)

	engine := &GoTemplateEngine{T: tpl}
	server := NewHTTPServer(ServerWithTemplateEngine(engine))
	server.Get("/login", func(ctx *Context) {
		err := ctx.Render("login.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})
	server.Start(":8003")
}
