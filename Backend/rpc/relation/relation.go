package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"IMM/rpc/relation/internal/config"
	friendserviceServer "IMM/rpc/relation/internal/server/friendservice"
	groupmemberserviceServer "IMM/rpc/relation/internal/server/groupmemberservice"
	groupserviceServer "IMM/rpc/relation/internal/server/groupservice"
	"IMM/rpc/relation/internal/svc"
	"IMM/rpc/relation/relation"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/relation.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	startConsumers(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		relation.RegisterFriendServiceServer(grpcServer, friendserviceServer.NewFriendServiceServer(ctx))
		relation.RegisterGroupServiceServer(grpcServer, groupserviceServer.NewGroupServiceServer(ctx))
		relation.RegisterGroupMemberServiceServer(grpcServer, groupmemberserviceServer.NewGroupMemberServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

func startConsumers(ctx *svc.ServiceContext) {
	// 消费者1：处理好友申请通知，自动启动协程监听
	err := ctx.RabbitMQ.Consume(
		"im.relation.friend.apply_queue", // 队列名
		"im.relation.friend.apply",       // 绑定的路由键
		func(ctx context.Context, msg []byte) error {
			// 在这里写具体的消费逻辑，比如调用推送服务、写入未读消息表
			logx.Infof("处理好友申请通知: %s", string(msg))
			// 可调用 svcCtx 中的其他依赖，如 DB、Redis
			return nil
		},
	)
	if err != nil {
		log.Fatalf("启动好友申请消费者失败: %v", err)
	}
}
