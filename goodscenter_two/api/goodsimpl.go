package api

import "context"

type GoodsApiService struct {
}

func (GoodsApiService) Find(context.Context, *GoodsRequest) (*GoodsResponse, error) {
	goods := &Goods{Id: 1000, Name: "商品中心9002商品,grpc提供"}
	res := &GoodsResponse{
		Code: 200,
		Msg:  "success",
		Data: goods,
	}
	return res, nil
}
func (GoodsApiService) mustEmbedUnimplementedGoodsApiServer() {}
