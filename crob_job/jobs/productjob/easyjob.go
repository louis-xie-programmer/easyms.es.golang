// Package productjob 商品维护Job
package productjob

import (
	easylib "easyms-es/crob_job/lib"
	"easyms-es/model"
	"easyms-es/service/products"
	"fmt"
)

type EasyJob[T model.ByIdConfig] struct {
	easylib.BaseJob[model.ByIdConfig]
}

func (job *EasyJob[T]) CallBackDo(data []interface{}, delData []interface{}) error {
	return nil
}

func (job *EasyJob[T]) GetDataPageList() ([]interface{}, []interface{}, interface{}, error) {
	jobConfig := job.JobConfig
	maxPid := easylib.GetMaxPid()
	if jobConfig.Lastid >= maxPid {
		job.Description += "jobConfig.LastId >= maxPid \n"
		return nil, nil, maxPid, nil
	}
	ps, delProducts, lastId, err := QueryProduct(jobConfig.Lastid, jobConfig.Limit)
	if err != nil {
		return nil, nil, jobConfig.Lastid, err
	}
	data := easylib.MapperFromProducts(ps)
	deleteData := easylib.MapperFromProducts(delProducts)
	return data, deleteData, lastId, nil
}

func (job *EasyJob[T]) RemoveEs(data []interface{}) error {
	//products := easylib.MapperToProducts(datas)
	//err := products.BulkRemove(products)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (job *EasyJob[T]) UpdateEs(data []interface{}) error {
	ps := easylib.MapperToProducts(data)
	err := products.BulkInsert(ps)
	if err != nil {
		return err
	}
	return nil
}

func QueryProduct(lastId int, limit int) ([]model.Product, []model.Product, int, error) {
	param := fmt.Sprintf("PID > %d ", lastId)
	return easylib.QueryProduct(param, limit)
}
