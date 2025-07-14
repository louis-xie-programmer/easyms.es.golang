// Package watchjob 检测数据变化（sqlserver 数据库中商品和价格数据是否新增数据）
// 当前job 是简单的JOB,这里没有继承baseJob
package watchjob

import (
	"easyms-es/config"
	easylib "easyms-es/crob_job/lib"
	"easyms-es/db"
	"easyms-es/model"
	"encoding/json"
	"log"
	"time"
)

type EasyJob struct {
	JobConfig   model.WatchjobConfig
	JobName     string
	Description string
	RetryCount  int
}

func (job *EasyJob) DoError(err error, description string) {
	job.RetryCount++
	log.Println(description, err)
	easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
		JobName:     job.JobName,
		Status:      -1,
		Description: description + "\n" + err.Error(),
		RetryCount:  job.RetryCount,
	})
}

func (job *EasyJob) GetSyncConfig() error {
	settings := config.EasyViperConfigListJobNameFirst(job.JobName).ConfigFile.AllSettings()
	jsonData, err := json.Marshal(settings[job.JobName])
	if err != nil {
		job.DoError(err, "序列化参数失败:")
	}
	return json.Unmarshal(jsonData, &job.JobConfig)
}

func (job *EasyJob) SaveSyncConfig(maxpid int) error {
	config.UpdateSyncConfig(job.JobName, job.JobName+".maxpid", maxpid)
	return config.Save(job.JobName)
}

func (job *EasyJob) SaveSyncStatusConfig(status int) error {
	config.UpdateSyncConfig(job.JobName, job.JobName+".status", status)
	return config.Save(job.JobName)
}

func (job *EasyJob) Run() {
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
		_ = job.SaveSyncStatusConfig(-2)
		return
	}

	easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
		JobName:     job.JobName,
		LastRun:     start,
		Status:      2,
		Description: "",
		Interval:    0,
	})

	sql := "SELECT max(PID) FROM Products with(nolock)"
	var maxPid int
	db.BasicDB.Raw(sql).Scan(&maxPid)

	if job.JobConfig.Maxpid < maxPid {
		err = job.SaveSyncConfig(maxPid)
		if err != nil {
			job.DoError(err, "SaveSyncConfig error:")
			return
		}
	}

	easylib.EasyJobManager.UpdateEasyJobInfo(&easylib.EasyJobParam{
		JobName:  job.JobName,
		Status:   1,
		Interval: time.Since(start),
	})
}
