package server

import "context"

type ConfigServer interface {
	Run(ctx context.Context)
}

type YuLeiControllerServer struct {
	ConfigServer ConfigServer
}

func (s *YuLeiControllerServer) Start(ctx context.Context) {
	s.ConfigServer.Run(ctx)
}
