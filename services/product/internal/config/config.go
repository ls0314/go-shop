package config

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

// ESConfig 商品搜索(Elasticsearch)。
// Addresses 留空表示不启用 ES,商品搜索自动降级为直查数据库。
type ESConfig struct {
	Addresses []string `json:",optional"`
}

// TaskConfig 后台定时任务(DS-A-26 §3:ES 对账与库存闸门对账归 product-service)。
//
// 取值写 Go 的时长字面量(如 5m / 12h),由 conf 的 mapping 直接解析成 time.Duration。
// **任一周期为 0 或负数 = 不启动该任务** —— 本地只想跑 RPC 不想跑任务时用得上。
type TaskConfig struct {
	// ESIncInterval ES 增量对账周期
	ESIncInterval time.Duration `json:",default=5m"`
	// ESFullInterval ES 全量对账周期(清孤儿文档)
	ESFullInterval time.Duration `json:",default=12h"`
	// StockInterval 库存闸门对账周期
	StockInterval time.Duration `json:",default=5m"`
}

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	ES         ESConfig `json:",optional"`
	Task       TaskConfig
}
