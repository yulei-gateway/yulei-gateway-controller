package service

import "github.com/yulei-gateway/yulei-gateway-controller/pkg/storage/driver/db"

type NodeGroupService struct {
	databaseStorage *db.DatabaseStorage
}

func NewNodeGroupService(databaseStorage *db.DatabaseStorage) *NodeGroupService {
	nodeGroupService := &NodeGroupService{databaseStorage: databaseStorage}
	return nodeGroupService
}

func (s *NodeGroupService) CreateNodeGroup() error {
	return nil
}

func (s *NodeGroupService) UpdateNodeGroup() error {
	return nil
}

func (s *NodeGroupService) ListNodeGroup() error {
	return nil
}

func (s *NodeGroupService) DeleteNodeGroup() error {
	return nil
}
