// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRolePermsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRolePermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolePermsLogic {
	return &AssignRolePermsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AssignRolePerms 全量替换角色的权限绑定。
//
// ============================================================
// **全量替换,不是追加**
// ============================================================
//
// proto 注释:"AssignRolePerms 全量替换角色的权限绑定(先清空再写入)"。
//
// 故前端传的必须是**完整的目标集合**:
//
//	角色原本有 [1,2],想加 3
//	  → 传 [1,2,3]   ✅
//	  → 传 [3]       ❌ [1,2] 被清掉了
//
// BFF 不做"聪明合并" —— 合并要先查当前绑定,一次写变成"读+写",
// 两个并发请求会互相覆盖且结果取决于时序。全量替换能在服务端的
// 一条事务里完成。
//
// ============================================================
// 空数组是合法输入:表示清空
// ============================================================
//
// 不要拦 perm_ids: [] —— 它与 ClearRolePerms 功能重叠但语义更自然
// (一次调用同时表达"设成哪些"),前端"全部取消勾选后保存"会用到。
//
// ============================================================
// 响应回显请求体(与单体的已知差异)
// ============================================================
//
// 单体是 utils.Success(c, req) —— 把绑定的 {role_id, perm_ids}
// 回显给前端。而 .api 里这条写的是 returns (Empty),故这里回 {}。
//
// **这是一处与单体的可见差异**。要不要对齐取决于前端是否用这个回显
// (见 bff.api 里关于 assign 响应的说明,那里记了完整理由:
// 写 returns (AssignRolePermsReq) 会让 goctl 生成"入参与出参同类型"
// 的签名,读代码时看不出是回显)。
//
// 待端到端验证时确认。若前端确实依赖,再给这 5 条各造一个响应类型。
func (l *AssignRolePermsLogic) AssignRolePerms(req *types.AssignRolePermsReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.AssignRolePerms(ctx, &v1_userv1.AssignRolePermsReq{
		RoleId:  req.RoleId,
		PermIds: req.PermIds,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 角色不存在 / 权限 ID 不存在 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
