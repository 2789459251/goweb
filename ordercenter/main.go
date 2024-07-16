package main

import (
	"context"
	"encoding/json"
	"fmt"
	"goodscenter/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"ordercenter/api"
	service2 "ordercenter/service"
	"web/zygo"
	"web/zygo/rpc"
)

func main() {
	r := zygo.Default()
	cli := rpc.NewHttpClient()
	cli.RegisterHttpService("goods", &service2.GoodsService{})
	group := r.Group("order")
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
		body, err := cli.Do("goods", "Find").(*service2.GoodsService).Find(params)
		if err != nil {
			panic(err)
		}
		v := &model.Result{}
		json.Unmarshal(body, v)
		ctx.JSON(http.StatusOK, v)
	})

	group.GET("/findGRPC", func(ctx *zygo.Context) {
		var serviceHost = "127.0.0.1:9111"
		conn, err := grpc.Dial(serviceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Println(err)
		}
		defer conn.Close()

		client := api.NewGoodsApiClient(conn)
		rsp, err := client.Find(context.TODO(), &api.GoodsRequest{})

		if err != nil {
			fmt.Println(err)
		}
		ctx.JSON(http.StatusOK, rsp)
	})
	r.Run(":9003", nil)
}
