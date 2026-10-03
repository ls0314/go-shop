package server

import (
	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/svc"
)

// CouponServiceServer 券服务的 gRPC 适配层。
// 只做"proto 入参 → logic → proto 出参"的搬运,不含业务规则;
// 各域方法分散在同包的 coupon.go / trade.go 里。
type CouponServiceServer struct {
	svcCtx *svc.ServiceContext
	v1_marketingv1.UnimplementedCouponServiceServer
}

func NewCouponServiceServer(svcCtx *svc.ServiceContext) *CouponServiceServer {
	return &CouponServiceServer{
		svcCtx: svcCtx,
	}
}
