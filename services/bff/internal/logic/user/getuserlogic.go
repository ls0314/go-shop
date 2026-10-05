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

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetUser 根据 ID 查用户(管理端)。
//
// ============================================================
// 这条**没有归属校验**,与 /admin/user/info/:id 不同
// ============================================================
//
// 两条路由的路径很像,语义完全不同:
//
//	GET /api/v1/admin/user/:id        ← 本接口。管理端查任意用户
//	GET /api/v1/admin/user/info/:id   ← 只能查**本人**(归属校验)
//
// 判据在单体的路由挂载上:/admin/user 那组挂 permMW(需权限码),
// 而 /admin/user/info 那组不挂(靠归属校验)。
//
// 故 BFF 这边:
//
//	本接口  → Auth 中间件验签即可,logic 里不比对 userId
//	info 接口 → logic 里必须比对 userId == req.Id
//
// **不要因为"路径都在 admin 下"就给两条加同样的校验** ——
// 给本接口加归属校验会让管理员查不了别人;
// 给 info 接口去掉校验会让任何登录用户能看别人的档案。
//
// 判权(权限码)是 BFF 的下一步,当前只做"验签 + 转发"。
func (l *GetUserLogic) GetUser(req *types.UserIdReq) (*types.GetUserResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.GetUser(ctx, &v1_userv1.GetUserReq{
		UserId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 用户不存在 → 400
		return nil, err
	}

	// data 是 proto User 本体(单体: utils.Success(c, user))
	return toGetUserResp(resp.GetUser()), nil
}
