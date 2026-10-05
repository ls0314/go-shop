// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserDeptsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserDeptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserDeptsLogic {
	return &ListUserDeptsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListUserDepts 取某用户所属的部门。
//
// 路径参数 :id 的语义是 user_id(与 ListUserRoles 同理)。
//
// ============================================================
// 比 ListUserRoles 多一个字段:primary_dept_id
// ============================================================
//
// proto 的 ListUserDeptsResp 有:
//
//	repeated Dept items = 1;
//	int64        total  = 2;
//	int64        primary_dept_id = 3;   // 主部门的**部门ID**,无主部门时为 0
//
// **注意这个字段与 AssignUserDepts 的 primary_index 不是一个东西**:
//
//	AssignUserDepts.primary_index   ← 入参,是 dept_ids 里的**下标**
//	ListUserDeptsResp.primary_dept_id ← 出参,是**部门 ID**
//
// 一个下标一个 ID,读代码时极易混。前端典型用法:
//
//	列表渲染时,把 dept_id == primary_dept_id 的那一项标为"主部门"
//
// 故这个字段必须透传,不能省 —— 省了前端无法标出哪个是主部门。
//
// 无主部门时为 0,直接透传 0 即可(不要"兜"成 null 或省略字段,
// 前端按 0 判断)。
func (l *ListUserDeptsLogic) ListUserDepts(req *types.RelIdReq) (*types.ListUserDeptsResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListUserDepts(ctx, &v1_userv1.ListUserDeptsReq{
		UserId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListUserDeptsResp{
		List:  converter.DeptItems(resp.GetItems()),
		Total: resp.GetTotal(),
		// 部门 ID,不是下标(见上方说明)
		PrimaryDeptId: resp.GetPrimaryDeptId(),
	}, nil
}
