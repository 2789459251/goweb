package service

import "goodscenter/model"

type GoodsRpcService struct {
}

func (*GoodsRpcService) Find(id int) *model.Result {
	goods := model.Goods{ID: 1000, Name: "商品中心9002商品"}
	return &model.Result{Code: 200, Msg: "success", Data: goods}
}
