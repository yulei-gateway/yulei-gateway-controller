package driver

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yulei-gateway/yulei-gateway-controller/pkg/storage"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var noticeChan chan string
var once = &sync.Once{}

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
	once.Do(
		func() {
			noticeChan = make(chan string)
		},
	)

	var client = &DatabaseStorage{
		DatabaseType: dbType,
		DatabaseHost: host,
		DatabasePort: port,
		Username:     username,
		Password:     password,
		DatabaseName: dbName,
		NoticeChan:   noticeChan,
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
	return c.db.AutoMigrate(&Cluster{}, &Endpoint{}, &Node{}, &Listener{}, &Route{}, &HeaderRoute{})
}

func (c *DatabaseStorage) getEnvoyConfig(nodeID string) (*storage.EnvoyConfig, error) {
	return nil, nil
}
func (c *DatabaseStorage) getChangeMsgChan() chan string {
	return c.NoticeChan
}

// Cluster the end point clusters
type Cluster struct {
	gorm.Model
	Name      string     `json:"name" gorm:"index"`
	Endpoints []Endpoint `json:"endpoints"  gorm:"many2many:cluster_endpoints;"`
	Routes    []Route    `json:"routes"  gorm:"many2many:route_clusters;"`
}

type Endpoint struct {
	gorm.Model
	Address  string    `json:"address" gorm:"index:idx_endpoint,priority:2"`
	Port     int       `json:"port" gorm:"index:idx_endpoint,priority:3"`
	Clusters []Cluster `json:"clusters" gorm:"many2many:cluster_endpoints;"`
}

type Node struct {
	gorm.Model
	EnvoyNodeID string     `json:"nodeID"`
	Listeners   []Listener `json:"listeners" gorm:"foreignKey:NodeID"`
}

type Listener struct {
	gorm.Model
	Name    string  `yaml:"name"`
	Address string  `yaml:"address"`
	Port    uint32  `yaml:"port"`
	Routes  []Route `yaml:"routes" json:"routes" gorm:"foreignKey:ListenerID;"`
	NodeID  int     `json:"nodeID" gorm:"index:idx_node_id;"`
}

type Route struct {
	gorm.Model
	Name       string        `yaml:"name"`
	Prefix     string        `yaml:"prefix"`
	Headers    []HeaderRoute `json:"headers" gorm:"foreignKey:RouteID"`
	Clusters   []Cluster     `json:"clusters" gorm:"many2many:route_clusters;"`
	ListenerID int           `json:"listenerID"  gorm:"index:idx_listener_id;"`
}

type HeaderRoute struct {
	gorm.Model
	HeaderName  string `json:"headerName"`
	HeaderValue string `json:"headerValue"`
	RouteID     int
}
