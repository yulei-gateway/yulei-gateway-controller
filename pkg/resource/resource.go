/*
 * @Author: lishijun10
 * @Email: lishijun1@jd.com
 * @Date: 2021-09-29 11:09:43
 * @LastEditTime: 2021-09-29 11:31:32
 * @LastEditors: lishijun1
 * @Description:
 * @FilePath: /yulei-gateway-controller/pkg/resource/resource.go
 */
package resource

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	_ "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	_ "github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

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

func (e *EnvoyConfig) BuildCluster() (*cluster.Cluster, error) {
	return nil, nil
}

func (e *EnvoyConfig) BuildListener() (*listener.Listener, error) {
	return nil, nil
}

func (e *EnvoyConfig) BuildRoute() (*route.RouteConfiguration, error) {
	return nil, nil
}

func (e *EnvoyConfig) BuildEndpoint() (*endpoint.ClusterLoadAssignment, error) {
	var endpoints []*endpoint.LbEndpoint
	for _, cluster := range e.Clusters {
		for _, endpointItem := range cluster.Endpoints {
			var endpoint = &endpoint.LbEndpoint{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  endpointItem.Address,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: endpointItem.Port,
									},
								},
							},
						},
					},
				},
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	if len(endpoints) > 0 {
		return &endpoint.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpoint.LocalityLbEndpoints{{
				LbEndpoints: endpoints,
			}},
		}, nil
	}

	return nil, nil
}
