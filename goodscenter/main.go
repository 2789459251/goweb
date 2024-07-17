package main

import (
	"encoding/gob"
	"goodscenter/model"
	"goodscenter/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"web/zygo"
	"web/zygo/rpc"
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

	//1.grpc
	//listen, _ := net.Listen("tcp", ":9111")
	//server := grpc.NewServer()
	//api.RegisterGoodsApiServer(server, &api.GoodsApiService{})
	//err := server.Serve(listen)
	//log.Println(err)

	//2.框架封装grpc
	//server, err := rpc.NewGrpcServer(":9111")
	//if err != nil {
	//	return
	//}
	//server.Register(func(grpServer *grpc.Server) {
	//	api.RegisterGoodsApiServer(grpServer, &api.GoodsApiService{})
	//})
	//server.Run()

	//3.tcp手写

	tcpServer := rpc.NewTcpServer("127.0.0.1", 9222)
	gob.Register(&model.Result{})
	gob.Register(&model.Goods{})
	tcpServer.Register("goods", &service.GoodsRpcService{})
	go tcpServer.Run()
	go r.Run(":9002", nil)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit
	tcpServer.Close()
	//
	//tcpServer := rpc.NewTcpServer("localhost", 9222)
	//tcpServer.Register("goods", &service.GoodsRpcService{})
	//tcpServer.Run()
	//r.Run(":9002", nil)
}
