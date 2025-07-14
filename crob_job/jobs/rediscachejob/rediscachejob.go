// Package rediscachejob
// 用redis 缓存来存放一些常用的聚合结果数据，减少es压力，提升整体网站效果
// 这里只提供相应的思路，具体的实现按照自身的项目需求
package rediscachejob

import (
	"easyms-es/config"
	easylib "easyms-es/crob_job/lib"
	"easyms-es/model"
	"encoding/json"
	"log"
	"time"
)

type RedisCacheEasyJob struct {
	JobConfig   model.LimitConfig
	JobName     string
	Description string
	RetryCount  int
}

func (job *RedisCacheEasyJob) DoError(err error, description string) {
	job.RetryCount++
	if err != nil {
		log.Println(description, err)
		easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
			JobName:     job.JobName,
			Status:      -1,
			Description: description + "\n" + err.Error(),
			RetryCount:  job.RetryCount,
		})
	}
	easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
		JobName:     job.JobName,
		Status:      -1,
		Description: description,
		RetryCount:  job.RetryCount,
	})
}

func (job *RedisCacheEasyJob) GetSyncConfig() error {
	settings := config.EasyViperConfigListJobNameFirst(job.JobName).ConfigFile.AllSettings()
	jsonData, err := json.Marshal(settings[job.JobName])
	if err != nil {
		job.DoError(err, "序列化参数失败:")
	}
	return json.Unmarshal(jsonData, &job.JobConfig)
}

func (job *RedisCacheEasyJob) SaveSyncConfig(lastTime int) error {
	config.UpdateSyncConfig(job.JobName, job.JobName+".lasttime", lastTime)
	return config.Save(job.JobName)
}

// SaveSyncStatusConfig 只对停止,或者因为错误重试过多停止保存
func (job *RedisCacheEasyJob) SaveSyncStatusConfig(status int) error {
	config.UpdateSyncConfig(job.JobName, job.JobName+".status", status)
	return config.Save(job.JobName)
}

func (job *RedisCacheEasyJob) Run() {
	err := job.GetSyncConfig()
	if err != nil {
		job.DoError(err, "GetSyncConfig error:")
		return
	}

	start := time.Now()

	if job.RetryCount >= 5 {
		easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
			JobName:     job.JobName,
			Status:      -2,
			Description: "出错重试5次",
			RetryCount:  job.RetryCount,
			Interval:    0,
		})
		job.RetryCount = 0
		err := job.SaveSyncStatusConfig(-2)
		if err != nil {

		}
		return
	}

	easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
		JobName:     job.JobName,
		LastRun:     start,
		Status:      2,
		Description: "",
		Interval:    0,
	})

	// 这里具体去实现相关操作
	// redis 操作 参阅 /db/redis.go

	easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
		JobName:     job.JobName,
		Status:      0,
		Description: job.Description,
		Interval:    time.Since(start),
	})
	err = easylib.EasyJobManager.PauseJob(job.JobName)
	if err != nil {
		job.DoError(err, "job pause error:")
		return
	}
}
