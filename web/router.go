package web

import (
	"fmt"
	"strings"
)

type router struct {
	//路由数定义在这上面
	trees map[string]*node
}

type node struct {
	//当前路径
	path string

	//子节点path => 子节点
	children map[string]*node

	handler HandleFunc
}

func NewRouter() router {
	r := router{
		trees: map[string]*node{},
	}
	return r
}

// addRoute 注册路由
// method HTTP方法
// path URL路径
// handler 业务逻辑
func (r *router) addRoute(method string, path string, handler HandleFunc) {
	//method: http.MethodGet,
	//path:   "/user/home",
	root, ok := r.trees[method]
	if !ok {
		root = &node{path: "/"}
		r.trees[method] = root
	}

	if path[0] != '/' {
		panic("路径必须以 / 开头")
	}

	if path == "/" {
		if root.handler != nil {
			panic(fmt.Sprintf("路由冲突，重复注册%s", path))
		}
		root.handler = handler
		return
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("path 不能以 / 结尾")
	}

	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == "" {
			panic("不能出现多个 / ")
		}
		root = root.childOrCreate(seg)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("路由冲突，重复注册%s", path))
	}
	root.handler = handler
}

func (n *node) childOrCreate(path string) *node {
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{path: path}
		n.children[path] = child
	}
	return child
}

// findRoute 查找路由
func (r *router) findRoute(method string, path string) (*node, bool) {
	//也是按照深度查找
	root, ok := r.trees[method]
	if !ok {
		//不存在这个http 方法
		return nil, false
	}

	if path == "/" {
		return root, true
	}

	//去掉前后缀的 / then 切割
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segs {
		child, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		root = child
	}
	return root, true
}

func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return nil, false
	}

	child, ok := n.children[path]
	return child, ok
}
