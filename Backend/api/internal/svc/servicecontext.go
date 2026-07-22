// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"IMM/api/internal/config"
	"IMM/rpc/chat/chatservice"
	"IMM/rpc/relation/client/friendservice"
	userservice "IMM/rpc/user/client"
	"log"
	"net/http"
	"time"

	"github.com/linxGnu/goseaweedfs"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	UserRpc   userservice.UserService
	FriendRPC friendservice.FriendService
	ChatRPC   chatservice.ChatService
	SeaWeed   *goseaweedfs.Seaweed
}

func NewServiceContext(c config.Config) *ServiceContext {
	sw, err := goseaweedfs.NewSeaweed(c.SeaweedFS.Master, c.SeaweedFS.Filers, c.SeaweedFS.ChunkSize, &http.Client{
		Timeout: time.Duration(c.SeaweedFS.Timeout) * time.Second,
		// 可以进一步配置 Transport（连接池等）
	})
	if err != nil {
		log.Fatal("SeaweedFS Connect Failed: ", err)
	}
	return &ServiceContext{
		Config:    c,
		UserRpc:   userservice.NewUserService(zrpc.MustNewClient(c.UserRPC)),
		FriendRPC: friendservice.NewFriendService(zrpc.MustNewClient(c.RelationRPC)),
		SeaWeed:   sw,
		ChatRPC:   chatservice.NewChatService(zrpc.MustNewClient(c.ChatRPC)),
	}
}
