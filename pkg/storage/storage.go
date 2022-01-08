/*
 * @Author: lishijun10
 * @Email: lishijun1@jd.com
 * @Date: 2021-09-17 16:39:12
 * @LastEditTime: 2021-09-29 11:10:43
 * @LastEditors: lishijun1
 * @Description:
 * @FilePath: /yulei-gateway-controller/pkg/storage/storage.go
 */
package storage

import "github.com/yulei-gateway/yulei-gateway-controller/pkg/resource"

type NodeGroup struct {
	Name string
	Desc string
}

type NodeGroupBind struct {
	// GroupName the node group name
	GroupName string
	//NodeID envoy node id
	NodeID string
}

type Listener struct {
	Name          string
	Address       string
	Port          uint32
	Metadata      string
	NodeGrouoName string
}

type FilterTemplate struct {
	Name     string
	Template string
	Metadata string
}

type ListenerBindFilter struct {
	ListenerName string
	FilterName   string
	// FilterConfig the filter config object ,Notice this is a json string
	FilterConfig string
}

type EndpointDiscoveryConfig struct {
	Name          string
	ServerAddress string
	DiscoveryType string
	Auth          string
}

type HealtyCheckConfig struct {
	CheckType string
	HttpPatch string
	HttpCode  string
}
type Cluster struct {
	Name              string
	EndpointDiscovery *EndpointDiscoveryConfig
	Desc              string
	HealtyCheck       *HealtyCheckConfig
}

type Endpoint struct {
	ClusterName string
	Address     string
	Port        uint32
}

type RouteConfig struct {
	Name         string
	ListenerName string
}

type RouteBindFilter struct {
	RouteConfigName string
	FilterName      string
	//Notice this is a json config
	FilterConfig string
}

type HeaderRouteConfig struct {
	MatchType   string
	HeaderKey   string
	HeaderValue string
}

type RouteClusterBind struct {
	RouteConfigName string
	ClusterName     string
	Path            string
	PathType        string
	HeaderConfig    *HeaderRouteConfig
}

type Storage interface {
	GetEnvoyConfig(nodeID string) (*resource.EnvoyConfig, error)
	GetChangeMsgChan() chan string
	GetNodeIDs() ([]string, error)
}

//ControllerStorage the Controller use data storage interface
type ControllerStorage interface {
	ClusterStorage
	NodeGroupStorage
	NodeGroupBindNodeStorage
	ListenerStorage
	RouteStorage
	FilterStorage
	ListenerBindFilterStorage
	EndpointStorage
	RouteBindFilterStorage
	RouteBindClusterStorage
}

//ClusterStorage the app cluster storage interface
type ClusterStorage interface {
	CreateCluster(cluster *Cluster) (*Cluster, error)
	UpdateCluster(cluster *Cluster) (*Cluster, error)
	ListClusters() ([]Cluster, error)
	DeleteCluster(name string) error
}

// NodeGroupStorage the envoy node group storage interface
type NodeGroupStorage interface {
	CreateNodeGroup(nodeGroup *NodeGroup) (*NodeGroup, error)
	UpdateNodeGroup(nodeGroup *NodeGroup) (*NodeGroup, error)
	ListNodeGroup() ([]NodeGroup, error)
	DeleteNodeGroup(name string) error
	ListNodeGroupNodes(nodeGroupName string) ([]string, error)
}

type NodeGroupBindNodeStorage interface {
	BindNode(nodeGroupName string, nodeName string) error
	UnBindNode(nodeGroupName string, nodeName string) error
}

//ListenerStorage the envoy node listener config storage
type ListenerStorage interface {
	CreateListener(listener *Listener) (*Listener, error)
	UpdateListener(listener *Listener) (*Listener, error)
	ListListener() ([]Listener, error)
	DeleteListener(name string) error
}

//RouteStorage the route config
type RouteStorage interface {
	CreateRoute(route *RouteConfig) (*RouteConfig, error)
	UpdateRoute(route *RouteConfig) (*RouteConfig, error)
	ListRoutes() ([]RouteConfig, error)
	DeleteRoute(name string) error
}

type FilterStorage interface {
	CreateFilterTemplate(filterTemplate *FilterTemplate) (*FilterTemplate, error)
	UpdateFilterTemplate(filterTemplate *FilterTemplate) (*FilterTemplate, error)
	DeleteFilterTemplate(name string) error
	ListFilterTemplate() ([]FilterTemplate, error)
}

type ListenerBindFilterStorage interface {
	BindListenerFilter(listenerBindFilter *ListenerBindFilter) error
	UnBindListenerFilter(listenerBindFilter *ListenerBindFilter) error
	//TODO this need modify
	ListListenerBindFilter(listenerName string) ([]FilterTemplate, error)
	ListFilterUseByListener(filterName string) ([]Listener, error)
}

type EndpointStorage interface {
	CreateEndpoint(endpoint *Endpoint) (*Endpoint, error)
	UpdateEndpoint(endpoint *Endpoint) (*Endpoint, error)
	ListEndpint(clusterName string) ([]Endpoint, error)
	DeleteEndpoint(clusterName string, address string, port uint32) error
}

type RouteBindFilterStorage interface {
	BindRouteFilter(routeBindFilter *RouteBindFilter) error
	UnBindRouteFilter(routeBindFilter *RouteBindFilter) error
}

type RouteBindClusterStorage interface {
	BindRouteCluster(bindRouteCluster *RouteClusterBind) error
	UnBindRouteCluster(bindRouteCluster *RouteClusterBind) error
	ListRouteClusters(routeName string) ([]Cluster, error)
}
