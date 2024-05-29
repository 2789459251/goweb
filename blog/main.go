package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
	"web/zygo"
	log_ "web/zygo/mylog"
	"web/zygo/mypool"
	err_ "web/zygo/zyerror"
)

type User struct {
	Name string   `json:"name" zygo:"required" xml:"name"`
	Age  int      `json:"age" validate:"min=18,max=20" xml:"age"`
	Addr []string `json:"addr" xml:"addr"`
}

func log(next zygo.HandlerFunc) zygo.HandlerFunc {
	return func(ctx *zygo.Context) {
		fmt.Println("打印请求函数")
		next(ctx)
		fmt.Println("返回执行时间")
	}
}

func main() {
	engine := zygo.Default()
	engine.RegisterErrorHandler(func(err error) (int, any) {
		switch e := err.(type) {
		case *BlogResponse:
			return http.StatusOK, e.Response()
		default:
			return http.StatusInternalServerError, "500 error"
		}
	})
	user := engine.Group("user")
	//user.Use(zygo.Logging)
	user.POST("/hello", func(ctx *zygo.Context) {
		//ctx.Logger.WithFields(mylog.Fields{
		//	"name": "码神之路",
		//	"id":   1000,
		//}).Debug("我是debug日志")
		//ctx.Logger.Info("我是info日志")
		//ctx.Logger.Error("我是error日志")
		//
		//ctx.JSON(http.StatusOK, user)
		fmt.Fprintln(ctx.W, "post hey bro!")
	})

	user.GET("/hello", func(ctx *zygo.Context) {
		fmt.Fprintln(ctx.W, "get hey bro!")
	})
	//user.Use(func(next zygo.HandlerFunc) zygo.HandlerFunc {
	//	return func(ctx *zygo.Context) {
	//		fmt.Println("pre handler")
	//		next(ctx)
	//		fmt.Println("post handler")
	//	}
	//})

	user.GET("/get/:id", func(ctx *zygo.Context) {
		fmt.Fprintln(ctx.W, "get id!")
	})
	user.GET("/hello/get", func(ctx *zygo.Context) {
		fmt.Fprintln(ctx.W, "go into *!")
	})
	user.GET("/html", func(ctx *zygo.Context) {
		ctx.HTML(http.StatusOK, "<h1>hello Zy</h1>")
	})
	//user.GET("/htmltemplate1", func(ctx *zygo.Context) {
	//	ctx.HTMLTemplate("index.html", "", "tem/index.html")
	//})
	user.GET("/htmltemplate2", func(ctx *zygo.Context) {
		u := &User{Name: "zyy"}
		err := ctx.HTMLTemplate("sc.html", u, "tem/sc.html", "tem/zy.html")
		if err != nil {
			fmt.Println(err)
		}
	})
	user.GET("/htmltemplate3", func(ctx *zygo.Context) {
		u := &User{Name: "zyy"}
		err := ctx.HTMLTemplateGlob("sc.html", u, "tem/*.html")
		if err != nil {
			fmt.Println(err)
		}
	})
	engine.LoadTemplate("tem/*.html")
	user.GET("/template", func(ctx *zygo.Context) {
		ctx.Template("sc.html", "")
	})
	user.GET("/json", func(ctx *zygo.Context) {
		u := &User{Name: "zzy"}
		err := ctx.JSON(http.StatusOK, u)
		if err != nil {
			fmt.Println(err)
		}
	})
	user.GET("/xml", func(ctx *zygo.Context) {
		u := &User{Name: "zzy"}
		err := ctx.XML(http.StatusOK, u)
		if err != nil {
			fmt.Println(err)
		}
	})
	user.GET("/excel", func(ctx *zygo.Context) {
		ctx.File("tem/test.txt")
	})
	user.GET("/excelName", func(ctx *zygo.Context) {
		ctx.FileAttachment("tem/test.txt", "aaaa.txt")
	})

	user.GET("/fs", func(ctx *zygo.Context) {
		ctx.FileFromFS("test.txt", http.Dir("tem"))
	})

	user.GET("/redirect", func(ctx *zygo.Context) {
		//状态会造成使用结果差别
		ctx.Redirect(http.StatusFound, "user/htmltemplate2")
	})

	user.GET("/string", func(ctx *zygo.Context) {
		//状态会造成使用结果差别
		ctx.String(http.StatusFound, "和好兄弟%s %s学习goweb框架", "zy", "sy")
	})
	user.GET("/add", func(ctx *zygo.Context) {
		//状态会造成使用结果差别
		id := ctx.GetDefaultQuery("name", "章三")
		fmt.Fprintf(ctx.W, "add name:%v \n", id)
	})

	user.GET("/user", func(ctx *zygo.Context) {
		//状态会造成使用结果差别
		id := ctx.QueryMap("user")
		ctx.JSON(http.StatusOK, id)
	})
	user.POST("/formPost", func(ctx *zygo.Context) {
		m, _ := ctx.GetPostFormMap("user")
		ctx.JSON(http.StatusOK, m)
	})
	user.POST("/file", func(ctx *zygo.Context) {
		m, _ := ctx.GetPostFormMap("user")
		file := ctx.FormFile("file")
		err := ctx.SaveUploadedFile(file, "./upload/"+file.Filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		form, err := ctx.MultipartForm()
		fmt.Println(err)
		dile := form.File
		d := dile["file"]
		for _, files := range d {
			err = ctx.SaveUploadedFile(file, "./upload/"+files.Filename)
		}
		fmt.Println(err)
		ctx.JSON(http.StatusOK, m)
	})
	user.POST("/files", func(ctx *zygo.Context) {
		m, _ := ctx.GetPostFormMap("user")
		form, err := ctx.MultipartForm()
		fmt.Println(err)
		dile := form.File
		d := dile["file"]
		for _, files := range d {
			err = ctx.SaveUploadedFile(files, "./upload/"+files.Filename)
		}
		fmt.Println(err)
		ctx.JSON(http.StatusOK, m)
	})

	user.POST("/jsonParam", func(ctx *zygo.Context) {
		u := &User{}

		ctx.DisallowUnknownFields = true
		ctx.IsValidate = true
		err := ctx.BindJson(u)
		if err == nil {
			ctx.JSON(http.StatusOK, u)
		} else {
			fmt.Println(err)
		}
	})
	//user.POST("/jsonParamSlice", func(ctx *zygo.Context) {
	//	u := make([]User, 0)
	//
	//	ctx.DisallowUnknownFields = true
	//	ctx.IsValidate = true
	//	err := ctx.Dealjson(&u)
	//	if err == nil {
	//		ctx.JSON(http.StatusOK, u)
	//	} else {
	//		fmt.Println(err)
	//	}
	//})
	//user.Use(zygo.Recovery)
	user.POST("/xmlParam", func(ctx *zygo.Context) {
		user := &User{}
		//u.Age = 10
		//user := &User{}
		//_ = ctx.BindXml(user)
		engine.Logger.Level = log_.LevelDebug
		//engine.Logger.Formatter = &log_.JsonFormatter{TimeDisplay: true}
		//logger.Outs = append(logger.Outs, &log_.LoggerWriter{
		//	Level: 2,
		//	Out:   log_.FileWriter("./log/log.log"),
		//})
		//engine.Logger.SetLogPath("./log")
		//engine.Logger.LogFileSize = 1 << 10 //1K
		//ctx.Logger.Debug("我是debug日志")
		//ctx.Logger.Info("我是info日志")
		//ctx.Logger.Error("我是error日志")
		//ctx.Logger.WithFields(log_.Fields{
		//	"name":    "zy",
		//	"emotion": "happy",
		//}).Error("这是字段测试")
		//fmt.Println(err)
		/* 统一触发recovery，处理错误*/
		var myerr *err_.MyError = err_.Default()
		myerr.Result(func(err *err_.MyError) {
			ctx.Logger.Info("我在统一解决问题,我不ok")
			ctx.JSON(http.StatusInternalServerError, myerr.Error())
		})
		a(1, myerr)
		b(1, myerr)
		c(1, myerr)
		ctx.JSON(http.StatusOK, user)
		err_ := login()
		ctx.HandleWithError(http.StatusOK, user, err_)
	})
	p, _ := mypool.NewPool(3)
	user.POST("/pool", func(ctx *zygo.Context) {
		currentTime := time.Now().UnixMilli()
		var wg sync.WaitGroup
		wg.Add(5)
		p.Submit(func() {
			defer func() {
				wg.Done()
			}()
			fmt.Println("1111111")
			//panic("这是1111的panic")
			time.Sleep(3 * time.Second)

		})
		p.Submit(func() {
			fmt.Println("22222222")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		p.Submit(func() {
			fmt.Println("33333333")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		p.Submit(func() {
			fmt.Println("44444")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		p.Submit(func() {
			fmt.Println("55555555")
			time.Sleep(3 * time.Second)
			wg.Done()
		})
		wg.Wait()
		fmt.Printf("time: %v \n", time.Now().UnixMilli()-currentTime)
		ctx.JSON(http.StatusOK, "success")
	})
	engine.Run(":8080", nil)
}

func a(int2 int, myError *err_.MyError) {
	if int2 == 1 {
		err := errors.New("a error")
		myError.Put(err)
	}
}
func b(int2 int, myError *err_.MyError) {
	if int2 == 1 {
		err := errors.New("b error")
		myError.Put(err)
	}
}

func c(int2 int, myError *err_.MyError) {
	if int2 == 1 {
		err := errors.New("c error")
		myError.Put(err)
	}
}

func login() *BlogResponse {
	return &BlogResponse{
		Success: false,
		Code:    99,
		Data:    nil,
		Msg:     "帐号密码错误，我在这里写了个blog响应错误，你看到了吗",
	}
}

type BlogResponse struct {
	Success bool
	Code    int
	Data    any
	Msg     string
}

type BlogNoataResponse struct {
	Success bool
	Code    int
	Msg     string
}

func (b *BlogResponse) Error() string {
	return b.Msg
}

func (b *BlogResponse) Response() any {
	if b.Data == nil {
		return &BlogNoataResponse{
			Success: b.Success,
			Code:    b.Code,
			Msg:     b.Msg,
		}
	}
	return b
}
