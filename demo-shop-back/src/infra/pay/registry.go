package pay

import (
	"fmt"
	"sync"
)

// ============================================================
// 支付网关注册表 — 全局单例，管理所有 PayGateway 实现
// ============================================================
//
// 设计目的：解耦 Service 层与具体网关实现。
// Service 层通过 pay.Get("wechat") 获取网关，不关心是哪个实现类。
// 新增支付渠道只需实现 PayGateway 接口并 Register，无需修改 Service 层代码。
//
// 线程安全：Register / Get 均加读写锁，支持启动时单线程注册 + 运行时多线程读取。

// GatewayRegistry 支付网关注册表
// 使用 sync.RWMutex 保证并发安全（读多写少场景，读不互斥）
type GatewayRegistry struct {
	mn      sync.RWMutex          // 读写锁
	gateway map[string]PayGateway // key=支付方式标识（mock/wechat/alipay），value=网关实例
}

// globalRegistry 全局唯一的网关注册表实例
// 包级别单例，避免外部直接引用
var globalRegistry = &GatewayRegistry{
	gateway: make(map[string]PayGateway),
}

// Register 注册一个支付网关到全局注册表
// 在 InitGateways() 中调用，启动时一次性注册所有网关
// 同一 Name 的网关重复注册会覆盖旧值（后注册的生效）
func Register(gw PayGateway) {
	globalRegistry.mn.Lock()
	defer globalRegistry.mn.Unlock()
	globalRegistry.gateway[gw.Name()] = gw
}

// Get 根据支付方式获取对应的网关实例
// Service 层通过 pay.Get(payMethod) 动态获取，支持运行时切换
// 参数：
//
//	method - 支付方式标识（model.PayMethodMock / PayMethodWechat / PayMethodAlipay）
//
// 返回值：
//
//	PayGateway - 对应的网关实现，若未注册返回 nil
//	error      - 未注册时返回 "gateway xxx not found" 错误
func Get(method string) (PayGateway, error) {
	globalRegistry.mn.RLock()
	defer globalRegistry.mn.RUnlock()
	gw, ok := globalRegistry.gateway[method]
	if !ok {
		return nil, fmt.Errorf("gateway %s not found", method)
	}
	return gw, nil
}

// InitGateways 初始化所有支付网关（应用启动时调用一次）
// 调用时机：infra.InitInfra() → pay.InitGateways()
//
// 当前仅注册 mock 网关（始终可用），wechat/alipay 待对接：
//
//	if cfg.WechatPay.Enabled {
//	    Register(NewWechatPayGateway(cfg.WechatPay.MchId, cfg.WechatPay.ApiKey, ...))
//	}
//	if cfg.Alipay.Enabled {
//	    Register(NewAlipayGateway(cfg.Alipay.AppId, cfg.Alipay.PrivateKey, ...))
//	}
func InitGateways() {
	Register(NewMockPayGateway())
}
