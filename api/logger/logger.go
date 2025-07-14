package logger

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// 日志数据库连接变量
var (
	db         *sql.DB
	logChannel = make(chan LogEntry, 100) // 异步日志通道
	wg         = &sync.WaitGroup{}       // 协程等待组
)

// LogEntry 日志实体结构
// 包含服务、方法、客户端信息、状态码等字段
type LogEntry struct {
	Service    string  // 服务名称
	Method     string  // 方法名称
	ClientID   string  // 客户端ID
	ClientIP   string  // 客户端IP
	UserIP     string  // 用户IP
	UserAgent  string  // 用户代理
	StatusCode int     // 状态码
	Latency    int     // 延迟（毫秒）
	Timestamp  time.Time // 时间戳
	Params     string  // 请求参数
	Error      string  // 错误信息
}

// InitLogger 初始化日志系统
// 参数:
//   connString - 数据库连接字符串
// 返回:
//   error - 错误信息
func InitLogger(connString string) error {
	// 日志操作
	var err error

	// sql server 数据库来记录访问日志
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}

	// Start log processor
	wg.Add(1)
	go logProcessor()

	return nil
}

// logProcessor 日志处理协程
// 负责异步写入文本日志和批量插入数据库
func logProcessor() {
	defer wg.Done()

	// 文本日志来记录错误
	file := "./logs/" + time.Now().Format("2024-01-01") + "_error.log"
	logFile, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Printf("error closing log file: %v\n", err)
		}
	}()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logFile)

	var logs []LogEntry
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case logEntry := <-logChannel:
			if logEntry.StatusCode > 0 {
				log.Println(logEntry.Method)
				log.Println(logEntry.Params)
				log.Println(logEntry.Error)
			}
			logs = append(logs, logEntry)
			if len(logs) >= 10 {
				insertDbLogs(logs)
				logs = []LogEntry{}
			}
		case <-ticker.C:
			if len(logs) > 0 {
				insertDbLogs(logs)
				logs = []LogEntry{}
			}
		}
	}
}

// insertDbLogs 批量插入数据库日志
// 参数:
//   logs - 日志条目数组
func insertDbLogs(logs []LogEntry) {
	if len(logs) == 0 {
		return
	}

	query := "INSERT INTO easyESTraceLogs (Service, Method, ClientID, ClientIP, UserIP, UserAgent, StatusCode, Latency, Timestamp) VALUES "
	values := ""

	// 插入数据库
	for i, logInfo := range logs {
		if i > 0 {
			values += ","
		}
		values += fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', %d, %d, '%s')",
			logInfo.Service, logInfo.Method, logInfo.ClientID, logInfo.ClientIP, logInfo.UserIP, logInfo.UserAgent, logInfo.StatusCode, logInfo.Latency, logInfo.Timestamp.Format("2006-01-02 15:04:05"))
	}

	_, err := db.Exec(query + values)
	if err != nil {
		log.Println(err.Error())
	}
}

// LogAsync 异步记录日志
// 参数:
//   log - 日志条目
func LogAsync(log LogEntry) {
	logChannel <- log
}

// CloseLogging 关闭日志系统
// 等待所有日志处理完成
func CloseLogging() {
	close(logChannel) // 关闭通道以停止 goroutine
	wg.Wait()         // 等待日志处理完成
}
