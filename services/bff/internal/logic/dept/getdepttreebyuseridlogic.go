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

type GetDeptTreeByUserIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeptTreeByUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptTreeByUserIdLogic {
	return &GetDeptTreeByUserIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetDeptTreeByUserId 取某用户所属部门的树。
//
// ============================================================
// 路径参数是 **:userId**,不是 :id
// ============================================================
//
// 同组还有 GET /admin/dept/:id(取单个部门),而这条是
// GET /admin/dept/tree/:userId —— **参数名不同**。
//
// 这不是风格问题:判权将按 c.FullPath() 反查权限码,
// FullPath 里含参数名。把 :userId 改成 :id 会让
// "/api/v1/admin/dept/tree/:id" 与既有权限数据对不上,
// 那一条接口的判权直接失效。
//
// 故 .api 里 types.GetDeptTreeByUserIdReq 的 tag 是 path:"userId"。
//
// ============================================================
// 响应是**裸数组**
// ============================================================
//
// 单体: utils.Success(c, deptTree)。故 .api 里是
// `returns ([]DeptItem)`,生成 (resp []types.DeptItem, err error)。
//
// ============================================================
// 与菜单树的一处语义差别(proto 注释明确写了)
// ============================================================
//
//	菜单树:无权限时返回**空数组**,不返回错误
//	部门树:用户**无所属部门时返回业务错误** ← 本接口
//
// 故这条的 ResultBiz 分支是**真会走到**的,不要像菜单树那样
// 在注释里写"基本走不到"。回 400,文案由服务端给。
func (l *GetDeptTreeByUserIdLogic) GetDeptTreeByUserId(req *types.GetDeptTreeByUserIdReq) ([]types.DeptItem, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetDeptTreeByUserId(ctx, &v1_userv1.GetDeptTreeByUserIdReq{
		UserId: req.UserId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 用户无所属部门 → 400(与菜单树不同,见上方说明)
		return nil, err
	}

	return converter.DeptItems(resp.GetItems()), nil
}
