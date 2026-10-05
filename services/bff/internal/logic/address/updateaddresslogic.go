// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package address

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAddressLogic {
	return &UpdateAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateAddress 局部更新地址。
//
// 响应是**裸 Address 对象**(单体: utils.Success(c, address)),
// 故 .api 里 UpdateAddressResp 是匿名内嵌 AddressItem。
//
// 服务端返回的是**合并后的完整对象**(与 user-service 的
// UpdateUserProfile 同款),故前端可以直接用它刷新本地状态,
// 不需要再查一次。
//
// ============================================================
// is_default 在这里传 true 的含义
// ============================================================
//
// 把它更新为 true = "设为默认地址"。此时服务端**必须同时**
// 把同用户的其它地址取消默认(只能有一个默认)——
// 那是一个事务里的两步,由服务端负责。
//
// 而把它更新为 false 呢?语义不清晰:是"取消这条的默认"
// 还是"把默认让给别人"?前者会导致"没有任何默认地址"的状态,
// 而 proto 的 CreateAddress 注释明确说那种状态是要避免的
// (结算页要预选默认地址)。
//
// **BFF 不拦 is_default=false** —— 服务端若不允许会返回错误。
// 但前端的交互通常应当用 PUT /:id/default 那条接口切换默认,
// 而不是在这里传 is_default。
//
// ============================================================
// nil 指针 = 不更新
// ============================================================
//
// 全部为 nil 时回 400(errNothingToUpdate)。
//
// 注意 .api 里 UpdateAddressReq 的字段名与 CreateAddressReq 一致
// (receiver_name / receiver_phone / ...),而 DB 列名也是同一套
// —— 故 FieldUpdate 的 field 名直接用这些 JSON 名。
func (l *UpdateAddressLogic) UpdateAddress(req *types.UpdateAddressReq) (*types.UpdateAddressResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	updates := compactFields(
		strField("receiver_name", req.ReceiverName),
		strField("receiver_phone", req.ReceiverPhone),
		strField("province", req.Province),
		strField("city", req.City),
		strField("district", req.District),
		strField("detail_address", req.DetailAddress),
		strField("postal_code", req.PostalCode),
		boolField("is_default", req.IsDefault),
		strField("address_tag", req.AddressTag),
	)

	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.AddressRPC.UpdateAddress(ctx, &v1_userv1.UpdateAddressReq{
		UserId:    userId,
		AddressId: req.Id,
		Updates:   updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 地址不存在 / 手机号格式错 → 400
		return nil, err
	}

	return &types.UpdateAddressResp{
		AddressItem: converter.AddressItem(resp.GetAddress()),
	}, nil
}
