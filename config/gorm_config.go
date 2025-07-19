package config

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type DBConfig struct {
	Tenants []TenantsConfig
	Tables  []TableConfig
}

type TenantsConfig struct {
	Id       string
	Database DataBaseSetting
}

type DataBaseSetting struct {
	host     string
	port     int
	user     string
	password string
	database string
}

type TableConfig struct {
	name     string
	database string
}

var dbConfig *DBConfig

// CreateDBConfig 创建数据库配置
func CreateDBConfig(appName string) *DBConfig {
	v := viper.New()
	v.SetConfigFile(fmt.Sprintf("/conf/%s/dbconfig.yaml", appName))
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error loading config file %s: %s", "dbconfig.yaml", err)
	}

	if err := viper.Unmarshal(&dbConfig); err != nil {
		log.Fatalf("Error loading config file %s: %s", "dbconfig.yaml", err)
	}

	return dbConfig
}

// GetDBConnString 获取数据库连接字符串
func GetDBConnString(setting DataBaseSetting) string {
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", setting.user, setting.password, setting.host, setting.port, setting.database)
}

func GetDBConfig() *DBConfig {
	return dbConfig
}

func GetDBNameFromTable(tableName string) (string, bool) {
	for _, table := range dbConfig.Tables {
		if table.name == tableName {
			return table.database, true
		}
	}
	return "", false
}

func NewGormConfig() *gorm.Config {
	return &gorm.Config{
		PrepareStmt: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      logger.Warn,
			}),
	}
}
