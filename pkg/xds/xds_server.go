package xds

import (
	_ "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	_ "github.com/fsnotify/fsnotify"
	_ "github.com/sirupsen/logrus"
	_ "google.golang.org/grpc"
	_ "gopkg.in/yaml.v3"
)

type YuLeiXDSServer struct {
}
