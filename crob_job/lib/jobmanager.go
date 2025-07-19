package lib

import (
	"easyms-es/config"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/vanh01/lingo"
)

var EasyJobManager *JobManager

// EasyJob 任务, jobName 是唯一值
type EasyJob struct {
	JobName     string
	EntryID     cron.EntryID
	Cron        string
	LastRun     time.Time
	Interval    time.Duration
	Status      int
	Description string
	RetryCount  int
	JobFunc     func()
	Limit       int
}

type EasyJobParam struct {
	JobName     string
	LastRun     time.Time
	Interval    time.Duration
	Status      int
	Description string
	RetryCount  int
}

type EasyJobResponseData struct {
	JobName     string
	Cron        string
	LastRun     string
	Interval    int
	Status      int
	Description string
	RetryCount  int
	Limit       int
}

type JobManager struct {
	cron      *cron.Cron
	jobs      map[string]*EasyJob
	nextJobID int
	mutex     sync.Mutex
}

func NewJobManager() *JobManager {
	return &JobManager{
		cron: cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))),
		jobs: make(map[string]*EasyJob),
	}
}

func (jm *JobManager) Start() {
	jm.cron.Start()
}

func (jm *JobManager) Stop() {
	jm.cron.Stop()
}

// AddJob 添加Job
func (jm *JobManager) AddJob(jobName string, jobFunc func()) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	config.CreateTaskConfig(jobName, "./conf/job/jobList/"+jobName+".yaml")

	job, ok := config.GetTaskConfigValue[EasyJob](jobName)
	if !ok {
		return errors.New("job not found")
	}

	//cronValue := config.GetSyncConfig(jobName, jobName+".cron")
	//statusValue := config.GetSyncConfig_Type[int](jobName, jobName+".status")
	//limitValue := config.GetSyncConfig_Type[int](jobName, jobName+".limit")
	//
	//job := &EasyJob{
	//	JobName:    jobName,
	//	Cron:       cronValue,
	//	JobFunc:    jobFunc,
	//	RetryCount: -1,
	//	Status:     statusValue,
	//	Limit:      limitValue,
	//}

	if job.Status > 0 {
		entryID, err := jm.cron.AddFunc(job.Cron, jobFunc)
		if err != nil {
			return err
		}
		job.EntryID = entryID
	} else {
		job.EntryID = 0
	}

	jm.jobs[jobName] = job

	return nil
}

// RemoveJob 移除操作, 物理移除, 谨慎操作
func (jm *JobManager) RemoveJob(jobName string) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job, exists := jm.jobs[jobName]
	if exists {
		jm.cron.Remove(job.EntryID)
		delete(jm.jobs, jobName)
	}

	config.UpdateTaskConfig(jobName, jobName, nil)
	err := config.SaveTaskConfig(jobName)

	if err != nil {
		return err
	}
	return nil
}

// PauseJob 暂停job
func (jm *JobManager) PauseJob(jobName string) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job, exists := jm.jobs[jobName]
	if exists && job.EntryID > 0 {
		jm.cron.Remove(job.EntryID)
		jm.jobs[jobName].Status = 0
		jm.jobs[jobName].EntryID = 0
	}

	config.UpdateTaskConfig(jobName, jobName+".status", 0)
	err := config.SaveTaskConfig(jobName)

	if err != nil {
		return err
	}
	return nil
}

// ResumeJob 编辑修改任务配置
func (jm *JobManager) ResumeJob(jobName string) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job, exists := jm.jobs[jobName]
	if exists && job.EntryID == 0 {
		entryID, err := jm.cron.AddFunc(job.Cron, job.JobFunc)
		if err != nil {
			return err
		}
		jm.jobs[jobName].EntryID = entryID
		jm.jobs[jobName].Status = 1
	}

	//删除旧配置
	config.RemoveTaskConfig(jobName)

	//重新追加配置文件
	config.CreateTaskConfig(jobName, "./conf/job/jobList/"+jobName+".yaml")

	config.UpdateTaskConfig(jobName, jobName+".status", 1)
	err := config.SaveTaskConfig(jobName)

	if err != nil {
		return err
	}
	return nil
}

// UpdateJobCron 更新Job Cron表达式
func (jm *JobManager) UpdateJobCron(jobName string, cron string, limit int) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	//暂停原job
	oldJob, exists := jm.jobs[jobName]

	if !exists {
		return errors.New("job not found")
	}

	if oldJob.EntryID != 0 {
		jm.cron.Remove(oldJob.EntryID)
		jm.jobs[jobName].Status = 0
		jm.jobs[jobName].EntryID = 0
	}

	entryID, err := jm.cron.AddFunc(cron, oldJob.JobFunc)
	if err != nil {
		fmt.Println("Error resuming job:", err)
		return err
	}
	jm.jobs[jobName].Cron = cron
	jm.jobs[jobName].Limit = limit
	jm.jobs[jobName].EntryID = entryID
	jm.jobs[jobName].Status = 1

	//删除旧配置
	config.RemoveTaskConfig(jobName)

	//重新追加配置文件
	config.CreateTaskConfig(jobName, "./conf/job/jobList/"+jobName+".yaml")

	config.UpdateTaskConfig(jobName, jobName+".cron", cron)
	config.UpdateTaskConfig(jobName, jobName+".limit", limit)
	config.UpdateTaskConfig(jobName, jobName+".status", 1)
	err = config.SaveTaskConfig(jobName)

	if err != nil {
		return err
	}
	return nil
}

// ListJobs 列出所有Job
func (jm *JobManager) ListJobs() []*EasyJobResponseData {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	var jobs []*EasyJobResponseData

	for _, job := range jm.jobs {
		var interval int
		interval = (int)(job.Interval) / 1000000
		jobs = append(jobs, &EasyJobResponseData{
			JobName:     job.JobName,
			Cron:        job.Cron,
			LastRun:     job.LastRun.Format("2006-01-02 15:04:05"),
			Interval:    interval,
			Status:      job.Status,
			Description: job.Description,
			RetryCount:  job.RetryCount,
			Limit:       job.Limit,
		})
	}
	jobs = lingo.AsEnumerable(jobs).OrderBy(func(data *EasyJobResponseData) any {
		return data.JobName
	}).ToSlice()
	return jobs
}

// UpdateEasyJobInfo 保存更新任务信息
func (jm *JobManager) UpdateEasyJobInfo(jobParam *EasyJobParam) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()

	job, exists := jm.jobs[jobParam.JobName]

	if !exists {
		return
	}
	if !jobParam.LastRun.IsZero() {
		job.LastRun = jobParam.LastRun
	}
	if jobParam.Interval > 0 {
		job.Interval = jobParam.Interval
	}

	job.Status = jobParam.Status

	if jobParam.Description != "" {
		job.Description = jobParam.Description
	}
	if jobParam.RetryCount != 0 {
		job.RetryCount = jobParam.RetryCount
	}
}
