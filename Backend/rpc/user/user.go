package main

import (
	"flag"
	"fmt"

	snowflake "IMM/common/pkg"
	redisListen "IMM/rpc/user/common"
	"IMM/rpc/user/internal/config"
	"IMM/rpc/user/internal/server"
	"IMM/rpc/user/internal/svc"
	"IMM/rpc/user/user"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// 初始化雪花ID生成器
	if err := snowflake.Init(c.Snowflake.MachineID); err != nil {
		panic(fmt.Sprintf("init snowflake failed: %v", err))
	}

	// 启动Redis过期事件监听
	go redisListen.StartRedisExpiredListener(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
