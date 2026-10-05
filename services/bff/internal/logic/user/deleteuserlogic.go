// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteUser 删除用户(管理端)。
//
// 与同组的 GetUser/UpdateUser 同一口径:靠权限码授权(下一步做),
// BFF 这边不做归属校验。
//
// ============================================================
// 删用户的边界由服务端决定,不在 BFF 判
// ============================================================
//
// 这里有几类**该不该拒绝**的判断,全部属于 user-service 的领域知识:
//
//	删自己?          单体没拦,BFF 也不拦
//	删最后一个管理员? 需要知道"谁是管理员"——BFF 看不到角色绑定
//	有订单能删吗?     需要知道 trade 的数据 —— 跨服务,更不该在 BFF 判
//	软删还是硬删?     服务端决定(proto 的 DeleteUserResp 只有 error_msg,
//	                  没有"是否软删"的信息)
//
// 故 BFF 只做一件事:转发。任何在这里"顺手加"的规则都会变成
// 两处规则,而 BFF 那份必然与真正的权威漂移。
func (l *DeleteUserLogic) DeleteUser(req *types.UserIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.DeleteUser(ctx, &v1_userv1.DeleteUserReq{
		UserId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// 单体: utils.Success(c, nil) → "data": null
	// 这里回 &types.Empty{} → "data": {}。差异见 deleteuserinfologic.go
	// 里关于 returns (Empty) 的说明。
	return &types.Empty{}, nil
}
