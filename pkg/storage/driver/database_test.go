package driver

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yulei-gateway/yulei-gateway-controller/pkg/resource"
)

func Test_GetEnvoyConfig(t *testing.T) {
	var mysqlTest = NewDatabaseStorage("mysql", "127.0.0.1", "root", "123456", "yulei_test", 3306)

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
	var mysqlTest = NewDatabaseStorage("mysql", "127.0.0.1", "root", "123456", "yulei_test", 3306)
	var envoyNode = &Node{}
	//err := mysqlTest.db.Debug().Table("nodes").Select("nodes.*,listeners.*").Joins("left join listeners on nodes.id=listeners.node_id").
	//	First(envoyNode, "envoy_node_id=?", "test").Error
	err := mysqlTest.db.Debug().Table("nodes").Preload("Listeners.Routes.Headers").Preload("Listeners.Routes.Clusters.Endpoints").First(envoyNode, "envoy_node_id=?", "test").Error
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
	var mysqlTest = NewDatabaseStorage("mysql", "127.0.0.1", "root", "123456", "yulei_test", 3306)
	var clusters = []Cluster{
		{
			Name: "vm",
			Endpoints: []Endpoint{
				{
					Address: "127.0.0.1",
					Port:    12345,
				},
				{
					Address: "127.0.0.1",
					Port:    12346,
				},
			},
		}, {
			Name: "vpc",
			Endpoints: []Endpoint{
				{
					Address: "127.0.0.1",
					Port:    12347,
				},
				{
					Address: "127.0.0.1",
					Port:    12348,
				},
			},
		},
		{
			Name: "disk",
			Endpoints: []Endpoint{
				{
					Address: "127.0.0.1",
					Port:    12349,
				},
				{
					Address: "127.0.0.1",
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
	}
	err := mysqlTest.db.Model(&Node{}).Create(node).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	err = mysqlTest.db.Model(&Cluster{}).CreateInBatches(clusters, len(clusters)).Error
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
				ListenerID: int(listeners[0].ID),
				Clusters: []Cluster{
					*vmcluster,
				},
			},
		}
		err = mysqlTest.db.Model(&Route{}).CreateInBatches(routes, len(routes)).Error
		if err != nil {
			t.Fatal(err)
			return
		}

	}

}
