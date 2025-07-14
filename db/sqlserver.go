package db

import (
	"easyms-es/config"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	BasicDB    *gorm.DB
	AnalysisDB *gorm.DB
	DataDB     *gorm.DB
	ManageDB   *gorm.DB
	JobDB      *gorm.DB
)

func InitSqlServerDB(dbName string) *gorm.DB {
	var (
		user     = config.GetSyncConfig("", "common."+dbName+".user")
		password = config.GetSyncConfig("", "common."+dbName+".password")
		host     = config.GetSyncConfig("", "common."+dbName+".server")
		port     = config.GetSyncConfig("", "common."+dbName+".port")
		database = config.GetSyncConfig("", "common."+dbName+".database")
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	)

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", user, password, host, port, database)
	db, err := gorm.Open(sqlserver.Open(connString), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("failed to connect basic database")
	}
	return db
}
