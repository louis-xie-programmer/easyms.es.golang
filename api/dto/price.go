package dto

import (
	ms "easyms-es/protos/messages"
	"easyms-es/service/models"
	"easyms-es/utility"
	"strconv"
)

func MapperToSearchPricesResult(prices *models.SearchPriceResult) (*ms.SearchPricesResult, error) {
	var pbPrices ms.SearchPricesResult
	pbPrices.Total = prices.Total
	pbPrices.From = prices.From
	pbPrices.Size = prices.Size

	if len(prices.Results) > 0 {
		pbPrices.PID = int32(prices.Results[0].PID)
	}

	for _, price := range prices.Results {
		var pbPrice ms.ESStockPrice
		pbPrice.SID = price.SPID
		pbPrice.IsAuthorizeddealer = price.IsAuthorizeddealer
		pbPrice.DistributorID = int32(price.DistributorID)
		pbPrice.DistributorType = int32(price.DistributorType)

		// 分销商扩展字段，这里可以拿到自定义的分销商扩展字段
		distributorExt := utility.SplitStrToMap(price.DistributorExt)
		if len(distributorExt[0]) > 0 {
			pbPrice.Distributor = distributorExt[0][0]
		}

		// 产品扩展字段，这里可以拿到自定义的产品扩展字段
		priceExt := utility.SplitStrToMap(price.PriceExt)
		if len(priceExt[0]) > 0 {
			if len(priceExt[0]) > 1 {
				pbPrice.ProductName = priceExt[0][1]
			}
		}

		// 库存
		if price.StockNum > 0 {
			pbPrice.StockNum = int32(price.StockNum)
		}

		// 币种
		if len(price.Currency) > 0 {
			pbPrice.Currency = price.Currency
		}

		if len(price.StepPrice1) > 0 {
			price1, err := strconv.ParseFloat(price.StepPrice1, 32)
			if err == nil {
				pbPrice.Price1 = float32(price1)
			}
		}
		if len(price.StepPrice2) > 0 {
			price2, err := strconv.ParseFloat(price.StepPrice2, 32)
			if err == nil {
				pbPrice.Price2 = float32(price2)
			}
		}
		if len(price.StepPrice3) > 0 {
			price3, err := strconv.ParseFloat(price.StepPrice3, 32)
			if err == nil {
				pbPrice.Price3 = float32(price3)
			}
		}
		if len(price.StepPrice4) > 0 {
			price4, err := strconv.ParseFloat(price.StepPrice4, 32)
			if err == nil {
				pbPrice.Price4 = float32(price4)
			}
		}
		if len(price.StepPrice5) > 0 {
			price1w, err := strconv.ParseFloat(price.StepPrice5, 32)
			if err == nil {
				pbPrice.Price5 = float32(price1w)
			}
		}

		pbPrice.UpdatedUtc = price.UpdateTime.Format("2006-01-02 15:04:05")

		pbPrices.Data = append(pbPrices.Data, &pbPrice)
	}

	return &pbPrices, nil
}
