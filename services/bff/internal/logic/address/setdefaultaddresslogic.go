// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package address

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetDefaultAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultAddressLogic {
	return &SetDefaultAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SetDefaultAddress 把某条地址设为默认。
//
// ============================================================
// 为什么单独一条接口,而不是用 UpdateAddress 传 is_default
// ============================================================
//
// 两者都改 is_default,但语义不同:
//
//	PUT /:id/default          "设为默认"(幂等,一定成功)
//	PUT /:id  {is_default:true} "更新这条地址,顺便把默认也改成 true"
//
// 前者是**动作**,后者是**字段更新**。分开的理由:
//
//	① 前端"设为默认"按钮的意图明确对应一个动作,
//	   不需要构造一个只含 is_default 的更新体;
//	② 幂等语义清晰:对已经是默认的地址再调一次应当成功
//	   (而不是报"没有变化"或"参数错误");
//	③ 审计日志能记下"设为默认"这个动作,而不是一条含糊的 update。
//
// ============================================================
// 服务端必须做两步(事务内)
// ============================================================
//
// 设为默认要同时:
//
//	① 把目标地址 is_default = true
//	② 把同用户的**其它**地址 is_default = false
//
// 只做①会得到多条默认地址 —— 那会让结算页预选到哪一条变得不确定,
// 而且不会报错。
//
// BFF 不参与(它是服务端事务的职责)。但**验证时要专门查这一点**:
// 连续把两条地址设为默认,然后查列表,确认只有一条 is_default=true。
//
// 响应:单体是 utils.Success(c, nil) → "data": null,
// 这里 returns (Empty) → "data": {}。已知差异。
func (l *SetDefaultAddressLogic) SetDefaultAddress(req *types.AddrIdReq) (*types.Empty, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.AddressRPC.SetDefaultAddress(ctx, &v1_userv1.SetDefaultAddressReq{
		UserId:    userId,
		AddressId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 地址不存在(或不属于当前用户)→ 400
		return nil, err
	}

	return &types.Empty{}, nil
}
