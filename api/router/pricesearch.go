package router

import (
	"context"
	"easyms-es/api/dto"
	"easyms-es/protos/messages"
	pb "easyms-es/protos/services"
	"easyms-es/service/prices"
)

type PriceEsServer struct {
	pb.PriceSearchServiceServer
}

// SearchPrices 单产品价格搜索
func (s *PriceEsServer) SearchPrices(ctx context.Context, req *messages.PriceSearchParam) (*messages.SearchPricesResult, error) {
	res, err := prices.Search(int(req.PID), int(req.Size), int(req.From))
	if err != nil {
		return nil, err
	}

	results, err := dto.MapperToSearchPricesResult(&res)
	if err != nil {
		return nil, err
	}

	return results, nil
}
