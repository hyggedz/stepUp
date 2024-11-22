package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_AddRoute(t *testing.T) {
	testRoute := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	r := NewRouter()
	var mockHandler HandleFunc = func(ctx *Context) {}
	for _, tr := range testRoute {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRoute := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
			},
			http.MethodPost: &node{
				path: "/",
				children: map[string]*node{
					"login": &node{
						path:    "login",
						handler: mockHandler,
					},
				},
			},
		},
	}

	//在这里断言
	//因为node包含handler不能直接使用Equal
	msg, ok := wantRoute.equal(r)
	assert.True(t, ok, msg)

	r = NewRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "a", mockHandler)
	}, "web:path必须以 / 开头")

	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a/", mockHandler)
	}, "web:path不能以 / 结尾")

	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	}, "web:不能出现多个 / ")

	r = NewRouter()
	r.addRoute(http.MethodGet, "/user/home", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/user/home", mockHandler)
	}, "路由冲突，重复注册[/user/home]")

	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	}, "路由冲突，重复注册[/]")
}

func (r router) equal(y router) (string, bool) {
	//总览
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标router中不存在%s方法", k), false
		}
		//每一颗树是不是相等
		msg, ok := v.equal(yv)
		if !ok {
			return msg, ok
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if n.path != y.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不匹配"), false
	}

	//对比handler
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)

	if nHandler != yHandler {
		return fmt.Sprintf("函数不相等"), false
	}

	for path, c := range n.children {
		yv, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在", path), false
		}

		msg, ok := c.equal(yv)
		if !ok {
			return msg, ok
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodDelete,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	r := NewRouter()
	var mockHandler HandleFunc = func(ctx *Context) {}
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	testCases := []struct {
		name string

		method string
		path   string

		wantFound bool
		wantNode  *node
	}{
		{
			name:   "method not found",
			method: http.MethodOptions,
			path:   "/order/detail",

			wantFound: false,
			wantNode:  nil,
		},
		{
			name:   "完全命中",
			method: http.MethodGet,
			path:   "/order/detail",

			wantNode: &node{
				path:    "detail",
				handler: mockHandler,
			},
			wantFound: true,
		},
		{
			name:   "order",
			method: http.MethodGet,
			path:   "/order",

			wantFound: true,
			wantNode: &node{
				path: "order",
				children: map[string]*node{
					"detail": &node{
						path:    "detail",
						handler: mockHandler,
					},
				},
			},
		},
		{
			name: "path not found",

			method: http.MethodGet,
			path:   "/abbbbbc",
		},
		{
			name: "root",

			method: http.MethodDelete,
			path:   "/",

			wantFound: true,
			wantNode: &node{
				path:    "/",
				handler: mockHandler,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)

			if !found {
				return
			}

			msg, ok := n.equal(tc.wantNode)
			assert.True(t, ok, msg)
		})
	}

}
