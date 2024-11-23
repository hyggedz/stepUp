package web

import (
	"fmt"
	"strings"
)

type router struct {
	//路由数定义在这上面
	trees map[string]*node
}

// 命中路由的规则
// 1.优先匹配静态路由
// 2.再考虑通配符路由
// 这是不回溯的
type node struct {
	//当前路径
	path string

	//子节点path => 子节点
	children map[string]*node

	handler HandleFunc

	//通配符路由
	starChild *node

	//参数路由匹配
	paramChild *node
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
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
	//通配符路由
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}

		if n.starChild == nil {
			n.starChild = &node{path: "*"}
		}
		return n.starChild
	}

	//参数路由
	if path[0] == ':' {
		if n.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}

		if n.paramChild != nil {
			if n.paramChild.path != path {
				panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
			}
		} else {
			n.paramChild = &node{
				path: path,
			}
		}
		return n.paramChild
	}

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
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	//也是按照深度查找
	root, ok := r.trees[method]
	if !ok {
		//不存在这个http 方法
		return nil, false
	}

	if path == "/" {
		return &matchInfo{
			n: root,
		}, true
	}

	//去掉前后缀的 / then 切割
	segs := strings.Split(strings.Trim(path, "/"), "/")
	info := &matchInfo{}
	for _, seg := range segs {
		child, pathParam, found := root.childOf(seg)
		if pathParam {
			if info.pathParams == nil {
				info.pathParams = map[string]string{child.path[1:]: seg}
			}
			info.pathParams[child.path[1:]] = seg
		}

		if !found {
			return nil, false
		}
		root = child
	}
	return &matchInfo{
		n:          root,
		pathParams: info.pathParams,
	}, true
}

// childOf
// 第一个是路径参数加节点
// 第二个是是否命中路径参数
// 第三个是是否匹配
func (n *node) childOf(path string) (*node, bool, bool) {
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}

	child, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return child, false, ok
}
