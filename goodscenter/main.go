package main

import (
	"encoding/gob"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"goodscenter/model"
	"goodscenter/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"web/zygo"
	"web/zygo/register"
	"web/zygo/rpc"
)

func main() {
	r := zygo.Default()
	r.Use(zygo.Limiter(1, 1))
	group := r.Group("goods")
	group.GET("/find", func(ctx *zygo.Context) {
		v := ctx.GetHeader("zy")
		fmt.Println("get zygo" + v)

		cli := register.NacosRegister{}
		err := cli.CreateCli(register.Option{
			DialTimeout: 5000,
			NacosServerConfig: []constant.ServerConfig{
				{
					IpAddr:      "127.0.0.1",
					ContextPath: "/nacos",
					Port:        8848,
					Scheme:      "http",
				},
			},
		})
		if err != nil {
			return
		}
		cli.RegisterService("goodsCenter", "127.0.0.1", 9002)
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

	//tcpServer := rpc.NewTcpServer("127.0.0.1", 9222)
	//tcpServer.RegisterType = "etcd"
	//tcpServer.RegisterOption = register.Option{
	//	Endpoints:   []string{"127.0.0.1:2379"},
	//	DialTimeout: 5 * time.Second,
	//	Host:        "127.0.0.1",
	//	Port:        9222,
	//}
	//注册中心抽象出来的应用
	tcpServer := rpc.NewTcpServer("localhost", 9112)
	tcpServer.SetRegister("nacos", register.Option{
		DialTimeout: 5000,
		NacosServerConfig: []constant.ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				ContextPath: "/nacos",
				Port:        8848,
				Scheme:      "http",
			},
		},
	})
	gob.Register(&model.Result{})
	gob.Register(&model.Goods{})
	tcpServer.Register("goods", &service.GoodsRpcService{})

	cli := register.NacosRegister{}
	err := cli.CreateCli(register.Option{
		DialTimeout: 5000,
		NacosServerConfig: []constant.ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				ContextPath: "/nacos",
				Port:        8848,
				Scheme:      "http",
			},
		},
	})
	if err != nil {
		return
	}
	tcpServer.SetLimiter(1, 1)
	tcpServer.LimiterTimeout = 1
	cli.RegisterService("goodsCenter", "127.0.0.1", 9002)
	go tcpServer.Run()
	go r.Run(":9002")
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
