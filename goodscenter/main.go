package main

import (
	"goodscenter/api"
	"goodscenter/model"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"web/zygo"
)

func main() {
	r := zygo.Default()
	group := r.Group("goods")
	group.GET("/find", func(ctx *zygo.Context) {
		good := &model.Goods{
			ID:   1,
			Name: "跳跳糖",
		}
		ctx.JSON(http.StatusOK, &model.Result{
			Code: 200,
			Msg:  "succeess",
			Data: good,
		})
	})
	group.POST("/find", func(ctx *zygo.Context) {
		good := &model.Goods{
			ID:   1,
			Name: "跳跳糖",
		}
		ctx.JSON(http.StatusOK, &model.Result{
			Code: 200,
			Msg:  "succeess",
			Data: good,
		})
	})

	listen, _ := net.Listen("tcp", ":9111")
	server := grpc.NewServer()
	api.RegisterGoodsApiServer(server, &api.GoodsApiService{})
	err := server.Serve(listen)
	log.Println(err)
	r.Run(":9002", nil)
}
