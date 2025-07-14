package prices

import (
	"easyms-es/easyes"
	"easyms-es/model"
	"easyms-es/service/models"
	"encoding/json"
)

var PriceStore easyes.Store

func Search(pid int, size int, from int) (models.SearchPriceResult, error) {
	var result models.SearchPriceResult
	result.From = int32(from)
	result.Size = int32(size)

	query := BuildSingleQuery(pid, from, size)
	res, err := PriceStore.Search(query)

	if err != nil {
		return result, err
	}

	err = ResponseToStockPrice(*res, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func ResponseToStockPrice(rep easyes.SearchResponse, result *models.SearchPriceResult) error {
	result.Total = int32(rep.Hits.Total.Value)
	result.Above = rep.Hits.Total.Relation == "gte"

	for _, hit := range rep.Hits.Hits {
		var h model.StockPrice

		if err := json.Unmarshal(hit.Source, &h); err != nil {
			return err
		}
		result.Results = append(result.Results, h)
	}

	return nil
}
