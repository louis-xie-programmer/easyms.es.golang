package config

import (
	"log"
	"os"
	"time"
)

// LogInfo 初始化日志配置
func LogInfo(path string) {
	file := "./logs/" + path + "/" + time.Now().Format("2024-01-01") + ".log"
	logFile, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(logFile)
}

// Init 读取初始化配置文件
func Init(path string) error {
	//初始化配置
	//if err := Config(path); err != nil {
	//	return err
	//}

	//初始化日志
	LogInfo(path)
	return nil
}

// Config viper解析配置文件
//func Config(path string) error {
//	viper.AddConfigPath("./conf/" + path)
//	viper.SetConfigName("config")
//	viper.SetConfigType("yaml")
//	if err := viper.ReadInConfig(); err != nil {
//		return err
//	}
//	return nil
//}
