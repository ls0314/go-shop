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

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// RefreshToken 用 refresh 令牌换新的 access 令牌。
//
// ============================================================
// 为什么这条路由既不验签也不限流(纯转发)
// ============================================================
//
// refresh 令牌只有 user-service 能验 —— 它持私钥,且能区分
// access / refresh(令牌里带 tokenType)。BFF 的 Auth 中间件
// 会拒绝 refresh 令牌(ErrWrongTokenType),那是**正确行为**:
// 用 refresh 令牌访问业务接口本该被拒。
//
// 所以本接口挂在 Public 组上,**不做本地验签**,把令牌原样转发,
// 由 user-service 自行判定。DS-A-26 §六 明确写了这条 ——
// 提前在任一侧删掉验签,会让未迁移模块整片 401。
//
// ============================================================
// 错误语义与登录**相反**,这不是笔误
// ============================================================
//
// 登录的凭据错误回 **401**(前端跳登录页)。
// 而刷新失败回 **400**,因为单体的 RefreshHandler 就是这么做的:
//
//	utils.Fail(c, 400, resp.ErrorMsg)
//
// 理由:刷新失败意味着 refresh 令牌过期或无效,前端应当直接
// **清掉凭据并回登录页**,不需要"因为 401 所以跳登录"这层间接;
// 而 400 表达"这个请求本身不成立"。
//
// **这个差异是既有的,不要为了"看起来一致"而统一** ——
// 前端对这两个接口的处理路径不同。
func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenReq) (*types.RefreshTokenResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.RefreshToken(ctx, &v1_userv1.RefreshTokenReq{
		RefreshToken: req.RefreshToken,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 业务失败 → 400(见上方注释,与登录的 401 不同)
		return nil, err
	}

	// **只回 access_token** —— refresh 令牌不轮换,
	// 与既有前端行为一致(刷新后前端仍持有旧 refresh_token)。
	// 故 types.RefreshTokenResp 只有这一个字段,这里也只映射一个。
	return &types.RefreshTokenResp{
		AccessToken: resp.GetAccessToken(),
	}, nil
}
