// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	RabbitMQ struct {
		Host     string
		Port     int
		Username string
		Password string
		VHost    string
	}
	// JWT 认证密钥，用于验证登录Token
	JwtAuth struct {
		AccessSecret string
	}
	RedisConf struct {
		Host string
		Pass string
		DB   int
	}
	ChatRPC zrpc.RpcClientConf

	SeaweedFS_S3 struct {
		Region    string
		Endpoint  string
		AccessKey string
		SecretKey string
	}
}
