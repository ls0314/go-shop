// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/middleware"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetUserInfo 取指定用户的档案(**仅限本人**)。
//
// ============================================================
// 归属校验在 BFF 做,不是权限码
// ============================================================
//
// 单体 user_info_handler.go:64-73:
//
//	userId, _, err := GetUserInfoByContext(c)   // JWT 里的
//	if userId != id {                            // 路径里的
//		utils.Fail(c, 400, model.UserInfoError.Error())
//		c.Abort()
//		return
//	}
//
// 即:**用令牌里的身份与路径参数比对**,不一致就拒绝。
// 这不是权限码(4 条路由都不挂 permMW),而是归属校验 ——
// 所以 BFF 这边也要在 logic 里做同样的事,不能只靠"有令牌就放行"。
//
// **为什么必须在 BFF 做**:路径参数 :id 是 BFF 收到的,
// user-service 的 GetUserProfileReq 只有一个 user_id ——
// 服务端不知道该请求原本想查谁。归属语义属于"接口层",
// 只有 BFF 同时看得到令牌身份与路径参数。
//
// ============================================================
// c.Abort() 的对应物
// ============================================================
//
// 单体那句 c.Abort() 是 gin 的"终止后续 handler"。BFF 这边
// 直接在 logic 里 return error,handler 会走 response.Failure 分支
// 并 return —— 天然就终止了,不需要额外机制。
//
// 回 400(不是 403):文案与状态码都对齐单体
// (utils.Fail(c, 400, ...) → 400)。
func (l *GetUserInfoLogic) GetUserInfo(req *types.UserIdReq) (*types.GetUserInfoResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	// 归属校验:令牌身份必须等于路径参数
	if userId != req.Id {
		return nil, errUserInfoForbidden
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	// 注意传的是 req.Id(我们已经校验它等于 userId)。
	// proto 的入参名是 user_id —— GetUserProfile 按用户 ID 查,不是按档案 ID。
	resp, grpcErr := l.svcCtx.UserRPC.GetUserProfile(ctx, &v1_userv1.GetUserProfileReq{
		UserId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 档案不存在等 → 400
		return nil, err
	}

	// data 是 UserProfile 本体(单体: utils.Success(c, profile))
	return toGetUserInfoResp(resp.GetProfile()), nil
}
