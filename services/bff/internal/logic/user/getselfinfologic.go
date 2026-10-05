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

type GetSelfInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSelfInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSelfInfoLogic {
	return &GetSelfInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetSelfInfo 取当前登录用户的身份与档案。
//
// ============================================================
// user_id 从哪来:context,不是请求体
// ============================================================
//
// 这条路由挂在 Auth 组上,Auth 中间件验签后把 user_id / username
// 放进 context(见 middleware/authmiddleware.go 的 UserID/Username)。
// logic 从中读取。
//
// **接口有请求体但它不含 user_id** —— types.Empty 是空的。
// 这是刻意的:让客户端传 user_id 等于允许"查别人的信息",
// 而归属必须由服务端从令牌决定。
//
// 若取不到身份,说明路由没挂 Auth(配置错误,不是用户错误):
// 返回错误让它落到 500 并在日志里现形,而不是静默返回空数据。
//
// ============================================================
// 响应形状:6 个扁平字段 + 嵌套 profile
// ============================================================
//
// 单体 GetUserInfo 的响应是 handler 内联拼的 gin.H:
//
//	utils.Success(c, gin.H{
//		"user_id": ..., "username": ..., "email": ...,
//		"phone": ..., "status": ..., "profile": profile,
//	})
//
// **注意 profile 是嵌套对象**(UserProfile 本体),而非把档案字段
// 摊平到外层。故 types.GetSelfInfoResp 里 Profile 是一个子结构 ——
// 这不是可选的风格问题,前端解的是 res.data.profile.nickname。
//
// 另外注意:**这不是 ListUsers 那种 items/total 形状**,也不与
// GetUserProfile(返回 UserProfile 本体)相同。
func (l *GetSelfInfoLogic) GetSelfInfo(req *types.Empty) (*types.GetSelfInfoResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		// 走到这里说明该路由没挂 Auth 中间件 —— 配置错误。
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.GetSelfInfo(ctx, &v1_userv1.GetSelfInfoReq{UserId: userId})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 用户不存在等 → 400
		return nil, err
	}

	u := resp.GetUser()
	p := resp.GetProfile()

	// profile 可能为 nil(档案未创建)——
	// 那时输出 "profile": {零值},与单体传入 nil UserProfile 后
	// gin 序列化出的空对象一致。故这里不做 nil 判断,直接取值,
	// Getter 与 formatTimestamp 都是 nil 安全的。
	return &types.GetSelfInfoResp{
		UserId:   u.GetUserId(),
		Username: u.GetUsername(),
		Email:    u.GetEmail(),
		Phone:    u.GetPhone(),
		Status:   u.GetStatus(),
		Profile: types.SelfProfile{
			UserInfoId: p.GetUserInfoId(),
			UserId:     p.GetUserId(),
			Nickname:   p.GetNickname(),
			RealName:   p.GetRealName(),
			Gender:     p.GetGender(),
			AvatarUrl:  p.GetAvatarUrl(),
			Birthdate:  formatTimestamp(p.GetBirthdate()),
			CreatedAt:  formatTimestamp(p.GetCreatedAt()),
			UpdatedAt:  formatTimestamp(p.GetUpdatedAt()),
		},
	}, nil
}
