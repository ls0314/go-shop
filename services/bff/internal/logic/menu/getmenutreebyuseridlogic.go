// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeByUserIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuTreeByUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeByUserIdLogic {
	return &GetMenuTreeByUserIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetMenuTreeByUserId 取当前登录用户的菜单树(前端动态路由用)。
//
// ============================================================
// 这条路由在 /admin/menu 下但**刻意不判权**
// ============================================================
//
// 单体 menu_routes.go 的注释写明了这个例外:前端登录后要立刻拿
// 动态路由,此时用户的权限码可能还没加载完 —— 若这条也判权会死锁
// (要拿菜单得先有权限,要有权限得先拿菜单)。
//
// 服务端按 userId 过滤(只返回该用户有权访问的菜单),
// 故"免判权"不会泄露别人能看什么。
//
// ============================================================
// user_id 从 context 取,不是从请求体
// ============================================================
//
// .api 里 GetMenuTreeByUserIdReq 有个 user_id 字段(因为单体是
// POST + body),但**这里不信任它** —— 用 JWT 里的身份。
//
// 为什么:若用请求体的 user_id,任何登录用户都能查别人的菜单树
// (间接得知别人的权限范围)。单体那边用的是
// GetUserInfoByContext(从上下文取),故 BFF 也要一致。
//
// 请求体里那个字段保留是为了兼容既有前端(它会传),
// 但值被忽略。
//
// ============================================================
// 响应是**裸数组**,不是 {list: [...]}
// ============================================================
//
// 单体: utils.Success(c, menuTree),而 menuTree 是 []*model.SysMenu。
// 故 .api 里这条是 `returns ([]MenuItem)` —— goctl 会生成
// (resp []types.MenuItem, err error),切片直接成为 data。
//
// **不要包一层** —— 前端是 res.data.map(...) 还是 res.data.list.map(...)
// 取决于此,包错了整个侧边栏渲染不出来。
func (l *GetMenuTreeByUserIdLogic) GetMenuTreeByUserId(req *types.GetMenuTreeByUserIdReq) ([]types.MenuItem, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.GetMenuTreeByUserId(ctx, &v1_userv1.GetMenuTreeByUserIdReq{
		// 用令牌身份,不用 req.UserId(见上方说明)
		UserId: userId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 单体注释:菜单树无权限时返回**空数组**,不返回错误
		// (与部门树的语义不同 —— 那个在无部门时报错)。
		// 故这里的 ResultBiz 分支基本走不到;走到了就照常回 400。
		return nil, err
	}

	// converter.MenuItems 返回 make(..., 0, len) —— 空时是 [] 而不是 null
	return converter.MenuItems(resp.GetItems()), nil
}
