package resource

import (
	"strconv"
	"strings"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	envoyresource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	_ "github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type EnvoyConfig struct {
	Name string `yaml:"name"`
	Spec `yaml:"spec"`
}

type isRouteMatchPathSpecifier interface {
	isRouteMatch_PathSpecifier()
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

type RoutePathType int
type HeaderMatchType int

const (
	Prefix RoutePathType = iota
	Path
	Regex

	Contains HeaderMatchType = iota
	SuffixMatch
	PrefixMatch
	PresentMatch
	RangeMatch
	SafeRegexMatch
	ExactMatch
)

type Route struct {
	Name      string               `yaml:"name"`
	PathType  RoutePathType        `yaml:"pathType"`
	PathValue string               `yaml:"pathValue"`
	Headers   []HeaderRoute        `yaml:"headers"`
	Clusters  []RouteWeightCluster `yaml:"clusters"`
}

type RouteWeightCluster struct {
	Name   string `yaml:"name"`
	Weight int    `yaml:weight`
}

type HeaderRoute struct {
	HeaderName        string          `yaml:"headerName"`
	HeaderMatcherType HeaderMatchType `yaml:"headerMatchType"`
	HeaderValue       string          `yaml:"headerValue"`
	InvertMatch       bool            `yaml:"invertMatch"`
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

func (e *EnvoyConfig) BuildRoutes() ([]route.RouteConfiguration, error) {
	var result []route.RouteConfiguration
	for _, listener := range e.Listeners {
		var rts []*route.Route
		for _, routeItem := range listener.Routes {
			var listenerRoute = &route.Route{}
			var routeMatch = &route.RouteMatch{}

			switch routeItem.PathType {
			case Prefix:
				routeMatch.PathSpecifier = &route.RouteMatch_Prefix{
					Prefix: routeItem.PathValue,
				}
			case Path:
				routeMatch.PathSpecifier = &route.RouteMatch_Path{
					Path: routeItem.PathValue,
				}
			case Regex:
				routeMatch.PathSpecifier = &route.RouteMatch_SafeRegex{
					SafeRegex: &envoy_type_matcher_v3.RegexMatcher{
						EngineType: &envoy_type_matcher_v3.RegexMatcher_GoogleRe2{
							GoogleRe2: &envoy_type_matcher_v3.RegexMatcher_GoogleRE2{},
						},
						Regex: routeItem.PathValue,
					},
				}
			}

			if len(routeItem.Headers) > 0 {
				var headerMatchers []*route.HeaderMatcher
				for _, headerRouteItem := range routeItem.Headers {
					routeHeaderMatcher := &route.HeaderMatcher{
						Name: headerRouteItem.HeaderName,
					}
					switch headerRouteItem.HeaderMatcherType {
					case ExactMatch:
						routeHeaderMatcher.HeaderMatchSpecifier = &route.HeaderMatcher_ExactMatch{
							ExactMatch: headerRouteItem.HeaderValue,
						}
					case Contains:
						routeHeaderMatcher.HeaderMatchSpecifier = &route.HeaderMatcher_ContainsMatch{
							ContainsMatch: headerRouteItem.HeaderValue,
						}
					case PrefixMatch:
						routeHeaderMatcher.HeaderMatchSpecifier = &route.HeaderMatcher_PrefixMatch{
							PrefixMatch: headerRouteItem.HeaderValue,
						}
					case SuffixMatch:
						routeHeaderMatcher.HeaderMatchSpecifier = &route.HeaderMatcher_SuffixMatch{
							SuffixMatch: headerRouteItem.HeaderValue,
						}
					case SafeRegexMatch:
						routeHeaderMatcher.HeaderMatchSpecifier = &route.HeaderMatcher_SafeRegexMatch{
							SafeRegexMatch: &envoy_type_matcher_v3.RegexMatcher{
								EngineType: &envoy_type_matcher_v3.RegexMatcher_GoogleRe2{
									GoogleRe2: &envoy_type_matcher_v3.RegexMatcher_GoogleRE2{},
								},
								Regex: headerRouteItem.HeaderValue,
							},
						}
					case PresentMatch:
						routeHeaderMatcher.HeaderMatchSpecifier = &route.HeaderMatcher_PresentMatch{
							PresentMatch: true,
						}
					case RangeMatch:
						if headerRouteItem.HeaderValue != "" {
							var rangeInfo = strings.Split(headerRouteItem.HeaderValue, ",")
							if len(rangeInfo) != 2 {
								break
							}
							start, err := strconv.Atoi(rangeInfo[0])
							if err != nil {
								break
							}
							end, err := strconv.Atoi(rangeInfo[1])
							if err != nil {
								break
							}
							routeHeaderMatcher.HeaderMatchSpecifier = &route.HeaderMatcher_RangeMatch{
								RangeMatch: &envoy_type_v3.Int64Range{
									Start: int64(start),
									End:   int64(end),
								},
							}
						}

					}
					routeHeaderMatcher.InvertMatch = headerRouteItem.InvertMatch
					headerMatchers = append(headerMatchers, routeHeaderMatcher)
				}
				routeMatch.Headers = headerMatchers
			}
			listenerRoute.Match = routeMatch
			listenerRoute.Action = &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: routeItem.ClusterName,
					},
				},
			}
			rts = append(rts, listenerRoute)
		}
	}

	return nil, nil
}

func (e *EnvoyConfig) BuildRoute(listenerName string) (*route.RouteConfiguration, error) {
	for _, listener := range e.Listeners {
		if listener.Name == listenerName {

		}
	}
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
