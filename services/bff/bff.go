// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"net/http"

	"demo-shop/pkg/auth"
	"demo-shop/services/bff/internal/config"
	"demo-shop/services/bff/internal/handler"
	"demo-shop/services/bff/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/bff.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	verifier, err := auth.NewVerifierFromFile(c.JwtPublicKeyPath)
	if err != nil {
		panic("加载 JWT 公钥失败: " + err.Error())
	}

	server := rest.MustNewServer(c.RestConf,
		rest.WithCustomCors(
			// headerFn:在默认头之外追加
			func(header http.Header) {
				header.Set("Access-Control-Allow-Headers",
					"Content-Type, Authorization, X-Requested-With")
			},
			// notAllowedFn:来源不在白名单时调用(nil = 只回 404/204)
			nil,
			c.AllowedOrigins...,
		),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c, verifier)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
