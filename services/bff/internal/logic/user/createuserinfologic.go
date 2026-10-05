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

type CreateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserInfoLogic {
	return &CreateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateUserInfo 创建当前登录用户的档案。
//
// ============================================================
// 路由在 /admin/user/info 下,但**不判权** —— 是"本人可写"
// ============================================================
//
// 单体的这 4 条路由挂在 AuthMiddleware 上,**没有权限码**
// (user_info_routes.go 里 4 条都不挂 permMW)。
// 它的授权模型是**归属校验**,不是权限码:
//
//	CreateUserInfo: 归属强制为 JWT 里的 userId,不接受请求体指定
//	其余 3 条:      JWT 里的 userId 必须等于路径 :id
//
// 故前缀里的 "admin" 是历史遗留的路径命名,**不代表需要管理员权限**。
// 不要因为看到 /admin 就给它加权限码 —— 那会让普通用户改不了
// 自己的资料。
//
// ============================================================
// 为什么不接受请求体传 user_id
// ============================================================
//
// 单体的做法:
//
//	profile.UserId = userId   // 归属强制为当前登录用户
//
// 请求体里即使带了 user_id 也被覆盖。BFF 这边更进一步:
// types.CreateUserInfoReq **根本没有 user_id 字段**,
// 客户端想传也传不了(传了会被 httpx.Parse 丢弃)。
//
// 这比"接收后覆盖"更好:契约层面就不给越权的可能。
func (l *CreateUserInfoLogic) CreateUserInfo(req *types.CreateUserInfoReq) (*types.CreateUserInfoResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	birthdate, err := parseBirthdate(req.Birthdate)
	if err != nil {
		return nil, err
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.CreateUserProfile(ctx, &v1_userv1.CreateUserProfileReq{
		Profile: &v1_userv1.UserProfile{
			// 归属由服务端决定,不信客户端
			UserId: userId,

			Nickname:  req.Nickname,
			RealName:  req.RealName,
			Gender:    req.Gender,
			AvatarUrl: req.AvatarUrl,
			Birthdate: birthdate,
		},
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return toCreateUserInfoResp(resp.GetProfile()), nil
}
