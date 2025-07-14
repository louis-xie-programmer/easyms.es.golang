package main

import (
	"easyms-es/config"
	"easyms-es/crob_job/jobs/productjob"
	"easyms-es/crob_job/jobs/rediscachejob"
	"easyms-es/crob_job/jobs/stockpricejob"
	"easyms-es/crob_job/jobs/watchjob"
	easylib "easyms-es/crob_job/lib"
	"easyms-es/db"
	"easyms-es/easyes"
	"easyms-es/model"
	"easyms-es/service/prices"
	"easyms-es/service/products"
	"fmt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

// 初始化项目
func init() {
	// 日志
	if err := config.Init("job"); err != nil {
		panic(err)
	}

	//初始化配置文件数组
	config.InitConfig()

	//追加主配置文件
	configName := "config"
	config.CreateConfig(configName, "./conf/job/"+configName+".yaml")

	// 数据库
	db.BasicDB = db.InitSqlServerDB("basicsqlserver")
	db.AnalysisDB = db.InitSqlServerDB("analysissqlserver")
	db.DataDB = db.InitSqlServerDB("datasqlserver")
	db.ManageDB = db.InitSqlServerDB("managesqlserver")
	db.JobDB = db.InitSqlServerDB("jobsqlserver")

	timeout := config.GetSyncConfig_Type[int]("", "common.elasticsearch.timeout")

	// es名称及store
	model.EsProductIndexName = config.GetSyncConfig("", "common.elasticsearch.mfgpartindex")
	model.EsPriceIndexName = config.GetSyncConfig("", "common.elasticsearch.pricestockindex")

	store1, err := easyes.NewStore(easyes.StoreConfig{
		IndexName: model.EsProductIndexName,
		Timeout:   time.Duration(timeout) * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	products.ProductStore = *store1

	store2, err := easyes.NewStore(easyes.StoreConfig{
		IndexName: model.EsPriceIndexName,
		Timeout:   time.Duration(timeout) * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	prices.PriceStore = *store2

	//redis
	db.InitRedis()
}

// InitEasyJobManager 通过配置文件初始化
func InitEasyJobManager() {
	easylib.EasyJobManager = easylib.NewJobManager()

	// 监控程序
	var watchJob = &watchjob.EasyJob{
		JobName: "watchJob",
	}
	_ = easylib.EasyJobManager.AddJob(watchJob.JobName, watchJob.Run)

	// 任务一 ： 新增型号
	var productInsertJob = &productjob.EasyJob[model.ByIdConfig]{}
	productInsertJob.EasyFunc = productInsertJob
	productInsertJob.JobName = "productInsert"
	_ = easylib.EasyJobManager.AddJob(productInsertJob.JobName, productInsertJob.Run)

	// 任务三 : 分销商价格
	var stockPriceJob = &stockpricejob.EasyJob[model.ByIdConfig]{}
	stockPriceJob.EasyFunc = stockPriceJob
	stockPriceJob.JobName = "stockPrice"
	_ = easylib.EasyJobManager.AddJob(stockPriceJob.JobName, stockPriceJob.Run)

	// 任务八: 缓存数据维护
	var cacheJob = &rediscachejob.RedisCacheEasyJob{
		JobName: "redisCache",
	}
	_ = easylib.EasyJobManager.AddJob(cacheJob.JobName, cacheJob.Run)

}

// 项目启动
func main() {
	InitEasyJobManager()

	easylib.EasyJobManager.Start()

	defer easylib.EasyJobManager.Stop()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		c.Next()
	})

	webui := config.GetSyncConfig("", "common.webui.root")

	r.Use(static.Serve("/", static.LocalFile(webui, false)))

	// 获取任务列表
	r.GET("/jobs", func(c *gin.Context) {
		jobs := easylib.EasyJobManager.ListJobs()
		c.JSON(http.StatusOK, jobs)
	})
	// 暂停任务
	r.POST("/jobs/:jobname/pause", func(c *gin.Context) {
		jobName := c.Param("jobname")
		err := easylib.EasyJobManager.PauseJob(jobName)
		message := "Job paused"
		if err != nil {
			message = err.Error()
		}
		c.JSON(http.StatusOK, gin.H{"message": message})
	})
	// 重启任务
	r.POST("/jobs/:jobname/resume", func(c *gin.Context) {
		jobName := c.Param("jobname")
		err := easylib.EasyJobManager.ResumeJob(jobName)
		message := "Job resumed"
		if err != nil {
			message = err.Error()
		}
		c.JSON(http.StatusOK, gin.H{"message": message})
	})
	// 修改任务执行周期, 未知原因:前端请求过来的int会变成string对象,而且是概率事件,怀疑是vite打包问题,暂时用any实现
	r.POST("/jobs/update", func(c *gin.Context) {
		var job struct {
			JobName string `json:"JobName,omitempty"`
			Cron    string `json:"Cron,omitempty"`
			Limit   any    `json:"Limit,omitempty"`
		}

		if err := c.ShouldBindJSON(&job); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		limit, err := strconv.Atoi(fmt.Sprintf("%v", job.Limit))
		message := "Job updated"
		if err != nil {
			message = err.Error()
		} else {
			err = easylib.EasyJobManager.UpdateJobCron(job.JobName, job.Cron, limit)
			if err != nil {
				message = err.Error()
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": message})
	})

	_ = r.Run(fmt.Sprintf(":%d", 8087))
}
