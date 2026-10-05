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

type AssignUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserRoleLogic {
	return &AssignUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AssignUserRole 全量替换用户的角色绑定。
//
// ============================================================
// **全量替换,不是追加** —— 这是最容易误用的一个语义
// ============================================================
//
// proto 注释写得很明确:
//
//	"AssignUserRoles 全量替换用户的角色绑定(先清空再写入)"
//
// 所以前端传的必须是**完整的目标集合**,而不是"要新增的那几个":
//
//	用户原本有 [1,2],想加 3
//	  → 传 [1,2,3]   ✅ 结果 [1,2,3]
//	  → 传 [3]       ❌ 结果 [3](1 和 2 被清掉了)
//
// BFF 不改这个语义(不"聪明地"去合并)。合并需要先查当前绑定,
// 会把一次写变成"读+写",引入竞态:两个并发的分配请求会互相覆盖,
// 且最终结果取决于时序。全量替换在服务端能用一条事务完成。
//
// ============================================================
// 空数组是合法输入
// ============================================================
//
// role_ids: [] 表示"清空该用户的所有角色"。
//
// **不要在这里拦空数组** —— 那与 ClearUserRoles 功能重叠,但语义
// 更自然(一次调用同时表达"设成哪些"),前端可能就用它。
// 拦截会让前端的"取消全部勾选后保存"失败。
//
// 与 errNothingToUpdate 的区别:那条针对**局部更新字段全空**
// (没有任何信息量);而这里是全量替换,空集本身就是一个有意义的值。
func (l *AssignUserRoleLogic) AssignUserRole(req *types.AssignUserRolesReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.AssignUserRoles(ctx, &v1_userv1.AssignUserRolesReq{
		UserId: req.UserId,
		// nil 与空切片在 proto 的 repeated 字段里编码相同(都是空),
		// 故不需要把 nil 兜成 []int64{}
		RoleIds: req.RoleIds,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 用户不存在 / 角色不存在 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
