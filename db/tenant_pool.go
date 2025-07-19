package db

import (
	"easyms-es/config"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"sync"
)

var TenantPoolInstance *TenantPool

type TenantPool struct {
	pools map[string]*gorm.DB
	lock  sync.Mutex
}

type TableConfig struct {
	name     string
	database string
}

func NewTenantPool() *TenantPool {
	pool := &TenantPool{
		pools: make(map[string]*gorm.DB),
	}
	dbConfig := config.GetDBConfig()
	for _, tenant := range dbConfig.Tenants {
		connString := config.GetDBConnString(tenant.Database)
		db, err := gorm.Open(sqlserver.Open(connString), config.NewGormConfig())
		if err != nil {
			panic("failed to connect database")
		}
		pool.pools[tenant.Id] = db
	}
	return pool
}

func (p *TenantPool) GetDB(tenant string) (*gorm.DB, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if db, ok := p.pools[tenant]; ok {
		return db, true
	} else {
		return nil, false
	}
}

func (p *TenantPool) GetTable(tableName string) *gorm.DB {
	dbName, ok := config.GetDBNameFromTable(tableName)
	if !ok {
		return p.pools["default"].Table(tableName)
	}
	db, ok := p.GetDB(dbName)
	if !ok {
		return p.pools["default"].Table(tableName)
	}
	return db.Table(tableName)
}
