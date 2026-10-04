package config

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

// Config marketing-service 配置。
//
// 与 product-service 的差异:本服务**没有 ES** —— 券域不需要搜索索引。
// 有效期的展示态由服务端按当前时间推导,也不依赖外部时钟源。
type Config struct {
	zrpc.RpcServerConf
	DataSource string

	// MetricsPort Prometheus 指标端点的独立监听端口(DS-A-26 §1.4)。
	//
	// **与 ListenOn 是两个端口,不能合并** —— go-zero 的 zrpc 是纯 gRPC
	// 服务(RpcServerConf 里没有任何 HTTP 字段),同一个端口无法同时
	// 服务 gRPC 与 HTTP 的 /metrics。
	//
	// §1.4 的端口分配表列的就是**专用 metrics 端口**(单体那一行写
	// 9002,而 9002 从来不是它的 RPC 端口,是纯 metrics 端口 ——
	// 表里"服务"列与"端口号"混着写了)。按表的编号:
	//
	//	9002 单体 metrics(过渡期)
	//	9004 user-service / 9005 product-service
	//	9006 **本服务(marketing)** / 9007 trade-service
	//
	// 那本服务的 RPC 用什么号?表里没留 —— 故沿用现状 9009
	// (9004~9007 已被上表占用,9008 留给 BFF)。
	MetricsPort string `json:",default=9006"`

	// Tasks 后台任务(券闸门对账)。周期 <= 0 = 不启动。
	Tasks TaskConfig
}

// TaskConfig 后台定时任务。
// 取值写 Go 的时长字面量(如 5m),由 conf 的 mapping 直接解析成 time.Duration。
type TaskConfig struct {
	// CouponInterval 券闸门对账周期:把 coupon:stock:* / coupon:limit:* 收敛到 DB 口径
	CouponInterval time.Duration `json:",default=5m"`
}
