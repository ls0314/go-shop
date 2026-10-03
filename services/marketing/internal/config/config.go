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

	// Tasks 后台任务(券闸门对账)。周期 <= 0 = 不启动。
	Tasks TaskConfig
}

// TaskConfig 后台定时任务。
// 取值写 Go 的时长字面量(如 5m),由 conf 的 mapping 直接解析成 time.Duration。
type TaskConfig struct {
	// CouponInterval 券闸门对账周期:把 coupon:stock:* / coupon:limit:* 收敛到 DB 口径
	CouponInterval time.Duration `json:",default=5m"`
}
