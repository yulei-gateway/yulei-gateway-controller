package xds

import (
	"context"
	"fmt"
	"log"
	"net"

	locallog "github.com/yulei-gateway/yulei-gateway-controller/pkg/log"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	"github.com/yulei-gateway/yulei-gateway-controller/pkg/storage"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimeservice "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"

	xdsv3server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	_ "github.com/fsnotify/fsnotify"
	_ "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	_ "gopkg.in/yaml.v3"
)

type YuLeiXDSServer struct {
	grpcServerOptions []grpc.ServerOption
	xdsV3Server       xdsv3server.Server
	port              uint32
	grpcServer        *grpc.Server
	Storage           storage.Storage
	log               *locallog.LocalLogger
	serverCache       cache.SnapshotCache
}

func (y *YuLeiXDSServer) registerServer() {
	// register services
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(y.grpcServer, y.xdsV3Server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(y.grpcServer, y.xdsV3Server)
	clusterservice.RegisterClusterDiscoveryServiceServer(y.grpcServer, y.xdsV3Server)
	routeservice.RegisterRouteDiscoveryServiceServer(y.grpcServer, y.xdsV3Server)
	listenerservice.RegisterListenerDiscoveryServiceServer(y.grpcServer, y.xdsV3Server)
	secretservice.RegisterSecretDiscoveryServiceServer(y.grpcServer, y.xdsV3Server)
	runtimeservice.RegisterRuntimeDiscoveryServiceServer(y.grpcServer, y.xdsV3Server)
}

//NewYuLeiXDSServer create a new gateway envoy xds server grpcServerOptions: the grpc server options ,xdsV3Server:
// the xds v3 server ,port: the grpc start listen port ,dataStorage: the data storage impl
func NewYuLeiXDSServer(grpcServerOptions []grpc.ServerOption,
	port uint32, dataStorage storage.Storage, log *locallog.LocalLogger) *YuLeiXDSServer {
	if dataStorage == nil {
		panic("data storage not config ,can not start server")
	}
	var yuLeiXDSServer = &YuLeiXDSServer{}
	yuLeiXDSServer.grpcServerOptions = grpcServerOptions
	yuLeiXDSServer.port = port
	yuLeiXDSServer.grpcServer = grpc.NewServer(grpcServerOptions...)
	yuLeiXDSServer.log = log
	//TODO the ads flag need test
	yuLeiXDSServer.serverCache = cache.NewSnapshotCache(false, cache.IDHash{}, yuLeiXDSServer.log)
	nodes, err := dataStorage.GetNodeIDs()
	if err != nil {
		panic(fmt.Sprintf("get nodes from data storage error,msg: %v", err.Error()))
	}
	for _, nodeItem := range nodes {
		yuLeiXDSServer.updateCache(nodeItem)
	}

	log.Infof("the yulei xds server create success ")
	return yuLeiXDSServer
}

func (y *YuLeiXDSServer) processDataSourceChange(ctx context.Context) {
	dataChangeChan := y.Storage.GetChangeMsgChan()
	for {
		select {
		case nodeID := <-dataChangeChan:
			if nodeID != "" {
				y.updateCache(nodeID)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (y *YuLeiXDSServer) updateCache(nodeID string) {
	envoyConfig, err := y.Storage.GetEnvoyConfig(nodeID)
	if err != nil {
		return
	}
	fmt.Println(envoyConfig)

}

//Start  the gateway envoy xds server start method
func (y *YuLeiXDSServer) Start(ctx context.Context) {
	//TODO create callbacks
	y.xdsV3Server = xdsv3server.NewServer(ctx, y.serverCache, nil)
	y.registerServer()
	go y.processDataSourceChange(ctx)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", y.port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("management server listening on %d\n", y.port)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if err = y.grpcServer.Serve(lis); err != nil {
		log.Println(err)
	}
}
