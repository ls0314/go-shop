package server

import (
	"context"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	couponservicelogic "demo-shop/services/marketing/internal/logic/couponservice"
)

// ---- 管理端:模板 ----

func (s *CouponServiceServer) GetCouponTemplate(ctx context.Context, in *v1_marketingv1.GetCouponTemplateReq) (*v1_marketingv1.GetCouponTemplateResp, error) {
	return couponservicelogic.NewGetCouponTemplateLogic(ctx, s.svcCtx).GetCouponTemplate(in)
}

func (s *CouponServiceServer) ListCouponTemplates(ctx context.Context, in *v1_marketingv1.ListCouponTemplatesReq) (*v1_marketingv1.ListCouponTemplatesResp, error) {
	return couponservicelogic.NewListCouponTemplatesLogic(ctx, s.svcCtx).ListCouponTemplates(in)
}

func (s *CouponServiceServer) CreateCouponTemplate(ctx context.Context, in *v1_marketingv1.CreateCouponTemplateReq) (*v1_marketingv1.CreateCouponTemplateResp, error) {
	return couponservicelogic.NewCreateCouponTemplateLogic(ctx, s.svcCtx).CreateCouponTemplate(in)
}

func (s *CouponServiceServer) UpdateCouponTemplate(ctx context.Context, in *v1_marketingv1.UpdateCouponTemplateReq) (*v1_marketingv1.UpdateCouponTemplateResp, error) {
	return couponservicelogic.NewUpdateCouponTemplateLogic(ctx, s.svcCtx).UpdateCouponTemplate(in)
}

func (s *CouponServiceServer) DeleteCouponTemplate(ctx context.Context, in *v1_marketingv1.DeleteCouponTemplateReq) (*v1_marketingv1.DeleteCouponTemplateResp, error) {
	return couponservicelogic.NewDeleteCouponTemplateLogic(ctx, s.svcCtx).DeleteCouponTemplate(in)
}

// ---- 用户端 ----

func (s *CouponServiceServer) ListUserCoupons(ctx context.Context, in *v1_marketingv1.ListUserCouponsReq) (*v1_marketingv1.ListUserCouponsResp, error) {
	return couponservicelogic.NewListUserCouponsLogic(ctx, s.svcCtx).ListUserCoupons(in)
}

func (s *CouponServiceServer) ListCouponTemplatesForUser(ctx context.Context, in *v1_marketingv1.ListCouponTemplatesForUserReq) (*v1_marketingv1.ListCouponTemplatesForUserResp, error) {
	return couponservicelogic.NewListCouponTemplatesForUserLogic(ctx, s.svcCtx).ListCouponTemplatesForUser(in)
}

func (s *CouponServiceServer) ListAvailableCoupons(ctx context.Context, in *v1_marketingv1.ListAvailableCouponsReq) (*v1_marketingv1.ListAvailableCouponsResp, error) {
	return couponservicelogic.NewListAvailableCouponsLogic(ctx, s.svcCtx).ListAvailableCoupons(in)
}

func (s *CouponServiceServer) ReceiveCoupon(ctx context.Context, in *v1_marketingv1.ReceiveCouponReq) (*v1_marketingv1.ReceiveCouponResp, error) {
	return couponservicelogic.NewReceiveCouponLogic(ctx, s.svcCtx).ReceiveCoupon(in)
}

// ---- 供 trade-service 的核销 / 归还 ----

func (s *CouponServiceServer) UseCoupon(ctx context.Context, in *v1_marketingv1.UseCouponReq) (*v1_marketingv1.UseCouponResp, error) {
	return couponservicelogic.NewUseCouponLogic(ctx, s.svcCtx).UseCoupon(in)
}

func (s *CouponServiceServer) ReturnCoupon(ctx context.Context, in *v1_marketingv1.ReturnCouponReq) (*v1_marketingv1.ReturnCouponResp, error) {
	return couponservicelogic.NewReturnCouponLogic(ctx, s.svcCtx).ReturnCoupon(in)
}
