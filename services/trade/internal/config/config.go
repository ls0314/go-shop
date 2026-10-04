package config

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

// DownstreamConfig 下游服务的服务发现键。
type DownstreamConfig struct {
	// EtcdKey 服务发现键,默认与服务的 Name 一致
	EtcdKey string `json:",default=product-service"`
}

// TaskConfig 后台定时任务。
type TaskConfig struct {
	// OrderScanInterval 订单超时扫描周期。
	OrderScanInterval time.Duration `json:",default=1m"`
	// OutboxInterval outbox 投递器轮询周期。
	//
	// 延迟取消的及时性要求是分钟级(支付时限 15 分钟),
	// 故秒级轮询绰绰有余。填 0s = 不启动(仅调试用,生产不可)。
	OutboxInterval time.Duration `json:",default=1s"`
}

// RabbitMQConfig 订单域的 RabbitMQ 连接配置。
//
// 只用于**投递**延迟取消消息。消费端尚未实现(超时取消当前靠定时扫描),
// 故这里只需要一条可写连接。
type RabbitMQConfig struct {
	// Host 空 = 不启用 MQ。此时投递器不启动,超时取消只剩定时扫描一重保险
	Host     string `json:",optional"`
	Port     int    `json:",default=5672"`
	User     string `json:",default=guest"`
	Password string `json:",default=guest"`
	// Vhost 虚拟主机,默认 "/"
	Vhost string `json:",default=/"`
}

// URL 拼出 amqp 连接串。
//
// 不用 fmt.Sprintf 直接拼 Host:Vhost 与 Password 都可能含
// URL 保留字符(@ / :),不转义会让连接串被解析成别的意思。
func (c RabbitMQConfig) URL() string {
	u := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.Vhost,
	}
	return u.String()
}

// Config trade-service 配置。
type Config struct {
	zrpc.RpcServerConf
	DataSource string
	// WorkerId 雪花算法的工作节点 ID(0~1023)。
	WorkerId int64 `json:",default=1"`
	// Product 商品域(锁库存 + 购物车回填的 SKU 快照)。
	Product DownstreamConfig
	// Marketing 券域(核销 + 取消时退还)。
	Marketing DownstreamConfig `json:",optional"`
	// RabbitMQ 延迟取消消息的投递目标(可选)
	RabbitMQ RabbitMQConfig `json:",optional"`
	Task     TaskConfig
}
