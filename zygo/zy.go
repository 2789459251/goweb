package zygo

import (
	"html/template"
	"log"
	"net/http"
	"sync"
	"web/zygo/config"
	"web/zygo/mylog"
	"web/zygo/render"
)

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
	HEAD    = "HEAD"
	ANY     = "ANY"
)

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

type HandlerFunc func(ctx *Context)

type routerGroup struct {
	name               string
	handlerFuncMap     map[string]map[string]HandlerFunc
	middlewaresFuncMap map[string]map[string][]MiddlewareFunc
	handlerMethodMap   map[string][]string
	treeNode           *treeNode
	Middlewares        []MiddlewareFunc
}

func (r *router) Group(name string) *routerGroup {
	Group := &routerGroup{
		name:               name,
		handlerFuncMap:     make(map[string]map[string]HandlerFunc),
		handlerMethodMap:   make(map[string][]string),
		middlewaresFuncMap: make(map[string]map[string][]MiddlewareFunc),
		treeNode:           &treeNode{name: "/", children: make([]*treeNode, 0)},
	}
	Group.Use(r.engine.middles...)
	r.routerGroups = append(r.routerGroups, Group)
	return Group
}

func (r *routerGroup) Use(middlewareFunc ...MiddlewareFunc) {
	r.Middlewares = append(r.Middlewares, middlewareFunc...)
}

func (r *routerGroup) methodHandle(routerName, method string, h HandlerFunc, ctx *Context) {
	//前置
	if r.Middlewares != nil {
		for _, middlewarefunc := range r.Middlewares {
			h = middlewarefunc(h)
		}
	}
	funcMidds := r.middlewaresFuncMap[routerName][method]
	//路由级别中间件
	if funcMidds != nil {
		for _, mid := range funcMidds {
			h = mid(h)
		}
	}
	h(ctx)

}

type router struct {
	routerGroups []*routerGroup
	engine       *Engine
}

//name:url method:请求方式
func (r *routerGroup) handle(name string, method string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	if _, ok := r.handlerFuncMap[name]; !ok {
		r.handlerFuncMap[name] = make(map[string]HandlerFunc)
		r.middlewaresFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	if _, ok := r.handlerFuncMap[name][method]; ok {
		panic("reset handlerFunc")
	}
	r.handlerFuncMap[name][method] = handlerFunc
	r.middlewaresFuncMap[name][method] = append(r.middlewaresFuncMap[name][method], middlewareFunc...)
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], name)
	r.treeNode.Put(name)
}

// functionMap[url][method]---function
func (r *routerGroup) GET(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, GET, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) POST(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, POST, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) PUT(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, PUT, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) DELETE(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, DELETE, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) PATCH(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, PATCH, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) OPTIONS(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, OPTIONS, handlerFunc, middlewareFunc...)
}

func (r *routerGroup) HEAD(name string, handlerFunc HandlerFunc, middlewareFunc ...MiddlewareFunc) {
	r.handle(name, HEAD, handlerFunc, middlewareFunc...)
}

type ErrorHandler func(err error) (int, any)

type Engine struct {
	router
	funcMap      template.FuncMap
	HTMLRender   render.HTMLRender
	pool         sync.Pool
	Logger       *mylog.Logger
	middles      []MiddlewareFunc
	errorHandler ErrorHandler
}

func Default() *Engine {
	engine := New()
	engine.Use(Logging, Recovery)
	engine.router.engine = engine
	engine.Logger = mylog.Default()
	logPath, ok := config.Conf.Log["path"]
	if ok {
		engine.Logger.SetLogPath(logPath.(string))
	}
	return engine
}

func (e *Engine) RegisterErrorHandler(err ErrorHandler) {
	e.errorHandler = err
}

func (e *Engine) Use(middles ...MiddlewareFunc) {
	e.middles = append(e.middles, middles...)
}

func New() *Engine {
	engine := &Engine{
		router: router{},
	}
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine
}

func (e *Engine) allocateContext() any {
	return &Context{engine: e}
}
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadTemplateConf() {
	//这里没有name阿
	pattern, ok := config.Conf.Template["pattern"]
	if ok {
		t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern.(string)))
		e.SetHtmlTemplate(t)
	}

}

func (e *Engine) LoadTemplate(pattern string) {
	//这里没有name阿
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetHtmlTemplate(t)
}

func (e *Engine) SetHtmlTemplate(t *template.Template) {
	e.HTMLRender = render.HTMLRender{Template: t}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := e.pool.Get().(*Context)
	ctx.W = w
	ctx.R = r
	ctx.Logger = e.Logger
	e.httpRequestHandle(ctx, w, r)
	e.pool.Put(ctx)
}
func (e *Engine) httpRequestHandle(ctx *Context, w http.ResponseWriter, r *http.Request) {
	method := r.Method

	for _, group := range e.routerGroups {
		routerName := SubStringLast(r.URL.Path, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node != nil && node.isEnd {
			handle, ok := group.handlerFuncMap[node.routerName][method]
			if ok {

				group.methodHandle(node.routerName, method, handle, ctx)

				return
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte("router exists but no same method"))
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("router not exists"))
		return
	}
}
func (e *Engine) Run(port string, handler http.Handler) (err error) {

	/*	for _, group := range e.routerGroups {
			for key, value := range group.handlerFuncMap {
				http.HandleFunc("/"+group.name+key, value)
			}
		}
	*/
	http.Handle("/", e)
	err = http.ListenAndServe(port, handler)
	return
}

func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	err := http.ListenAndServeTLS(addr, certFile, keyFile, e.Handler())
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) Handler() http.Handler {
	return e
}
