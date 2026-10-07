// Package rpc 对 gRPC 客户端的一层极薄封装。
package rpc

import (
	"context"
	"fmt"
	"time"

	"demo-shop/services/bff/internal/response"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// CallTimeout 单次下游调用的超时。
const CallTimeout = 2 * time.Second

// CtxWithTimeout 统一的调用超时口径。
func CtxWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, CallTimeout)
}

type Result int

const (
	// ResultOK 成功
	ResultOK Result = iota

	// ResultBiz 业务失败
	ResultBiz

	// ResultInfra 基础设施失败
	ResultInfra
)

func Classify(errorMsg string, grpcErr error) (Result, error) {
	if grpcErr != nil {
		return ResultInfra, grpcErr
	}
	if errorMsg != "" {
		// Biz 分支同理必须包哨兵,不能直接 errors.New(errorMsg)。
		//
		// response.Failure 的业务分支白名单是 response.ErrInvalidParam,
		// 裸 errors.New 认不出来 → 落 default → 回 **500**。而单体对下游
		// 业务错误回的是 utils.Fail(c, 400, msg) —— 即 **400**。裸 error
		// 会让全线 111 条路由的业务错误码与单体不一致(前端看到 500 会
		// 以为服务端挂了,实际是"库存不足"这类用户可理解的原因)。
		//
		// 复用 ErrInvalidParam 而不是新加哨兵:白名单里两者等价,而新哨兵
		// 要同时改 response 包与它的白名单 —— 多一处需要同步的地方。
		// 对外文案不受影响:Failure 用的是 err.Error(),仍是下游给的原文。
		//
		// 这与 WrapUnauthorized 是同一个手法(那个把业务失败标成凭据问题
		// 让 Failure 回 401),顺序也对:调用方若想要 401,会把这里的
		// 结果再过一次 WrapUnauthorized。
		return ResultBiz, fmt.Errorf("%s: %w", errorMsg, response.ErrInvalidParam)
	}
	return ResultOK, nil
}

type unauthorizedError struct{ msg string }

func (e *unauthorizedError) Error() string { return e.msg }
func (e *unauthorizedError) Unwrap() error { return response.ErrUnauthorized }

// WrapUnauthorized 把业务失败标成"凭据无效",让 Failure 回 401。
func WrapUnauthorized(err error) error {
	if err == nil {
		return nil
	}
	return &unauthorizedError{msg: err.Error()}
}

// Connect 建一条到下游的连接。
func Connect(etcdHosts []string, etcdKey string) *grpc.ClientConn {
	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: etcdHosts,
			Key:   etcdKey,
		},
	})
	return client.Conn()
}
