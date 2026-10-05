// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionLogic {
	return &GetPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetPermission 根据 ID 查权限点(管理端)。
//
// 响应是**裸权限对象**(单体: utils.Success(c, perm)),
// 没有 permission 那一层包装。故 .api 里 GetPermissionResp 是
// 匿名内嵌 PermissionItem —— 序列化后字段提升,与裸对象一致。
//
// 内嵌的代价:*PermissionItem 与 *GetPermissionResp 是两个具名类型,
// 要先取出来再拷进内嵌字段。
func (l *GetPermissionLogic) GetPermission(req *types.PermissionIdReq) (*types.GetPermissionResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetPermission(ctx, &v1_userv1.GetPermissionReq{
		PermissionId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 权限不存在 → 400
		return nil, err
	}

	return &types.GetPermissionResp{
		PermissionItem: converter.PermissionItem(resp.GetPermission()),
	}, nil
}
