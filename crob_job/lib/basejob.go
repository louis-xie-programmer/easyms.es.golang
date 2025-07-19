package lib

import (
	"easyms-es/config"
	"easyms-es/model"
	"fmt"
	"log"
	"time"
)

// BaseJob Job基类
type BaseJob[T model.JobConfig] struct {
	JobConfig   T
	JobName     string
	Description string
	RetryCount  int
	EasyFunc    interface{}
}

// DoError 任务过程中储物处理（Job状态更新，新增错误日志）
func (job *BaseJob[T]) DoError(err error, description string) {
	job.RetryCount++
	log.Println(description, err)
	EasyJobManager.UpdateEasyJobInfo(&EasyJobParam{
		JobName:     job.JobName,
		Status:      -1,
		Description: description + "\n" + err.Error(),
		RetryCount:  job.RetryCount,
	})
}

// EasyFunc 需要反射调用的方法定义
type EasyFunc interface {
	GetDataPageList() ([]interface{}, []interface{}, interface{}, error)
	UpdateEs([]interface{}) error
	RemoveEs([]interface{}) error
	CallBackDo([]interface{}, []interface{}) error
}

// GetSyncConfig 获取项目配置（任务配置）
func (job *BaseJob[T]) GetSyncConfig() error {
	settings, exit := config.GetTaskConfigValue[T](job.JobName)
	if exit {
		return fmt.Errorf("config is not exit: %s", job.JobName)
	}
	job.JobConfig = *settings
	return nil
}

// SaveSyncConfig 报错任务配置表（状态）
func (job *BaseJob[T]) SaveSyncConfig(lastId int, lastTime string) error {
	if lastId > 0 {
		config.UpdateTaskConfig(job.JobName, job.JobName+".lasTid", lastId)
	}
	if lastTime != "" {
		config.UpdateTaskConfig(job.JobName, job.JobName+".lastTime", lastTime)
	}

	return config.SaveTaskConfig(job.JobName)
}

// CallBackDo 完成单次任务后，回调方法
func (job *BaseJob[T]) CallBackDo(data []interface{}, delData []interface{}) error {
	return nil
}

// SaveSyncStatusConfig 只对停止,或者因为错误重试过多停止保存
func (job *BaseJob[T]) SaveSyncStatusConfig(status int) error {
	config.UpdateTaskConfig(job.JobName, job.JobName+".status", status)
	return config.SaveTaskConfig(job.JobName)
}

// GetDataPageList 获取Job中要处理的数据（新增或更新数据，删除数据）
func (job *BaseJob[T]) GetDataPageList() ([]interface{}, []interface{}, interface{}, error) {
	return nil, nil, nil, fmt.Errorf("GetDataPageList 未定义")
}

// UpdateEs 更新es数据
func (job *BaseJob[T]) UpdateEs([]interface{}) error {
	return fmt.Errorf("UpdateEs 未定义")
}

// RemoveEs 删除es数据
func (job *BaseJob[T]) RemoveEs([]interface{}) error {
	return fmt.Errorf("RemoveEs 未定义")
}

// Run Job 运行方法
func (job *BaseJob[T]) Run() {
	start := time.Now()

	if job.RetryCount >= 5 {
		EasyJobManager.UpdateEasyJobInfo(&EasyJobParam{
			JobName:     job.JobName,
			Status:      -2,
			Description: "出错重试5次",
			RetryCount:  job.RetryCount,
			Interval:    0,
		})
		job.RetryCount = 0
		var err = job.SaveSyncStatusConfig(-2)
		if err != nil {
			job.DoError(err, "SaveSyncStatusConfig error:")
		}
		return
	}

	EasyJobManager.UpdateEasyJobInfo(&EasyJobParam{
		JobName:     job.JobName,
		LastRun:     start,
		Status:      2,
		Description: "",
		Interval:    0,
	})

	err := job.GetSyncConfig()
	if err != nil {
		job.DoError(err, "GetSyncConfig error:")
		return
	}

	data, delData, last, err := job.EasyFunc.(EasyFunc).GetDataPageList()
	if err != nil {
		job.DoError(err, "GetDataPageList error:")
		return
	}

	if len(data) == 0 {
		job.Description += "UpdateEs datas is null.\n"
	} else {
		job.Description += fmt.Sprintf("update pid count: %d \n", len(data))
		err := job.EasyFunc.(EasyFunc).UpdateEs(data)
		if err != nil {
			job.DoError(err, "UpdateEs error:")
			return
		}
	}
	if len(delData) > 0 {
		job.Description += fmt.Sprintf("remove pid count: %d \n", len(delData))
		err = job.EasyFunc.(EasyFunc).RemoveEs(delData)
		if err != nil {
			job.DoError(err, "Job Run error:")
			return
		}
	}

	switch v := last.(type) {
	case int:
		if v > 0 {
			err = job.SaveSyncConfig(v, "")
			if err != nil {
				job.DoError(err, "SaveSyncConfig error:")
				return
			}
		}
	case string:
		if v != "" {
			err = job.SaveSyncConfig(0, v)
			if err != nil {
				job.DoError(err, "SaveSyncConfig error:")
				return
			}
		}
	}

	err = job.EasyFunc.(EasyFunc).CallBackDo(data, delData)

	if err != nil {
		job.DoError(err, "CallBackDo error:")
		return
	}

	EasyJobManager.UpdateEasyJobInfo(&EasyJobParam{
		JobName:     job.JobName,
		Status:      1,
		Description: job.Description,
		Interval:    time.Since(start),
	})
	job.Description = ""
}
