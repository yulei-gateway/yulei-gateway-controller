package driver

import "testing"

func Test_database(t *testing.T) {
	var mysqlTest = NewDatabaseStorage("mysql", "127.0.0.1", "root", "123456", "yulei_test", 3306)
	var clusters = []Cluster{
		{
			Name: "vm",
		}, {
			Name: "vpc",
		},
		{
			Name: "disk",
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
				Name:   "test",
				Prefix: "/",
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
