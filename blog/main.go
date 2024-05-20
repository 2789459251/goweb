package main

import (
	"fmt"
	"net/http"
	"web/zygo"
	log_ "web/zygo/mylog"
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
	user.Use(func(next zygo.HandlerFunc) zygo.HandlerFunc {
		return func(ctx *zygo.Context) {
			fmt.Println("pre handler")
			next(ctx)
			fmt.Println("post handler")
		}
	})

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
	var u *User
	//user.Use(zygo.Recovery)
	user.POST("/xmlParam", func(ctx *zygo.Context) {
		u.Age = 10
		user := &User{}
		err := ctx.BindXml(user)

		engine.Logger.Level = log_.LevelDebug
		//engine.Logger.Formatter = &log_.JsonFormatter{TimeDisplay: true}
		//logger.Outs = append(logger.Outs, &log_.LoggerWriter{
		//	Level: 2,
		//	Out:   log_.FileWriter("./log/log.log"),
		//})
		engine.Logger.SetLogPath("./log")
		engine.Logger.LogFileSize = 1 << 10 //1K
		ctx.Logger.Debug("我是debug日志")
		ctx.Logger.Info("我是info日志")
		ctx.Logger.Error("我是error日志")
		ctx.Logger.WithFields(log_.Fields{
			"name":    "zy",
			"emotion": "happy",
		}).Error("这是字段测试")
		fmt.Println(err)
		ctx.JSON(http.StatusOK, user)
	})
	engine.Run(":8080", nil)
}
