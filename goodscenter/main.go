package main

import (
	"errors"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"goodscenter/model"
	"log"
	"net/http"
	"web/zygo"
	"web/zygo/breaker"
	"web/zygo/register"
)

func main() {
	r := zygo.Default()
	//r.Use(zygo.Limiter(1, 1))

	var cb = breaker.NewCircuitBreaker(breaker.Settings{})
	group := r.Group("goods")
	group.GET("/find", func(ctx *zygo.Context) {
		result, err := cb.Execute(func() (any, error) {
			//网关可以配置header信息
			//v := ctx.GetHeader("zy")
			//fmt.Println("get zygo" + v)

			query := ctx.GetQuery("id")
			if query == "2" {
				return nil, errors.New("测试熔断")
			}
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
				return nil, err
			}
			cli.RegisterService("goodsCenter", "127.0.0.1", 9002)
			good := &model.Goods{
				ID:   1,
				Name: "跳跳糖",
			}

			return good, nil
		})
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusInternalServerError, &model.Result{
				Code: 500,
				Msg:  err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, &model.Result{
			Code: 200,
			Msg:  "succeess",
			Data: result,
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
	/*电脑*/ /*
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
		tcpServer.SetLimiter(10, 100)
		tcpServer.LimiterTimeout = 1
		cli.RegisterService("goodsCenter", "127.0.0.1", 9002)
		go tcpServer.Run()*/
	r.Run(":9002")
	//quit := make(chan os.Signal)
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	//<-quit
	//tcpServer.Close()

	//tcpServer := rpc.NewTcpServer("localhost", 9222)
	//tcpServer.Register("goods", &service.GoodsRpcService{})
	//tcpServer.Run()
	//r.Run(":9002", nil)

}
