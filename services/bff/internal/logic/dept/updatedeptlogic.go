// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dept

import (
	"context"

	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	v1_userv1 "demo-shop/api/gen/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDeptLogic {
	return &UpdateDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateDept 局部更新部门(管理端)。
//
// ============================================================
// 可改字段里没有 parent_id —— 刻意如此
// ============================================================
//
// .api 里 UpdateDeptReq 是 dept_name / dept_type / leader_id /
// sort_order / status 五个指针,**没有 parent_id**。
//
// 理由:改 parent_id 等于"把整棵子树挪到另一个父节点下" ——
// 服务的实现里有额外逻辑(见 user-service 的 UpdateDept:
// 换父节点时会 `GetMaxSortId(新父)` 并把 sort_order 设成末尾)。
// 那是一个**移动**语义,不是逐字段更新。
//
// 更重要的是它需要校验:不能把节点挪到自己的后代下面(会形成环)。
// 这类校验属于服务端,但它必须**先知道这是一次移动**才做得对 ——
// 混在普通的字段更新里容易漏。
//
// 故不暴露 parent_id。前端要做移动应当走一个显式的接口
// (本项目目前没有,若需要应新增而不是在这里加字段)。
//
// ============================================================
// leader_id 传 0 表示"取消负责人"
// ============================================================
//
// 它是 *int64,故:
//
//	不传 leader_id          → nil,不更新
//	传 "leader_id": 0       → 更新为 0(即无负责人)
//	传 "leader_id": 123     → 更新为 123
//
// **注意 0 的语义要服务端认可**。若它把 0 当成"用户ID=0 的用户"
// 去查,会返回"用户不存在"而不是"清空负责人"。
// 端到端验证时试这一条。
//
// ============================================================
// nil 指针 = 不更新
// ============================================================
//
// 全部为 nil 时回 400(errNothingToUpdate)。
func (l *UpdateDeptLogic) UpdateDept(req *types.UpdateDeptReq) (*types.UpdateDeptResp, error) {
	updates := compactFields(
		strField("dept_name", req.DeptName),
		strField("dept_type", req.DeptType),
		int64Field("leader_id", req.LeaderId),
		int64Field("sort_order", req.SortOrder),
		strField("status", req.Status),
		// parent_id 不在列 —— 见上方说明(它是"移动",不是"更新")
	)

	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.UpdateDept(ctx, &v1_userv1.UpdateDeptReq{
		DeptId:  req.Id,
		Updates: updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 部门不存在 / 同父下重名 → 400
		return nil, err
	}

	// proto 注释:UpdateDeptResp.dept 返回**合并后的完整对象**
	return &types.UpdateDeptResp{
		DeptItem: converter.DeptItem(resp.GetDept()),
	}, nil
}
