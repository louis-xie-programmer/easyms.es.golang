package prices

import (
	"easyms-es/model"
)

// BulkInsert 批量插入
func BulkInsert(prices []model.StockPrice) error {
	return bulkPrices(prices)
}

// BulkRemove 批量删除
func BulkRemove(prices []model.StockPrice) error {
	return bulkPrices(prices, "delete")
}

// 批量插入索引 对象要进行处理_id赋值
func bulkPrices(prices []model.StockPrice, flag ...string) error {
	var data []interface{}
	for _, price := range prices {
		data = append(data, price)
	}
	return PriceStore.Bulk(data, flag...)
}
