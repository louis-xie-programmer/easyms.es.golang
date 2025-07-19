// Package watchjob 检测数据变化（sqlserver 数据库中商品和价格数据是否新增数据）
// 当前job 是简单的JOB,这里没有继承baseJob
package watchjob

import (
	"easyms-es/config"
	easylib "easyms-es/crob_job/lib"
	"easyms-es/db"
	"easyms-es/model"
	"fmt"
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
	settings, exit := config.GetTaskConfigValue[model.WatchjobConfig](job.JobName)
	if !exit {
		return fmt.Errorf("config is not exit: %s", job.JobName)
	}
	job.JobConfig = *settings
	return nil
}

func (job *EasyJob) SaveSyncConfig(maxpid int) error {
	config.UpdateTaskConfig(job.JobName, job.JobName+".maxpid", maxpid)
	return config.SaveTaskConfig(job.JobName)
}

func (job *EasyJob) SaveSyncStatusConfig(status int) error {
	config.UpdateTaskConfig(job.JobName, job.JobName+".status", status)
	return config.SaveTaskConfig(job.JobName)
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

	var maxPid int
	//sql := "SELECT max(PID) FROM Products with(nolock)"
	//db.BasicDB.Raw(sql).Scan(&maxPid)

	type Product struct {
		PID int
	}
	var lastUser Product
	db.TenantPoolInstance.GetTable("Products").Last(&lastUser)
	maxPid = lastUser.PID

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
