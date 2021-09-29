package resource

import (
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	envoyresource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	_ "github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type EnvoyConfig struct {
	Name string `yaml:"name"`
	Spec `yaml:"spec"`
}

type CircuitBreakers struct {
	Priority           int32
	MaxConnections     uint32
	MaxPendingRequests uint32
	MaxRequests        uint32
	MaxRetries         uint32
	TrackRemaining     bool
	MaxConnectionPools uint32
	RetryBudget        *CircuitBreakersThresholdsRetryBudget
}

type CircuitBreakersThresholdsRetryBudget struct {
	BudgetPercent       float64
	MinRetryConcurrency uint32
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

func (e *EnvoyConfig) BuildClusters() ([]cluster.Cluster, error) {
	var result []cluster.Cluster
	for _, item := range e.Clusters {
		result = append(result, cluster.Cluster{
			Name:                 item.Name,
			ConnectTimeout:       durationpb.New(5 * time.Second),
			ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
			LbPolicy:             cluster.Cluster_ROUND_ROBIN,
			//LoadAssignment:       makeEndpoint(clusterName, UpstreamHost),
			DnsLookupFamily:  cluster.Cluster_V4_ONLY,
			EdsClusterConfig: makeEDSCluster(),
		})
	}
	return result, nil
}

func makeEDSCluster() *cluster.Cluster_EdsClusterConfig {
	return &cluster.Cluster_EdsClusterConfig{
		EdsConfig: makeConfigSource(),
	}
}
func makeConfigSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = envoyresource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       envoyresource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					//this is the xds server name
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
				},
			}},
		},
	}
	return source
}

func (e *EnvoyConfig) BuildCluster(clusterName string) (*cluster.Cluster, error) {
	clusters, _ := e.BuildClusters()
	for _, item := range clusters {
		if item.Name == clusterName {
			return &item, nil
		}
	}
	return nil, nil
}

func (e *EnvoyConfig) BuildListener() (*listener.Listener, error) {
	return nil, nil
}

func (e *EnvoyConfig) BuildRoute() (*route.RouteConfiguration, error) {
	return nil, nil
}

func (e *EnvoyConfig) BuildEndpoint(clusterName string) (*endpoint.ClusterLoadAssignment, error) {
	clusters, err := e.BuildEndpoints()
	if err != nil {
		return nil, err
	}
	for _, item := range clusters {
		if item.ClusterName == clusterName {
			return &item, nil
		}
	}
	return nil, nil
}

func (e *EnvoyConfig) BuildEndpoints() ([]endpoint.ClusterLoadAssignment, error) {
	for _, cluster := range e.Clusters {
		var endpoints []*endpoint.LbEndpoint
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
		if len(endpoints) > 0 {
			var result []endpoint.ClusterLoadAssignment
			result = append(result, endpoint.ClusterLoadAssignment{
				ClusterName: cluster.Name,
				Endpoints: []*endpoint.LocalityLbEndpoints{{
					LbEndpoints: endpoints,
				}},
			})
			return result, nil
		}
	}
	return nil, nil
}
