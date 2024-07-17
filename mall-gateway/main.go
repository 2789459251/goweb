package main

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"net/http"
	"web/zygo"
	"web/zygo/gateway"
	"web/zygo/register"
)

func main() {
	engine := zygo.Default()
	engine.OpenGateway = true
	var configs []gateway.GWConfig
	configs = append(configs, gateway.GWConfig{
		Name: "order",
		Path: "/order/**",
		Host: "127.0.0.1",
		Port: 9003,
		Header: func(req *http.Request) {
			req.Header.Set("zy", "画家")
		},
		ServiceName: "oederCenter",
	}, gateway.GWConfig{
		Name: "goods",
		Path: "/goods/**",
		//Host: "127.0.0.1",
		//Port: 9002,
		Header: func(req *http.Request) {
			req.Header.Set("zy", "画家")
		},
		ServiceName: "goodsCenter",
	})
	engine.SetGateConfigs(configs)
	engine.RegisterType = "nacos"
	engine.RegisterOption = register.Option{
		DialTimeout: 5000,
		NacosServerConfig: []constant.ServerConfig{
			{
				IpAddr:      "127.0.0.1",
				ContextPath: "/nacos",
				Port:        8848,
				Scheme:      "http",
			},
		},
	}
	engine.Run(":80")
}
