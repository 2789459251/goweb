package zygo

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"web/zygo/config"
	"web/zygo/gateway"
	"web/zygo/mylog"
	"web/zygo/register"
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
	OpenGateway  bool
	//维护网关的路径节点
	gatewayTreeNode  *gateway.TreeNode
	gatewayConfigMap map[string]gateway.GWConfig
	gatewayConfigs   []gateway.GWConfig
	RegisterType     string
	RegisterOption   register.Option
	RegisterCli      register.MyRegister
}

func (e *Engine) SetGateConfigs(configs []gateway.GWConfig) {
	e.gatewayConfigs = configs
	//存储路径 如果符合就匹配
	for _, v := range e.gatewayConfigs {
		e.gatewayTreeNode.Put(v.Path, v.Name)
		e.gatewayConfigMap[v.Name] = v
	}
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
		router:           router{},
		gatewayTreeNode:  &gateway.TreeNode{Name: "/", Children: make([]*gateway.TreeNode, 0)},
		gatewayConfigMap: make(map[string]gateway.GWConfig, 0),
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

/*加入网关时逻辑修改 找到目标 对应替换*/
func (e *Engine) httpRequestHandle(ctx *Context, w http.ResponseWriter, r *http.Request) {
	if e.OpenGateway {
		//req -> 网关 -> 配置分发
		path := r.URL.Path
		node := e.gatewayTreeNode.Get(path)
		if node == nil {
			ctx.W.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(ctx.W, ctx.R.RequestURI+" not found")
			return
		}
		gwConfig := e.gatewayConfigMap[node.GwName]
		gwConfig.Header(ctx.R)

		addr, err2 := e.RegisterCli.GetValue(gwConfig.ServiceName)

		if err2 != nil {
			ctx.W.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(ctx.W, ctx.R.RequestURI+" register cli get value err:"+err2.Error())
			return
		}
		target, err := url.Parse(fmt.Sprintf("http://%s%s", addr, path))

		if err != nil {
			ctx.W.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(ctx.W, err.Error())
		}
		//网关处理逻辑
		director := func(req *http.Request) {
			req.Host = target.Host
			req.URL.Host = target.Host
			req.URL.Path = target.Path
			req.URL.Scheme = target.Scheme
			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "")
			}
		}
		//todo 代理的响应处理和错误修改
		response := func(response *http.Response) error {
			log.Println("响应修改")
			return nil
		}
		handler := func(writer http.ResponseWriter, request *http.Request, err error) {
			log.Println("错误处理：" + err.Error())
		}
		proxy := httputil.ReverseProxy{Director: director, ModifyResponse: response, ErrorHandler: handler}
		//代理帮我转发
		proxy.ServeHTTP(ctx.W, ctx.R)
		return
	}

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
func (e *Engine) Run(addr string) {

	/*	for _, group := range e.routerGroups {
			for key, value := range group.handlerFuncMap {
				http.HandleFunc("/"+group.name+key, value)
			}
		}
	*/
	if e.RegisterType == "nacos" {
		r := &register.NacosRegister{}
		err := r.CreateCli(e.RegisterOption)
		if err != nil {
			panic(err)
		}
		e.RegisterCli = r
	}
	if e.RegisterType == "etcd" {
		r := &register.EtcdRegister{}
		err := r.CreateCli(e.RegisterOption)
		if err != nil {
			panic(err)
		}
		e.RegisterCli = r
	}
	http.Handle("/", e)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
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
