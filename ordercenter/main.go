package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"goodscenter/model"
	"log"
	"net/http"
	service2 "ordercenter/service"
	"time"
	"web/zygo"
	"web/zygo/register"
	"web/zygo/rpc"
	"web/zygo/tracer"
)

func main() {
	r := zygo.Default()

	//http调用
	cli := rpc.NewHttpClient()
	cli.RegisterHttpService("goods", &service2.GoodsService{})
	group := r.Group("order")
	createTracer, closer, err := tracer.CreateTracer("orderCenter",
		&config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		&config.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: "http://127.0.0.1:14268/api/traces",
		}, config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		log.Println(err)
	}
	defer closer.Close()
	group.GET("/find", func(ctx *zygo.Context) {
		//调用商品模块
		//http -> 调用
		//body, err := cli.Get("http://localhost:9002/goods/find")
		params := make(map[string]interface{})
		params["id"] = 1000
		params["name"] = "zy"
		//req, err := cli.FormRequest("GET", "http://localhost:9002/goods/find", params)
		//if err != nil {
		//	panic(err)
		//}
		//body, err := cli.Response(req)
		//if err != nil {
		//	return
		//}
		//log.Println(string(body))
		span := createTracer.StartSpan("find")
		defer span.Finish()
		session := cli.NewSession()

		session.ReqHandler = func(req *http.Request) {
			ext.SpanKindRPCClient.Set(span)
			//携带信息
			createTracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		}
		body, err := session.Do("goods", "Find").(*service2.GoodsService).Find(params)
		if err != nil {
			panic(err)
		}
		v := &model.Result{}
		json.Unmarshal(body, v)
		ctx.JSON(http.StatusOK, v)
	})

	group.GET("/findGRPC", func(ctx *zygo.Context) {
		//var serviceHost = "127.0.0.1:9111"
		/*1.grpc*/
		//conn, err := grpc.Dial(serviceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		//if err != nil {
		//	fmt.Println(err)
		//}
		//defer conn.Close()
		//
		//client := api.NewGoodsApiClient(conn)
		//rsp, err := client.Find(context.TODO(), &api.GoodsRequest{})
		//
		//if err != nil {
		//	fmt.Println(err)
		//}

		/*2.封装grpc*/
		//Config := rpc.DefaultGrpcClientConfig()
		//Config.Address = serviceHost
		//client, _ := rpc.NewGrpcClient(Config)
		//defer client.Conn.Close()
		//
		//goodsApiClient := api.NewGoodsApiClient(client.Conn)
		//rsp, _ := goodsApiClient.Find(context.Background(), &api.GoodsRequest{})

		/*3.基于tcp实现的rpc*/

		option := rpc.DefaultOption
		option.SerializeType = rpc.ProtoBuff
		option.RegisterType = "nacos"
		option.RegisterOption = register.Option{
			DialTimeout: 5 * time.Second,
			NacosServerConfig: []constant.ServerConfig{
				{
					IpAddr:      "127.0.0.1",
					ContextPath: "/nacos",
					Port:        8848,
					Scheme:      "http",
				},
			},
			NacosClientConfig: constant.NewClientConfig(
				constant.WithNamespaceId(""), //当namespace是public时，此处填空字符串。
				constant.WithTimeoutMs(5000),
				constant.WithNotLoadCacheAtStart(true),
				constant.WithLogDir("/tmp/nacos/log"),
				constant.WithCacheDir("/tmp/nacos/cache"),
				constant.WithLogLevel("debug"),
			),
		}

		proxy := rpc.NewMyTcpClientProxy(option)
		//params := make([]any, 1)
		//params[0] = int64(1)
		////todo 调用方法完善
		//->这样的 body, err := cli.Do("goods", "Find").(*service2.GoodsService).Find(params)
		//result, err := proxy.Call(context.Background(), "goods", "Find", params)
		//log.Panicln(err)

		gob.Register(&model.Result{})
		gob.Register(&model.Goods{})
		args := make([]any, 1)
		args[0] = 1
		result, err := proxy.Call(context.Background(), "goods", "Find", args)
		if err != nil {
			panic(err)
		}
		ctx.JSON(http.StatusOK, result)
	})
	r.Run(":9003")
}
