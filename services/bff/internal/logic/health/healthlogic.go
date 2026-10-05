// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package health

import (
	"context"

	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Health 探活。
//
// ============================================================
// 刻意**不检查下游**
// ============================================================
//
// 这是本项目里最容易做错的一个设计选择。探活的目的是回答
// **"这个进程还能不能服务请求"**,而不是"整条依赖链是否健康"。
//
// 若在这里去 ping 所有下游:
//
//	某个下游抖动 → BFF 的探活失败 → 编排层重启 BFF
//	→ BFF 重启期间**所有**请求失败(包括不依赖那个下游的)
//	→ 雪崩被放大
//
// 单体的 GlobalRateLimit 里有一条同样的注释("healthz 不能被限流
// 打死 —— 否则依赖抖动时高频请求吃光令牌,编排层把健康实例判死
// 并重启,雪崩被放大成故障"),说的是同一个道理。
//
// 故这里恒返回 ok。
//
// ============================================================
// "下游挂了怎么办"是另一个问题,由别处回答
// ============================================================
//
//	请求级的判断   → 各 logic 的 ResultInfra 分支 → 503
//	运维级的观测   → Prometheus 的 grpc client 指标 + 告警
//	                (docker/prometheus.yml 里按服务分别配 target)
//
// 用"探活失败"来表达"下游不可用"会把两个问题混在一起,
// 代价是上面那条放大链路。
//
// ============================================================
// 响应形状对齐单体
// ============================================================
//
// 单体的 routes.go 里 healthz 是内联闭包,返回
// `{"status":"ok"}`。注意**它是裸 JSON,没有 {code,message,data}
// 那层信封**(那个闭包直接 c.JSON,没走 utils.Success)。
//
// 而 BFF 这边因为 handler 模板统一封信封,返回的是:
//
//	{"code":200, "message":"Success", "data":{"status":"ok"}}
//
// **这是一处可见差异。** 若探针(或运维脚本)按 `$.status` 取值,
// 它会读到空(BFF 的在 `$.data.status`)。
//
// 要不要对齐:探针配置改一行比让 59 个接口的信封不一致更划算,
// 故保持 BFF 的统一信封。**但要记得改探针的取值路径。**
func (l *HealthLogic) Health(req *types.Empty) (*types.HealthResp, error) {
	return &types.HealthResp{Status: "ok"}, nil
}
