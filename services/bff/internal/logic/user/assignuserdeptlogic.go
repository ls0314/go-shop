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

type AssignUserDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignUserDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserDeptLogic {
	return &AssignUserDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AssignUserDept 全量替换用户的部门绑定。
//
// 全量替换语义与 AssignUserRole 相同(见那边的说明),
// 但**多一个容易踩的参数**:primary_index。
//
// ============================================================
// primary_index 是**下标**,不是部门 ID
// ============================================================
//
// proto 注释:
//
//	"主部门在 dept_ids 中的下标,不是部门ID。越界或为负表示无主部门。"
//
// 这个设计很反直觉,但**不要在这里"顺手修正"成 ID**:
//
//	① 服务端按下标解析,传 ID 会被当成下标 ——
//	   若 ID 恰好落在 [0, len(dept_ids)) 区间,会静默选错主部门;
//	   否则被当成"无主部门",同样是错的。
//	② 前端也是按下标传的(选中列表里的第几个)。
//	   改语义要 BFF + 前端 + 服务端同时改。
//
// 故 BFF 原样透传,只在注释里写清。
//
// ============================================================
// 空数组时的 primary_index 必须为 0
// ============================================================
//
// dept_ids 为空(清空所有部门)时,primary_index 传任何值都是无意义的:
// 没有部门,自然没有主部门。proto 说"越界表示无主部门",
// 故 0 在空数组里也算越界(len=0),会被正确解释为无主部门。
//
// 不做特判 —— 服务端已经能正确处理。
func (l *AssignUserDeptLogic) AssignUserDept(req *types.AssignUserDeptsReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.AssignUserDepts(ctx, &v1_userv1.AssignUserDeptsReq{
		UserId:  req.UserId,
		DeptIds: req.DeptIds,
		// 下标,不是 ID(见上方说明)
		PrimaryIndex: req.PrimaryIndex,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 用户不存在 / 部门不存在 / primary_index 越界但服务端选择报错 → 400
		return nil, err
	}

	// AssignUserDeptsResp 只有 error_msg,故回空
	return &types.Empty{}, nil
}
