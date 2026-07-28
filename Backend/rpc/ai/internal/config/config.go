package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	LLM struct {
		Api         string
		Key         string
		Model       string
		Temperature float32
		TopP        float32
		MaxTokens   int
	}
	DB struct {
		DataSource string
	}
	RabbitMQ struct {
		DSN      string // 连接地址
		Exchange string // 业务默认交换机名
		Vhost    string // 虚拟主机
	}
	RedisConf redis.RedisConf
}
