package driver

import (
	"fmt"
	"strings"
	"time"

	"github.com/yulei-gateway/yulei-gateway-controller/pkg/storage"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type DatabaseStorage struct {
	DatabaseName string
	DatabaseType string
	DatabaseHost string
	DatabasePort int
	Username     string
	Password     string
	db           *gorm.DB
	OtherDSN     string
	NoticeChan   chan string
}

func NewDatabaseStorage(dbType, host, username, password, dbName string, port int) *DatabaseStorage {
	var client = &DatabaseStorage{
		DatabaseType: dbType,
		DatabaseHost: host,
		DatabasePort: port,
		Username:     username,
		Password:     password,
		DatabaseName: dbName,
		NoticeChan:   make(chan string),
	}
	err := client.getDBPool()
	if err != nil {
		panic(err)
	}
	err = client.migrate()
	if err != nil {
		panic(err)
	}

	return client
}

func (c *DatabaseStorage) getDBPool() error {
	if c.db != nil {
		return nil
	}
	var db *gorm.DB
	var err error
	switch strings.ToLower(c.DatabaseType) {
	case "mysql":
		if c.OtherDSN == "" {
			c.OtherDSN = "parseTime=True&loc=Local"
		}
		// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&%s",
			c.Username, c.Password, c.DatabaseHost, c.DatabasePort, c.DatabaseName, c.OtherDSN)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
	case "postgres":
		if c.OtherDSN == "" {
			c.OtherDSN = "sslmode=disable TimeZone=Asia/Shanghai"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d %s",
			c.DatabaseHost, c.Username, c.Password, c.DatabaseName, c.DatabasePort, c.OtherDSN)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", c.Username, c.Password, c.DatabaseHost, c.DatabasePort, c.DatabaseName)
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}

	case "clickhouse":
		if c.OtherDSN == "" {
			c.OtherDSN = "read_timeout=10&write_timeout=20"
		}
		dsn := fmt.Sprintf("tcp://%s:%d?database=%s&username=%s&password=%s&%s",
			c.DatabaseHost, c.DatabasePort, c.DatabaseName, c.Username, c.Password, c.OtherDSN)
		db, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
	default:
		// github.com/mattn/go-sqlite3
		db, err = gorm.Open(sqlite.Open(c.DatabaseName), &gorm.Config{})
		if err != nil {
			return err
		}

	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	c.db = db

	return nil
}
func (c *DatabaseStorage) migrate() error {
	return nil
}

func (c *DatabaseStorage) getEnvoyConfig(nodeID string) (*storage.EnvoyConfig, error) {
	return nil, nil
}
func (c *DatabaseStorage) getChangeMsgChan() chan string {
	return c.NoticeChan
}
