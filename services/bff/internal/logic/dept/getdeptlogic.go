// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dept

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptLogic {
	return &GetDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetDept 根据 ID 查部门(管理端)。
//
// 响应是**裸对象**(单体: utils.Success(c, dept)),
// 故 .api 里 GetDeptResp 是匿名内嵌 DeptItem。
//
// 路径参数是普通的 :id(与同组的 :userId 那条树查询不同)。
func (l *GetDeptLogic) GetDept(req *types.DeptIdReq) (*types.GetDeptResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetDept(ctx, &v1_userv1.GetDeptReq{
		DeptId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 部门不存在 → 400
		return nil, err
	}

	return &types.GetDeptResp{
		DeptItem: converter.DeptItem(resp.GetDept()),
	}, nil
}
