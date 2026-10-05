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

type ClearUserDeptsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearUserDeptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearUserDeptsLogic {
	return &ClearUserDeptsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ClearUserDepts 清空某用户的部门绑定。
//
// 与 ClearUserRoles 同一口径(见那边的说明):
// 与 AssignUserDepts 传空数组功能重叠但保留两者,幂等,不预检。
//
// ============================================================
// 一个服务端才关心的问题:主部门怎么处理
// ============================================================
//
// 清空部门绑定后,**主部门自然不存在了**(user_dept 里没有记录,
// 也就没有"哪个是主部门")。
//
// 服务端应当同时把主部门标记清掉。若它漏了,后续
// ListUserDepts 会返回 primary_dept_id 指向一个用户已不属于的部门 ——
// 前端会在列表里找不到那一项,于是"主部门"标记凭空消失,
// 而数据里还留着脏值。
//
// **BFF 无法从响应里验证这一点**(ClearUserDeptsResp 只有 error_msg),
// 故它属于"迁移后要专门验一次"的行为,不是靠读代码能保证的。
// 验证方法:清空后调 GET /:id/dept,断言 primary_dept_id == 0。
func (l *ClearUserDeptsLogic) ClearUserDepts(req *types.RelIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ClearUserDepts(ctx, &v1_userv1.ClearUserDeptsReq{
		UserId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.Empty{}, nil
}
