package driver

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yulei-gateway/yulei-gateway-controller/pkg/resource"
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

//GetEnvoyConfig get a envoy node config
func (c *DatabaseStorage) GetEnvoyConfig(nodeID string) (*resource.EnvoyConfig, error) {
	var envoyNode = &Node{}
	//err := mysqlTest.db.Debug().Table("nodes").Select("nodes.*,listeners.*").Joins("left join listeners on nodes.id=listeners.node_id").
	//	First(envoyNode, "envoy_node_id=?", "test").Error
	err := c.db.Debug().Table("nodes").Preload("Listeners.Routes.Headers").Preload("Listeners.Routes.Clusters.Endpoints").First(envoyNode, "envoy_node_id=?", nodeID).Error
	if err != nil {
		return nil, err
	}
	result := &resource.EnvoyConfig{}
	var envoyListeners = envoyNode.Listeners
	var storageListeners []resource.Listener
	var storageClusters []resource.Cluster
	for _, envoyListenerItem := range envoyListeners {
		storageListener := resource.Listener{}
		storageListener.Address = envoyListenerItem.Address
		storageListener.Name = envoyListenerItem.Name
		storageListener.Port = envoyListenerItem.Port
		var routeConfig = envoyListenerItem.RouteConfig
		var virtualHosts []resource.VirtualHost
		for _, virtualHostItem := range routeConfig.VirtualHosts {
			var storageRoutes []resource.Route

			for _, dbRouteItem := range virtualHostItem.Routes {
				storageRoute := resource.Route{}
				storageRoute.Name = dbRouteItem.Name
				storageRoute.PathType = resource.RoutePathType(dbRouteItem.PathType)
				storageRoute.PathValue = dbRouteItem.PathValue
				var storageHeaderRoutes []resource.HeaderRoute
				for _, dbHeaderItem := range dbRouteItem.Headers {
					storageHeaderRoute := resource.HeaderRoute{}
					storageHeaderRoute.HeaderName = dbHeaderItem.HeaderName
					storageHeaderRoute.HeaderValue = dbHeaderItem.HeaderValue
					storageHeaderRoutes = append(storageHeaderRoutes, storageHeaderRoute)
				}
				storageRoute.Headers = storageHeaderRoutes
				var clusterNames []resource.RouteWeightCluster
				for _, dbRouteCluster := range dbRouteItem.Clusters {
					storageCluster := resource.Cluster{}
					storageCluster.Name = dbRouteCluster.Name
					clusterNames = append(clusterNames, resource.RouteWeightCluster{
						ClusterName: dbRouteCluster.Name,
						Weight:      dbRouteCluster.Weight,
					})
					var storageEndpoints []resource.Endpoint
					for _, dbEndpoint := range dbRouteCluster.Endpoints {
						storageEndpoint := resource.Endpoint{}
						storageEndpoint.Address = dbEndpoint.Address
						storageEndpoint.Port = uint32(dbEndpoint.Port)
						storageEndpoints = append(storageEndpoints, storageEndpoint)
					}
					storageCluster.Endpoints = storageEndpoints
					storageClusters = append(storageClusters, storageCluster)
				}
				storageRoute.Clusters = clusterNames
				storageRoutes = append(storageRoutes, storageRoute)
			}
			virtualHosts = append(virtualHosts, resource.VirtualHost{
				Name:    virtualHostItem.Name,
				Routes:  storageRoutes,
				Domains: virtualHostItem.Domain,
			})
		}
		var storageRouteConfig = &resource.RouteConfig{
			Name:         routeConfig.Name,
			VirtualHosts: virtualHosts,
		}
		storageListener.RouteConfig = storageRouteConfig
		storageListeners = append(storageListeners, storageListener)
	}
	result.Name = nodeID
	result.Spec.Clusters = storageClusters
	result.Spec.Listeners = storageListeners
	return result, nil
}

//GetChangeMsgChan get the chan which data change will send notice message
func (c *DatabaseStorage) GetChangeMsgChan() chan string {
	return c.NoticeChan
}

// Cluster the end point clusters
type Cluster struct {
	gorm.Model
	Name      string     `json:"name" gorm:"index"`
	Weight    uint32     `json:"weight" yaml:"weight"`
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
	EnvoyNodeID string     `yaml:"nodeID" json:"nodeID"`
	Listeners   []Listener `yaml:"listeners" json:"listeners" gorm:"foreignKey:NodeID"`
}

type Listener struct {
	gorm.Model
	Name        string      `yaml:"name" json:"name"`
	Address     string      `yaml:"address" json:"address"`
	Port        uint32      `yaml:"port" json:"port"`
	RouteConfig RouteConfig `yaml:"routeConfig" json:"routeConfig" gorm:"foreignKey:ListenerID"`
	NodeID      int         `yaml:"nodeID" json:"nodeID" gorm:"index:idx_node_id;"`
}

type RouteConfig struct {
	gorm.Model
	Name         string        `yaml:"name" json:"name"`
	ListenerID   int           `yaml:"listenerID" json:"listenerID"`
	VirtualHosts []VirtualHost `yaml:"virtualHost" json:"virtualHost" gorm:"foreignKey:RouteConfigID;"`
}

type VirtualHost struct {
	gorm.Model
	Name          string   `yaml:"name" json:"name"`
	Domain        []string `yaml:"domain" json:"domain"`
	RouteConfigID int      `yaml:"routeConfigID" json:"routeConfigID"`
	Routes        []Route  `yaml:"routes" json:"routes" gorm:"foreignKey:ListenerID;"`
}

type Route struct {
	gorm.Model
	Name       string        `yaml:"name"`
	PathType   int           `yaml:"pathType"`
	PathValue  string        `yaml:"pathValue"`
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
