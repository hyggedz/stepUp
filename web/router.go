package web

import (
	"fmt"
	"regexp"
	"strings"
)

type router struct {
	//路由数定义在这上面
	trees map[string]*node
}

type nodeType int

const (
	//静态路由
	nodeTypeStatic = iota
	//正则表达式路由
	nodeTypeReg
	//参数路由
	nodeTypeParam
	//通配符路由
	nodeTypeAny
)

// 命中路由的规则
// 1.优先匹配静态路由
// 2.再考虑通配符路由
// 这是不回溯的
type node struct {
	nTyp nodeType
	//当前路径
	path string

	//到当前节点的路由
	route string

	//子节点path => 子节点
	children map[string]*node

	handler HandleFunc

	//通配符路由
	starChild *node

	//参数路由匹配
	paramChild *node
	//参数名称
	//正则匹配和参数匹配都用到
	paramName string
	//正则表达式匹配
	regChild *node
	regExpr  *regexp.Regexp
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
		root = &node{path: "/", nTyp: nodeTypeStatic}
		r.trees[method] = root
	}
	if path == "" {
		panic("web: 路由是空字符串")
	}

	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	if path == "/" {
		if root.handler != nil {
			panic(fmt.Sprintf("web: 路由冲突[/]"))
		}
		root.handler = handler
		return
	}

	segs := strings.Split(path[1:], "/")
	for _, seg := range segs {
		if seg == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.childOrCreate(seg)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handler
	root.route = path
}

func (n *node) childOrCreate(path string) *node {
	//通配符路由
	if path == "*" {
		if n.paramChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册通配符路由和参数路由 [%s]", path))
		}

		if n.regChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册通配符路由和正则路由 [%s]", path))
		}

		if n.starChild == nil {
			n.starChild = &node{path: "*", nTyp: nodeTypeAny}
		}
		return n.starChild
	}

	//参数路由 正则表达式路由
	if path[0] == ':' {
		var child *node
		//解析路径 判断是哪一种路由
		pN, regExpr, isReg := n.prasePath(path)
		if isReg {
			child = n.childOrCreateReg(path, pN, regExpr)
		} else {
			child = n.childOrCreateParam(pN)
		}
		return child
	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{path: path, nTyp: nodeTypeStatic}
		n.children[path] = child
	}
	return child
}

func (n *node) childOrCreateParam(path string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
	}
	if n.regChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.paramChild != nil {
		if n.paramChild.path != path {
			panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.paramChild.path, path))
		}
	} else {
		n.paramChild = &node{
			path:      path,
			paramName: path[1:],
			nTyp:      nodeTypeParam,
		}
	}
	return n.paramChild
}

func (n *node) childOrCreateReg(path string, pN string, regExpr string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.paramChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}

	if n.regChild != nil {
		if n.regChild.regExpr.String() != regExpr || n.regChild.paramName != pN {
			panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
		}
	} else {
		regx, err := regexp.Compile(regExpr)
		if err != nil {
			panic(fmt.Errorf("web: 正则表达式错误 %w", err))
		}

		n.regChild = &node{
			path:      path,
			paramName: pN,
			regExpr:   regx,
			nTyp:      nodeTypeReg,
		}
	}
	return n.regChild
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
				info.pathParams = map[string]string{child.paramName: seg}
			}
			info.pathParams[child.paramName] = seg
		}

		if !found {
			//通配符在最后
			if root.nTyp == nodeTypeAny {
				return &matchInfo{n: root}, true
			}
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
	//user/*
	//user/abc/d
	if n.children == nil {
		return n.childOfNonStatic(path)
	}

	child, ok := n.children[path]
	if !ok {
		return n.childOfNonStatic(path)
	}
	return child, false, ok
}

func (n *node) childOfNonStatic(path string) (*node, bool, bool) {
	if n.regChild != nil && n.regChild.regExpr.MatchString(path) {
		return n.regChild, true, true
	}
	if n.paramChild != nil {
		return n.paramChild, true, true
	}
	return n.starChild, false, n.starChild != nil
}

// 第一个是paramName
// 第二个是正则表达式
// 第三个是是否是正则路由
func (n *node) prasePath(path string) (string, string, bool) {

	// /:id()
	segs := strings.Split(path, "(")
	// :id ..)
	if len(segs) == 2 {
		// id
		pN := segs[0][1:]
		reg := segs[1][:len(segs[1])-1]
		return pN, reg, true
	}
	//否则，返回参数名称就可以
	return path, "", false
}
