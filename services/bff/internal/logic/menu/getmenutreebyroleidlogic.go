// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeByRoleIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuTreeByRoleIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeByRoleIdLogic {
	return &GetMenuTreeByRoleIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetMenuTreeByRoleId 取某角色已绑定的菜单树(角色分配菜单页回显)。
//
// ============================================================
// 与 GetMenuTreeByUserId 的两处关键差别
// ============================================================
//
//	① **是 GET 而不是 POST** —— 路径 GET /admin/menu/:id/tree,
//	   路径参数 :id 的语义是 **role_id**。
//	   而 GetMenuTreeByUserId 是 POST /admin/menu/tree(body 带 user_id)。
//	   两条路由长得像但方法、入参来源、语义都不同。
//
//	② **用路径参数,不用令牌身份** —— 这条是"看某个角色绑了哪些菜单",
//	   是管理操作,角色 ID 来自路径。而上面那条是"看我能访问哪些菜单",
//	   身份必须来自令牌。
//
// 把这两条的取值方式搞混的后果很具体:
//
//	用令牌身份查角色树 → 管理页永远显示"当前登录用户的菜单"
//	用路径参数查用户树 → 任何登录用户能查别人的权限范围
//
// ============================================================
// 响应是**裸数组**(同 GetMenuTreeByUserId)
// ============================================================
//
// 单体: utils.Success(c, menuTree)。前端 el-tree 的 :data 直接绑
// res.data。
func (l *GetMenuTreeByRoleIdLogic) GetMenuTreeByRoleId(req *types.GetMenuTreeByRoleIdReq) ([]types.MenuItem, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetMenuTreeByRoleId(ctx, &v1_userv1.GetMenuTreeByRoleIdReq{
		// 路径参数 :id,语义是 role_id(见上方说明)
		RoleId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return converter.MenuItems(resp.GetItems()), nil
}
