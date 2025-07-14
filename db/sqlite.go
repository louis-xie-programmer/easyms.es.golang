package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// SqliteDBManager DBManager 结构体，负责管理数据库连接
type SqliteDBManager struct {
	memoryDB *sql.DB
	diskDB   *sql.DB
}

var EasySqliteDBManager *SqliteDBManager

// InitSqliteDBManager 初始化连接
func InitSqliteDBManager(diskDBPath string) error {
	// 初始化内存数据库（一级缓存）
	memoryDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return fmt.Errorf("failed to open memory database: %v", err)
	}

	// 初始化磁盘数据库（二级缓存）
	diskDB, err := sql.Open("sqlite3", diskDBPath)
	if err != nil {
		return fmt.Errorf("failed to open disk database: %v", err)
	}

	// 测试连接是否正常
	if err := diskDB.Ping(); err != nil {
		return err
	}

	EasySqliteDBManager = &SqliteDBManager{
		memoryDB: memoryDB,
		diskDB:   diskDB,
	}

	return nil
}

// SyncSchema 同步数据库结构
func (manager *SqliteDBManager) SyncSchema() error {
	// 查询磁盘数据库中的表结构信息
	rows, err := manager.diskDB.Query(`SELECT sql FROM sqlite_master WHERE type='table'`)
	if err != nil {
		return fmt.Errorf("failed to query table schema from disk database: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}(rows)

	for rows.Next() {
		var createTableSQL string
		if err := rows.Scan(&createTableSQL); err != nil {
			return fmt.Errorf("failed to scan table schema: %v", err)
		}

		// 在内存数据库中执行创建表的 SQL 语句
		_, err = manager.memoryDB.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("failed to create table in memory database: %v", err)
		}
	}

	return nil
}

// Close 关闭数据库连接
func (manager *SqliteDBManager) Close() error {
	if err := manager.memoryDB.Close(); err != nil {
		return err
	}
	if err := manager.diskDB.Close(); err != nil {
		return err
	}
	return nil
}

// Query 查询数据
func (manager *SqliteDBManager) Query(tableName string, args ...interface{}) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	// 从内存数据中
	rows, err := manager.memoryDB.Query(query, args...)
	if err == nil && rows.Next() {
		return rows, nil // 如果内存命中，直接返回
	}

	// 2. 如果内存数据库未命中，则从磁盘数据库查询
	rows, err = manager.diskDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query disk database: %v", err)
	}

	// 3. 将查询结果缓存到内存数据库（只缓存结果集）
	for rows.Next() {
		var insertQuery string
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))
		for i := range values {
			valuePointers[i] = &values[i]
		}

		if err := rows.Scan(valuePointers...); err != nil {
			return nil, err
		}

		// 动态生成插入语句，插入到内存数据库中
		insertQuery = generateInsertQuery(tableName, columns)
		_, err = manager.memoryDB.Exec(insertQuery, values...)
		if err != nil {
			return nil, fmt.Errorf("failed to cache data in memory database: %v", err)
		}

	}

	return rows, nil
}

// generateInsertQuery 生成插入语句
func generateInsertQuery(tableName string, columns []string) string {
	// 生成插入语句（这里需要根据实际情况修改）
	insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %s (", tableName)
	for i, col := range columns {
		if i > 0 {
			insertQuery += ", "
		}
		insertQuery += col
	}
	insertQuery += ") VALUES ("
	for i := range columns {
		if i > 0 {
			insertQuery += ", "
		}
		insertQuery += "?"
	}
	insertQuery += ")"
	return insertQuery
}

// Insert 插入数据
func (manager *SqliteDBManager) Insert(query string, args ...interface{}) (int64, error) {
	stmt, err := manager.diskDB.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert: %v", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("failed to close stmt: %v", err)
		}
	}(stmt)

	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %v", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %v", err)
	}

	return lastInsertID, nil
}

// Update 更新数据
func (manager *SqliteDBManager) Update(query string, args ...interface{}) (int64, error) {
	stmt, err := manager.diskDB.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare update: %v", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("failed to close stmt: %v", err)
		}
	}(stmt)

	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute update: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve rows affected: %v", err)
	}

	return rowsAffected, nil
}

// Delete 删除数据
func (manager *SqliteDBManager) Delete(query string, args ...interface{}) (int64, error) {
	stmt, err := manager.diskDB.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare delete: %v", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Printf("failed to close stmt: %v", err)
		}
	}(stmt)

	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute delete: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve rows affected: %v", err)
	}

	return rowsAffected, nil
}
