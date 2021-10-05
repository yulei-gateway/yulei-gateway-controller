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
