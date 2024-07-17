package zygo

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"web/zygo/binding"
	"web/zygo/mylog"
	"web/zygo/render"
)

var defaultMaxMemory int64 = 32 << 20          //32兆
var defaultMultipartMaxMemory int64 = 32 << 20 //32兆

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	engine                *Engine
	queryCache            url.Values
	formCache             url.Values
	DisallowUnknownFields bool
	IsValidate            bool
	StatusCode            int
	Logger                *mylog.Logger
	Keys                  map[string]any
	mu                    sync.RWMutex
	//安全性操作
	sameSite http.SameSite
}

func (c *Context) SetSameSite(s http.SameSite) {
	c.sameSite = s
}

// 保存cookie
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.W, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
	c.mu.Unlock()
}

func (c *Context) Get(key string) (v any, ok bool) {
	c.mu.RLock()
	v, ok = c.Keys[key]
	c.mu.RUnlock()
	return
}
func (c *Context) GetHeader(key string) string {
	return c.R.Header.Get(key)
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
	//如果设置了code对Header修改不生效
	//c.W.WriteHeader(status)
	err := r.Render(c.W, status)
	c.StatusCode = status
	return err
}

func (c *Context) initPostFormCache() {
	if c.R != nil {
		if err := c.R.ParseMultipartForm(defaultMaxMemory); err != nil { //支持文件，最大存
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Panicln(err)
			}
			//不带文件,也会解析错误，需要忽略
		}
		c.formCache = c.R.PostForm
	} else {
		c.formCache = url.Values{}
	}
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initPostFormCache()
	values, ok := c.queryCache[key]
	return values, ok
}

func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initPostFormCache()
	return c.get(c.formCache, key)
}

func (c *Context) PostFormMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetPostFormMap(key)
	return
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArray(key string) (values []string) {
	values, _ = c.GetPostFormArray(key)
	return
}

func (c *Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Panicln(err)
	}
	defer file.Close()
	return header
}
func (c *Context) FormFileArray(name string) []*multipart.FileHeader {
	form, err := c.MultipartForm()
	if err != nil {
		log.Panicln(err)
		return nil
	}
	return form.File[name]
}

// dst表示目标路径
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dist, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dist.Close()
	_, err = io.Copy(dist, src)
	return err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.R.ParseMultipartForm(defaultMultipartMaxMemory)
	return c.R.MultipartForm, err
}

// 解析json
// func (c *Context) Dealjson(obj any) error {
//
// }
func (c *Context) BindJson(obj any) error {
	json := binding.JSON
	json.DisallowUnknownFields = c.DisallowUnknownFields
	json.IsValidate = c.IsValidate
	return c.MustBindWith(obj, &binding.JSON)
}

func (c *Context) BindXml(obj any) error {
	xml := binding.XML

	return c.MustBindWith(obj, &xml)
}

func (c *Context) MustBindWith(obj any, binding binding.Binding) error {
	if err := c.ShouldBind(obj, binding); err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

func (c *Context) ShouldBind(obj any, bind binding.Binding) error {
	return bind.Bind(c.R, obj)
}

func (c *Context) Fail(code int, msg string) {
	c.String(code, msg)
}

func (c *Context) HandleWithError(statuscode int, obj any, err error) {
	if err != nil {
		code, data := c.engine.errorHandler(err)
		c.JSON(code, data)
		return
	}
	c.JSON(statuscode, obj)
}

//func validateStruct(obj any) error {
//	return validator_.ValidateStruct(obj)
//}

//func (err SliceValidationError) Error() string {
//	n := len(err)
//	switch n {
//	case 0:
//		return ""
//	default:
//		var b strings.Builder
//		if err[0] != nil {
//			fmt.Fprintf(&b, "[%d]:%s,", 0, err[0].Error())
//		}
//		if n > 1 {
//			for i := 1; i < n; i++ {
//				if err[i] != nil {
//					b.WriteString("\n")
//					fmt.Fprintf(&b, "[%d]:%s,", i, err[i].Error())
//				}
//			}
//		}
//		return b.String()
//	}
//
//}
