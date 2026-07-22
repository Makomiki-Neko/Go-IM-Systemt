package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB struct {
		DataSource string
	}
	RedisConf redis.RedisConf
	RabbitMQ  struct {
		DSN      string // 连接地址
		Exchange string // 业务默认交换机名
		Vhost    string // 虚拟主机
	}
}
