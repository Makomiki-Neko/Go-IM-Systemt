// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"
	"net/http"

	"IMM/api/internal/config"
	"IMM/api/internal/handler"
	"IMM/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/user-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithUnauthorizedCallback(JwtAuthFailed)) // 加上鉴权失败回调函数
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// 鉴权失败 自定义消息
func JwtAuthFailed(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"code": 401, "msg": "认证失败"}`))
}
