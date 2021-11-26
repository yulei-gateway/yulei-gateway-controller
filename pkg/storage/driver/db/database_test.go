package db

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yulei-gateway/yulei-gateway-controller/pkg/resource"
)

// drop table node_clusters,node_listener,cluster_endpoints,header_routes,route_clusters,endpoints,routes,virtual_hosts,route_configs,listeners,clusters,nodes;

func Test_GetEnvoyConfig(t *testing.T) {
	var mysqlTest = NewDatabaseStorage("mysql", "192.168.15.3", "root", "123456", "yulei_test", 3306)

	result, err := mysqlTest.GetEnvoyConfig("test")
	if err != nil {
		t.Fatal(err)
		return
	}
	data, err := json.MarshalIndent(result, "", "	")
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(string(data))
}

func Test_getNodes(t *testing.T) {
	var mysqlTest = NewDatabaseStorage("mysql", "192.168.15.3", "root", "123456", "yulei_test", 3306)
	var envoyNode = &Node{}
	//err := mysqlTest.db.Debug().Table("nodes").Select("nodes.*,listeners.*").Joins("left join listeners on nodes.id=listeners.node_id").
	//	First(envoyNode, "envoy_node_id=?", "test").Error
	err := mysqlTest.db.Debug().Table("nodes").Preload("Listeners.RouteConfig.VirtualHosts.Routes.Headers").Preload("Listeners.RouteConfig.VirtualHosts.Routes.Clusters.Endpoints").First(envoyNode, "envoy_node_id=?", "test").Error
	if err != nil {
		t.Fatal(err)
		return

	}

	data, err := json.MarshalIndent(envoyNode, "", "	")
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(string(data))

}

func Test_database(t *testing.T) {
	var mysqlTest = NewDatabaseStorage("mysql", "192.168.49.1", "root", "123456", "yulei_test", 3306)
	var clusters = []Cluster{
		{
			Name: "vm",
			Endpoints: []Endpoint{
				{
					Address: "192.168.0.247",
					Port:    12345,
				},
				{
					Address: "192.168.0.247",
					Port:    12346,
				},
			},
		}, {
			Name: "vpc",
			Endpoints: []Endpoint{
				{
					Address: "192.168.0.247",
					Port:    12347,
				},
				{
					Address: "192.168.0.247",
					Port:    12348,
				},
			},
		},
		{
			Name: "disk",
			Endpoints: []Endpoint{
				{
					Address: "192.168.0.247",
					Port:    12349,
				},
				{
					Address: "192.168.0.247",
					Port:    12350,
				},
			},
		},
	}
	var node = &Node{
		EnvoyNodeID: "test",
		Listeners: []Listener{
			{
				Name:    "test_80",
				Address: "0.0.0.0",
				Port:    80,
			},
		},
		Clusters: clusters,
	}
	err := mysqlTest.db.Model(&Node{}).Create(node).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	var testClusters []Cluster

	err = mysqlTest.db.Model(&Cluster{}).Find(&testClusters).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	var listeners = []Listener{}
	err = mysqlTest.db.Model(&Listener{}).Find(&listeners).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	var vmcluster = &Cluster{}
	err = mysqlTest.db.Model(&Cluster{}).First(vmcluster, " name=?", "vm").Error
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(listeners) > 0 {
		var routes = []Route{
			{
				Name:      "test",
				PathType:  int(resource.Prefix),
				PathValue: "/",
				Headers: []HeaderRoute{
					{
						HeaderName:  "x-test-service",
						HeaderValue: "vm",
					},
				},
				//ListenerID: int(listeners[0].ID),
				Clusters: []Cluster{
					*vmcluster,
				},
			},
		}
		var visualHosts = []VirtualHost{
			{
				Name:    "vm",
				Domains: "*",
				Routes:  routes,
			},
		}
		var routeConfig = &RouteConfig{
			Name:         "vm",
			ListenerID:   listeners[0].ID,
			VirtualHosts: visualHosts,
		}

		err = mysqlTest.db.Model(&RouteConfig{}).Create(routeConfig).Error
		if err != nil {
			t.Fatal(err)
			return
		}

	}

}
