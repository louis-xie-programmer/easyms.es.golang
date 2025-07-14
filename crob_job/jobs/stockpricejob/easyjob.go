// Package stockpricejob 分销商价格维护Job
package stockpricejob

import (
	"easyms-es/model"
	"easyms-es/service/prices"
	"fmt"

	easylib "easyms-es/crob_job/lib"
)

type EasyJob[T model.ByIdConfig] struct {
	easylib.BaseJob[model.ByIdConfig]
}

func (job *EasyJob[T]) CallBackDo(data []interface{}, delData []interface{}) error {
	return nil
}

// GetDataPageList 获取新增或更新的商品价格，待删除的价格
func (job *EasyJob[T]) GetDataPageList() ([]interface{}, []interface{}, interface{}, error) {
	jobConfig := job.JobConfig
	stockPrices, delPrices, lastId, err := QueryStockPrice(jobConfig.Lastid, jobConfig.Limit)
	if err != nil {
		return nil, nil, jobConfig.Lastid, err
	}
	data := easylib.MapperFromPrices(stockPrices)
	deleteData := easylib.MapperFromPrices(delPrices)
	return data, deleteData, lastId, nil
}

// RemoveEs 删除
func (job *EasyJob[T]) RemoveEs(data []interface{}) error {
	stockPrices := easylib.MapperToPrices(data)
	err := prices.BulkRemove(stockPrices)
	if err != nil {
		return err
	}
	return nil
}

// UpdateEs 更新， 为追求效率，这里用批量插入，原始价格数据直接使用逻辑删除加新增，不做修改操作
func (job *EasyJob[T]) UpdateEs(data []interface{}) error {
	stockPrices := easylib.MapperToPrices(data)
	err := prices.BulkInsert(stockPrices)
	if err != nil {
		return err
	}
	return nil
}

func QueryStockPrice(lastSid int, limit int) ([]model.StockPrice, []model.StockPrice, int, error) {
	// 分销商价格表查询（入驻）
	sql := fmt.Sprintf(`SELECT TOP (%d) SID,PID,ProductName,Brand,BrandID,DistributorID
				,Distributor,DistributorProductUrl,StockNum,Currency,StepPrice
				,UpdateTime,IsDeleted
			FROM PriceStock with(nolock)
			where SID > %d and IsDeleted = 0 order by sid`, limit, lastSid)

	return easylib.QueryStockPrice(sql, model.EsPriceJoinStart)
}
