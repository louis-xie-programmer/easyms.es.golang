package router

import (
	"context"
	"easyms-es/api/dto"
	"easyms-es/protos/messages"
	pb "easyms-es/protos/services"
	"easyms-es/service/products"
)

type ProductEsServer struct {
	pb.ProductsSearchServiceServer
}

// Analyze 商品搜索关键字分析
func (s *ProductEsServer) Analyze(ctx context.Context, req *messages.ProductSearchParam) (*messages.Tokens, error) {
	tokens, err := products.Analyze("easy_all", req.KeyWord)
	if err != nil {
		return nil, err
	}
	rst := dto.MapperToPdToken(tokens)
	return rst, nil
}
