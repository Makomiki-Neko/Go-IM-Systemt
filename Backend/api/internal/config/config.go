// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret     string
		AccessExpire     int64
		RefreshThreshold int64
	}
	UserRPC     zrpc.RpcClientConf
	RelationRPC zrpc.RpcClientConf
	ChatRPC     zrpc.RpcClientConf
	SeaweedFS   struct {
		Master    string   `json:",env=SEAWEEDFS_MASTER"` // 服务主机 "http://localhost:9333"
		Filers    []string `json:",optional"`             // 文件系统服务器，若为空则直接访问主机 ["http://filer1:8888", "http://filer2:8888"]
		ChunkSize int64    `json:",default=0"`            // 文件分块大小 表示使用库默认
		Timeout   int64    `json:",default=30"`           // 超时
	}
}
