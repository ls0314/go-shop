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

type DeleteAddressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddressLogic {
	return &DeleteAddressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteAddress 删除收货地址。
//
// ============================================================
// 删掉的若是默认地址,服务端要补一个新的默认
// ============================================================
//
// proto 的 CreateAddress 注释说明了为什么必须有默认地址:
//
//	"否则会得到一个'有地址但没有默认地址'的状态,
//	 而结算页要预选默认地址"
//
// 故删掉默认地址后,服务端**必须**把剩下的某一条(按它的策略,
// 通常是 updated_at 最新的那条)提升为默认。
//
// 这正是地址域那条"删默认地址会提升 updated_at 最大的为默认"
// 的既有行为(见 CLAUDE 记录里的迁移说明)。
//
// BFF 不参与这个决策 —— 它是服务端事务的一部分。
//
// ============================================================
// 删最后一条地址是允许的
// ============================================================
//
// 与"必须有默认地址"不矛盾:一条都没有时就不存在"有没有默认"
// 的问题。故 DeleteAddress 不做"至少留一条"的校验。
//
// (对比:CreateAddress 有"最多 20 条"的上限,见 AddressMaxCount。)
//
// 响应:单体是 utils.Success(c, nil) → "data": null,
// 而这里 returns (Empty) → "data": {}。已知差异。
func (l *DeleteAddressLogic) DeleteAddress(req *types.AddrIdReq) (*types.Empty, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.AddressRPC.DeleteAddress(ctx, &v1_userv1.DeleteAddressReq{
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
