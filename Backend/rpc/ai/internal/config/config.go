package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	LLM struct {
		Api   string
		Key   string
		Model string
	}
}
