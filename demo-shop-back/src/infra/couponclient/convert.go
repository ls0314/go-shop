package couponclient

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/response"
	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
)

// ============================================================
// proto → 单体 model 的转换。
//
// 为什么保留单体原有的 response 结构体而不是直接把 proto 结构体返给 handler:
// HTTP 响应的 JSON 字段名是对前端的契约,已在用的字段不能因为换实现而改名。
// 单体的 response.* 结构体带有 json tag,proto 结构体带的是 protobuf 的
// json_name,两者并不总是一致 —— 故在此显式搬运。
// ============================================================

// toModelCouponList 模板分页 → 管理端列表响应
func toModelCouponList(resp *v1_marketingv1.ListCouponTemplatesResp) *response.GetCouponListResp {
	list := make([]response.GetCouponList, 0, len(resp.Items))
	for _, it := range resp.Items {
		if it == nil {
			continue
		}
		item := response.GetCouponList{
			TemplateId:      it.TemplateId,
			CouponName:      it.CouponName,
			CouponType:      it.CouponType,
			ThresholdAmount: it.ThresholdAmount,
			DiscountAmount:  it.DiscountAmount,
			TotalCount:      it.TotalCount,
			ReceivedCount:   it.ReceivedCount,
			PerUserLimit:    it.PerUserLimit,
			UsableDays:      it.UsableDays,
			Status:          it.Status,
		}
		if it.StartTime != nil {
			item.StartTime = it.StartTime.AsTime()
		}
		if it.EndTime != nil {
			item.EndTime = it.EndTime.AsTime()
		}
		list = append(list, item)
	}
	return &response.GetCouponListResp{
		List:     list,
		Total:    resp.Total,
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}
}

// toModelUserCouponList 我的卡券分页 → 用户端列表响应
func toModelUserCouponList(resp *v1_marketingv1.ListUserCouponsResp) *response.UserGetCouponListResp {
	list := make([]response.UserGetCouponList, 0, len(resp.Items))
	for _, it := range resp.Items {
		if it == nil {
			continue
		}
		item := response.UserGetCouponList{
			UserCouponId:    it.UserCouponId,
			CouponName:      it.CouponName,
			CouponType:      it.CouponType,
			ThresholdAmount: it.ThresholdAmount,
			DiscountAmount:  it.DiscountAmount,
			Status:          it.Status,
			OrderNo:         it.OrderNo,
			UserId:          it.UserId,
		}
		if it.ExpireAt != nil {
			item.ExpireAt = it.ExpireAt.AsTime()
		}
		if it.UsedAt != nil {
			item.UsedAt = it.UsedAt.AsTime()
		}
		list = append(list, item)
	}
	return &response.UserGetCouponListResp{
		List:     list,
		Total:    resp.Total,
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}
}

// toModelTemplateList 领券中心 → 用户端模板列表响应
func toModelTemplateList(resp *v1_marketingv1.ListCouponTemplatesForUserResp) *response.UserCouponTemplateListResp {
	list := make([]response.UserCouponTemplate, 0, len(resp.Items))
	for _, it := range resp.Items {
		if it == nil || it.Template == nil {
			continue
		}
		tpl := it.Template
		item := response.UserCouponTemplate{
			TemplateId:      tpl.TemplateId,
			CouponName:      tpl.CouponName,
			CouponType:      tpl.CouponType,
			ThresholdAmount: tpl.ThresholdAmount,
			DiscountAmount:  tpl.DiscountAmount,
			PerUserLimit:    tpl.PerUserLimit,
			HeldCount:       it.HeldCount,
			RemainingCount:  it.RemainingCount,
			UsableDays:      tpl.UsableDays,
		}
		if tpl.StartTime != nil {
			item.StartTime = tpl.StartTime.AsTime()
		}
		if tpl.EndTime != nil {
			item.EndTime = tpl.EndTime.AsTime()
		}
		list = append(list, item)
	}
	return &response.UserCouponTemplateListResp{
		List:     list,
		Total:    resp.Total,
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
	}
}

// toModelAvailableList 结算可用券 → 响应(列表已由服务端按实付升序排好)
func toModelAvailableList(resp *v1_marketingv1.ListAvailableCouponsResp) *response.GetAvailableCouponResp {
	list := make([]response.GetAvailableCouponList, 0, len(resp.Items))
	for _, it := range resp.Items {
		if it == nil {
			continue
		}
		list = append(list, response.GetAvailableCouponList{
			UserCouponId:    it.UserCouponId,
			CouponName:      it.CouponName,
			CouponType:      it.CouponType,
			ThresholdAmount: it.ThresholdAmount,
			DiscountAmount:  it.DiscountAmount,
			PayAfter:        it.PayAfter,
		})
	}
	return &response.GetAvailableCouponResp{List: list}
}

// toModelUserCoupon proto 券 → 单体 model(订单链路用)。
// 只填订单侧真正要用的字段:券类型/门槛/优惠用于算账,状态与订单号用于核对。
func toModelUserCoupon(p *v1_marketingv1.UserCoupon) *model.UserCoupon {
	if p == nil {
		return nil
	}
	out := &model.UserCoupon{
		UserCouponId: p.UserCouponId,
		TemplateId:   p.TemplateId,
		UserId:       p.UserId,
		Status:       p.Status,
		OrderNo:      p.OrderNo,
	}
	if p.UsedAt != nil {
		t := p.UsedAt.AsTime()
		out.UsedAt = &t
	}
	if p.ExpireAt != nil {
		out.ExpireAt = p.ExpireAt.AsTime()
	}
	return out
}
