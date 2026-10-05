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

type GetAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetAddress 取单条地址。
//
// ============================================================
// 响应是**裸 Address 对象**
// ============================================================
//
// 单体: utils.Success(c, address)
//
// data 就是那个地址对象,没有 address 那一层包装:
//
//	{ "code":200, "message":"Success", "data": {"address_id":7, ...} }
//
// 故 .api 里 GetAddressResp 是匿名内嵌 AddressItem
// —— 序列化后字段提升,与裸对象一致。
//
// ============================================================
// 归属校验由服务端做,不需要 BFF 预检
// ============================================================
//
// proto 的 GetAddressReq 是 { user_id, address_id } ——
// 服务端的查询是 `WHERE user_id = ? AND address_id = ?`。
//
// 故传别人的 address_id 时,服务端**查不到**并返回业务错误
// (如"地址不存在"),BFF 回 400。
//
// 这与档案域(/admin/user/info/:id)不同 —— 那个的 proto 只收一个
// user_id,服务端无法判断"请求者是不是本人",故归属校验必须在 BFF 做。
//
// **不要因为档案域要预检就在这里也加一个** —— 那会多一次 RPC,
// 且引入 TOCTOU(预检通过后地址被别人删了)。
func (l *GetAddressLogic) GetAddress(req *types.AddrIdReq) (*types.GetAddressResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.AddressRPC.GetAddress(ctx, &v1_userv1.GetAddressReq{
		UserId:    userId,
		AddressId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 地址不存在(或不属于当前用户,服务端不区分这两种)→ 400
		return nil, err
	}

	return &types.GetAddressResp{
		AddressItem: converter.AddressItem(resp.GetAddress()),
	}, nil
}
