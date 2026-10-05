// Code scaffolded by goctl. Safe to edit.
// goctl {{.version}}

package main

import (
	"flag"
	"fmt"

	"demo-shop/pkg/auth"
	{{.importPackages}}
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 验签器在启动时加载一次。
	//
	// **加载失败直接 panic**:验签不可用等于全站 401,带病启动只会让
	// 错误更晚、更难定位。这与单体 middleware.InitJWT 的做法一致。
	//
	// 只加载公钥 —— 签发需私钥,而私钥只在 user-service。
	// BFF 能验签但签不出来,这是刻意的。
	verifier, err := auth.NewVerifierFromFile(c.JwtPublicKeyPath)
	if err != nil {
		panic("加载 JWT 公钥失败: " + err.Error())
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 传 verifier 而不是让 svc 自己加载:公钥路径配错应当在**启动时**
	// 暴露,而不是变成运行时某个请求的 401。
	ctx := svc.NewServiceContext(c, verifier)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
