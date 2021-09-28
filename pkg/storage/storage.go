/*
 * @Author: lishijun10
 * @Email: lishijun1@jd.com
 * @Date: 2021-09-17 16:39:12
 * @LastEditTime: 2021-09-17 18:24:47
 * @LastEditors: lishijun1
 * @Description:
 * @FilePath: /yulei-gateway-controller/pkg/storage/storage.go
 */
package storage

type Storage interface {
	GetEnvoyConfig(nodeID string) (*EnvoyConfig, error)
	GetChangeMsgChan() chan string
}

type EnvoyConfig struct {
	Name string `yaml:"name"`
	Spec `yaml:"spec"`
}

type Spec struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
}

type Listener struct {
	Name    string  `yaml:"name"`
	Address string  `yaml:"address"`
	Port    uint32  `yaml:"port"`
	Routes  []Route `yaml:"routes"`
}

type Route struct {
	Name         string        `yaml:"name"`
	Prefix       string        `yaml:"prefix"`
	Headers      []HeaderRoute `yaml:"headers"`
	ClusterNames []string      `yaml:"clusters"`
}

type HeaderRoute struct {
	HeaderName  string `yaml:"headerName"`
	HeaderValue string `yaml:"headerValue"`
}

type Cluster struct {
	Name      string     `yaml:"name"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Address string `yaml:"address"`
	Port    uint32 `yaml:"port"`
}
