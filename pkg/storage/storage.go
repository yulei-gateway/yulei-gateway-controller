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
	CreateCluster(cluster *resource.Cluster) (*resource.Cluster, error)
	UpdateCluster(cluster *resource.Cluster) (*resource.Cluster, error)
	ListClusters() ([]resource.Cluster, error)
	DeleteCluster(name string) error
}

// NodeGroupStorage the envoy node group storage interface
type NodeGroupStorage interface {
	CreateNodeGroup(nodeGroup *resource.NodeGroup) (*resource.NodeGroup, error)
	UpdateNodeGroup(nodeGroup *resource.NodeGroup) (*resource.NodeGroup, error)
	ListNodeGroup() ([]resource.NodeGroup, error)
	DeleteNodeGroup(name string) error
	ListNodeGroupNodes(nodeGroupName string) ([]string, error)
}

type NodeGroupBindNodeStorage interface {
	BindNode(nodeGroupName string, nodeName string) error
	UnBindNode(nodeGroupName string, nodeName string) error
}

//ListenerStorage the envoy node listener config storage
type ListenerStorage interface {
	CreateListener(listener *resource.Listener) (*resource.Listener, error)
	UpdateListener(listener *resource.Listener) (*resource.Listener, error)
	ListListener() ([]resource.Listener, error)
	DeleteListener(name string) error
}

//RouteStorage the route config
type RouteStorage interface {
	CreateRoute(route *resource.RouteConfig) (*resource.RouteConfig, error)
	UpdateRoute(route *resource.RouteConfig) (*resource.RouteConfig, error)
	ListRoutes() ([]resource.RouteConfig, error)
	DeleteRoute(name string) error
}

type FilterStorage interface {
	CreateFilterTemplate(filterTemplate *resource.FilterTemplate) (*resource.FilterTemplate, error)
	UpdateFilterTemplate(filterTemplate *resource.FilterTemplate) (*resource.FilterTemplate, error)
	DeleteFilterTemplate(name string) error
	ListFilterTemplate() ([]resource.FilterTemplate, error)
}

type ListenerBindFilterStorage interface {
	BindListenerFilter(listenerBindFilter *resource.ListenerBindFilter) error
	UnBindListenerFilter(listenerBindFilter *resource.ListenerBindFilter) error
	//TODO this need modify
	ListListenerBindFilter(listenerName string) ([]resource.FilterTemplate, error)
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
