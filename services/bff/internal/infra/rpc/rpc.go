// Package rpc 对 gRPC 客户端的一层极薄封装。
package rpc

import (
	"context"
	"errors"
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
		return ResultBiz, errors.New(errorMsg)
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
