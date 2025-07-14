package products

import (
	"easyms-es/easyes"
	"easyms-es/model"
)

var ProductStore easyes.Store

// BulkInsert 批量插入
func BulkInsert(products []model.Product) error {
	return bulkProducts(products, "index")
}

// 批量插入索引 对象要进行处理_id赋值
func bulkProducts(products []model.Product, flag ...string) error {
	var data []interface{}
	for _, part := range products {
		data = append(data, part)
	}
	return ProductStore.Bulk(data, flag...)
}
