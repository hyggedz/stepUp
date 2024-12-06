package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type Context struct {
	Req        *http.Request
	Resp       http.ResponseWriter
	PathParams map[string]string

	//把查询参数缓存住
	//queryValues caches query parameters
	queryValues url.Values

	//命中路由
	MatchRoute string

	RespStatusCode int
	RespData       []byte
}

func (c *Context) BindJSON(val any) error {
	if c.Req.Body == nil {
		return errors.New("web:body 为 nil")
	}
	decoder := json.NewDecoder(c.Req.Body)
	return decoder.Decode(val)
}

func (c *Context) FormValue(key string) (string, error) {
	err := c.Req.ParseForm()
	if err != nil {
		return "", err
	}
	return c.Req.FormValue(key), nil
}

func (c *Context) QueryValue(key string) (string, error) {
	if c.queryValues == nil {
		c.queryValues = c.Req.URL.Query()
	}
	val, ok := c.queryValues[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return val[0], nil
}

func (c *Context) PathValue(key string) (string, error) {
	val, ok := c.PathParams[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return val, nil
}

func (c *Context) RespJSON(status int, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.Resp.WriteHeader(status)
	_, err = c.Resp.Write(data)
	return err
}

func (c *Context) RespJSONOK(val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	c.Resp.WriteHeader(http.StatusOK)
	n, err := c.Resp.Write(data)
	if n != len(data) {
		return errors.New("web: 数据未全部写出")
	}
	return err
}

func (c *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(c.Resp, ck)
}
