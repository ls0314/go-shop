package converter

import (
	"time"

	v1_marketingv1 "demo-shop/api/gen/marketing/v1"
	"demo-shop/services/marketing/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoCouponTemplate 把 model 转成 proto。status 按当前时间推导。
func ToProtoCouponTemplate(t *model.CouponTemplate) *v1_marketingv1.CouponTemplate {
	if t == nil {
		return nil
	}
	return &v1_marketingv1.CouponTemplate{
		TemplateId:      t.TemplateId,
		CouponName:      t.CouponName,
		CouponType:      t.CouponType,
		ThresholdAmount: t.ThresholdAmount,
		DiscountAmount:  t.DiscountAmount,
		TotalCount:      t.TotalCount,
		ReceivedCount:   t.ReceivedCount,
		PerUserLimit:    t.PerUserLimit,
		UsableDays:      t.UsableDays,
		StartTime:       timestamppb.New(t.StartTime),
		EndTime:         timestamppb.New(t.EndTime),
		IsDeleted:       t.IsDeleted,
		CreatedAt:       timestamppb.New(t.CreatedAt),
		UpdatedAt:       timestamppb.New(t.UpdatedAt),
		Status:          model.DeriveCouponStatus(t, time.Now()),
	}
}

// ToProtoCouponTemplateList 批量转换
func ToProtoCouponTemplateList(list []*model.CouponTemplate) []*v1_marketingv1.CouponTemplate {
	out := make([]*v1_marketingv1.CouponTemplate, 0, len(list))
	for _, t := range list {
		out = append(out, ToProtoCouponTemplate(t))
	}
	return out
}

// FromProtoCouponTemplate 把 proto 入参转成 model。
// created_at / updated_at 由数据库生成,不从入参取。
func FromProtoCouponTemplate(p *v1_marketingv1.CouponTemplate) *model.CouponTemplate {
	if p == nil {
		return &model.CouponTemplate{}
	}
	tpl := &model.CouponTemplate{
		TemplateId:      p.GetTemplateId(),
		CouponName:      p.GetCouponName(),
		CouponType:      p.GetCouponType(),
		ThresholdAmount: p.GetThresholdAmount(),
		DiscountAmount:  p.GetDiscountAmount(),
		TotalCount:      p.GetTotalCount(),
		PerUserLimit:    p.GetPerUserLimit(),
		UsableDays:      p.GetUsableDays(),
	}
	// 固定有效期可缺省(相对有效期模式下不带),零值即"未设置"
	if p.StartTime != nil {
		tpl.StartTime = p.GetStartTime().AsTime()
	}
	if p.EndTime != nil {
		tpl.EndTime = p.GetEndTime().AsTime()
	}
	return tpl
}

// ToProtoUserCoupon 把 model 转成 proto。
// 四个模板字段来自 LEFT JOIN,由 UserCouponView 一起带出来 ——
// 故这里接收的是视图而非裸实体。
func ToProtoUserCoupon(v *model.UserCouponView) *v1_marketingv1.UserCoupon {
	if v == nil {
		return nil
	}
	out := &v1_marketingv1.UserCoupon{
		UserCouponId:    v.UserCouponId,
		TemplateId:      v.TemplateId,
		UserId:          v.UserId,
		Status:          v.Status,
		OrderNo:         v.OrderNo,
		ExpireAt:        timestamppb.New(v.ExpireAt),
		CreatedAt:       timestamppb.New(v.CreatedAt),
		CouponName:      v.CouponName,
		CouponType:      v.CouponType,
		ThresholdAmount: v.ThresholdAmount,
		DiscountAmount:  v.DiscountAmount,
	}
	// used_at 可空:未使用的券不返回该字段
	if v.UsedAt != nil {
		out.UsedAt = timestamppb.New(*v.UsedAt)
	}
	return out
}

// ToProtoUserCouponList 批量转换
func ToProtoUserCouponList(list []*model.UserCouponView) []*v1_marketingv1.UserCoupon {
	out := make([]*v1_marketingv1.UserCoupon, 0, len(list))
	for _, v := range list {
		out = append(out, ToProtoUserCoupon(v))
	}
	return out
}

// ToProtoCouponForUser 领券中心项转 proto。
// held_count / remaining_count 及模板 status 都在这层补齐,调用方不必自己算。
func ToProtoCouponForUser(c *model.CouponForUser) *v1_marketingv1.CouponTemplateForUser {
	if c == nil {
		return nil
	}
	return &v1_marketingv1.CouponTemplateForUser{
		Template:       ToProtoCouponTemplate(&c.CouponTemplate),
		HeldCount:      c.HeldCount,
		RemainingCount: c.RemainingCount,
	}
}

// ToProtoCouponForUserList 批量转换
func ToProtoCouponForUserList(list []*model.CouponForUser) []*v1_marketingv1.CouponTemplateForUser {
	out := make([]*v1_marketingv1.CouponTemplateForUser, 0, len(list))
	for _, c := range list {
		out = append(out, ToProtoCouponForUser(c))
	}
	return out
}
