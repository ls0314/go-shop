// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dept

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDeptLogic {
	return &DeleteDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteDept 删除部门(管理端)。
//
// ============================================================
// 删部门的连带影响面比其它域都大
// ============================================================
//
// 一个部门可能被这些引用:
//
//	user_dept      用户-部门绑定(删了它,那些用户少一个部门)
//	                → 若它是某用户的**主部门**,primary_dept_id 会悬空
//	dept.parent_id 子部门(删了父部门,子部门变孤儿)
//	scope          数据权限规则可能引用 dept_id 作为过滤值
//	                (如 `dept_id IN (1,2,3)` —— 删了部门3,
//	                 那条规则里就多了一个不存在的 ID)
//
// **前两个由服务端决定策略**(拒绝删除 / 级联 / 把子部门上提),
// BFF 不做预检(需要跨表查,且是 TOCTOU)。
//
// **第三个 BFF 与服务端都看不全**:scope.condition_value 是自由文本,
// 服务端不会去解析并更新它。故删部门后可能留下悬空的权限规则值,
// 而那条规则仍然"生效"(只是那个 ID 永远匹配不到数据)。
//
// 这是**数据治理层面的遗留**,不是接口 bug。但值得在运维文档里写明:
// 删部门后要检查 scope 表里有没有引用它的条件值。
//
// ============================================================
// 与 GetDeptTreeByUserId 的关系
// ============================================================
//
// 删掉某用户的**唯一**部门后,GET /admin/dept/tree/:userId
// 会返回业务错误(proto 注释:"用户无所属部门时返回业务错误
// (与菜单树返回空数组的语义不同)")。
//
// 故前端在删部门后若立刻刷新那个用户的部门树,会从"有数据"
// 变成"报错" —— 那不是 bug。
func (l *DeleteDeptLogic) DeleteDept(req *types.DeptIdReq) (*types.Empty, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.DeleteDept(ctx, &v1_userv1.DeleteDeptReq{
		DeptId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 部门不存在 / 有子部门不可删 / 有用户归属 → 400
		return nil, err
	}

	return &types.Empty{}, nil
}
