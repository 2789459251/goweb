package service

import "web/zygo/rpc"

type GoodsService struct {
	Find func(args map[string]interface{}) ([]byte, error) `myrpc:"GET,/goods/find"`
}

func (*GoodsService) Env() rpc.HttpConfig {
	return rpc.HttpConfig{
		Host: "localhost",
		Port: 9002,
	}
}
