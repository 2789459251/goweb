package main

import (
	"web/zygo"
	"web/zygo/gateway"
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
	}, gateway.GWConfig{
		Name: "goods",
		Path: "/goods/**",
		Host: "127.0.0.1",
		Port: 9002,
	})
	engine.SetGateConfigs(configs)
	engine.Run(":80")
}
