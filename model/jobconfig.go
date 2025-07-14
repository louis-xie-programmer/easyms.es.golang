package model

import (
	"time"
)

type JobConfig interface {
	LimitConfig | ByIdConfig | WatchjobConfig | ByTimeConfig
}

// Status: 1: 准备就绪; 2:正在运行; 0: 已停止; -1: 作业失败; -2: 因失败暂停
type LimitConfig struct {
	Cron   string
	Limit  int
	Status int
}

// Status:  1: 准备就绪; 2:正在运行; 0: 已停止; -1: 作业失败; -2: 因失败暂停
type ByIdConfig struct {
	Cron     string
	Limit    int
	Maxid    int
	Lastid   int
	Status   int
	LastTime time.Time
}

type ByTimeConfig struct {
	Cron     string
	Limit    int
	LastTime string
	Status   int
}

// Status:  1: 准备就绪; 2:正在运行; 0: 已停止; -1: 作业失败; -2: 因失败暂停
type WatchjobConfig struct {
	Cron   string
	Maxpid int
	Status int
}
