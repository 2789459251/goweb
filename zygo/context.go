package zygo

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"web/zygo/render"
)

type Context struct {
	W          http.ResponseWriter
	R          *http.Request
	engine     *Engine
	queryCache url.Values
}

func (c *Context) initQueryCache() {
	//if c.queryCache == nil {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
	//}
}

func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}

func (c *Context) QueryArray(key string) []string {
	c.initQueryCache()
	values, _ := c.queryCache[key]
	return values
}

func (c *Context) GetDefaultQuery(key string, defaultvalue string) string {
	values, ok := c.GetQueryArray(key)
	if !ok {
		return defaultvalue
	}
	return values[0]
}

func (c *Context) HTML(status int, html string) (err error) {
	//状态200
	c.Render(status, &render.HTML{Data: html, IsTemplate: false})
	return err
}
func (c *Context) HTMLTemplate(name string, data any, filenames ...string) (err error) {
	//状态200
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.New(name)
	t, err = t.ParseFiles(filenames...)
	if err != nil {
		return
	}
	err = t.Execute(c.W, data)
	return err
}

func (c *Context) get(cache map[string][]string, key string) (map[string]string, bool) {
	//user[id]=1&user[name]=zy
	dicts := make(map[string]string)
	exist := false
	for k, v := range cache {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dicts[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dicts, exist
}

func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

func (c *Context) QueryMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetQueryMap(key)
	return
}

func (c *Context) HTMLTemplateGlob(name string, data any, pattern string) (err error) {
	c.W.Header().Set("Content-type", "text/html;charset=utf-8")
	t := template.New(name)
	t, err = t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	t.Execute(c.W, data)
	return err
}

func (c *Context) Template(name string, data any) (err error) {
	c.Render(http.StatusOK, &render.HTML{Data: data, Name: name, IsTemplate: true, Templete: c.engine.HTMLRender.Template})
	return err
}

func (c *Context) JSON(status int, data any) (err error) {
	c.Render(status, &render.JSON{Data: data})
	return err
}

func (c *Context) XML(status int, data any) (err error) {
	c.Render(status, &render.Xml{Data: data})
	return err
}

func (c *Context) File(fileName string) {
	http.ServeFile(c.W, c.R, fileName)
}

// 指定下载名称
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment;filename="`+filename+`"`)
	} else {
		//这里拼接的怪怪的
		c.W.Header().Set("Content-Disposition", `attachment;filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		fmt.Println(old)
		c.R.URL.Path = old
	}(c.R.URL.Path) //	恢复

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

func (c *Context) Redirect(status int, url string) error {
	return c.Render(status, &render.Redirect{Code: status, Request: c.R, Location: url})
}

func (c *Context) String(status int, format string, values ...any) error {
	err := c.Render(status, &render.String{Format: format, Data: values})
	return err
}

func (c *Context) Render(status int, r render.Render) error {
	err := r.Render(c.W)
	c.W.WriteHeader(status)
	return err
}
