package main

import (
	"fmt"
	"net/http"
	"web/zygo"
)

type User struct {
	Name string
}

func log(next zygo.HandlerFunc) zygo.HandlerFunc {
	return func(ctx *zygo.Context) {
		fmt.Println("打印请求函数")
		next(ctx)
		fmt.Println("返回执行时间")
	}
}

func main() {
	engine := zygo.New()
	user := engine.Group("user")
	user.POST("/hello", func(ctx *zygo.Context) {
		fmt.Fprintln(ctx.W, "post hey bro!")
	}, log)

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
	engine.Run(":8080", nil)
}
