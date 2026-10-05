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

type DeleteUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserInfoLogic {
	return &DeleteUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteUserInfo 删除指定用户的档案(**仅限本人**)。
//
// 归属校验与同组另两条一致:令牌身份必须等于路径 :id。
//
// ============================================================
// 响应是 nil,不是空对象
// ============================================================
//
// 单体: utils.Success(c, nil) → "data": null
//
// 而 .api 里这条声明的是 returns (Empty),handler 传的是
// 生成物给出的 **&types.Empty{}** —— 序列化成 "data": {}
// 而不是 null。
//
// **这是 BFF 与单体的一个可见差异。** 是否要紧取决于前端:
//
//	若前端判 res.data === null → 会走 else 分支,可能误判失败
//	若前端只判 res.code === 200 → 无影响
//
// 受影响的一共是 14 条 returns (Empty) 的路由。
//
// 彻底对齐的做法是让 handler 在无返回值时传 nil ——
// 那是 goctl handler.tpl 里 {{if .HasResp}}...{{else}}response.OK(w, nil){{end}}
// 那个分支,但它只对"声明为无返回值"的路由生效,而
// returns (Empty) 会被 goctl 当作**有**返回值(HasResp=true,
// 类型是 *types.Empty)。
//
// 故要在模板层做成 nil 需要额外判断"resp 是不是 *types.Empty" ——
// 那会让模板复杂且脆弱。**当前先保持 {}**,并在前端对接时确认一次。
// 若前端确实判 null,再单独处理这条(见 TODO)。
func (l *DeleteUserInfoLogic) DeleteUserInfo(req *types.UserIdReq) (*types.Empty, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	if userId != req.Id {
		return nil, errUserInfoForbidden
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.DeleteUserProfile(ctx, &v1_userv1.DeleteUserProfileReq{
		// 与 UpdateUserProfile 同一情况:proto 字段叫 user_profile_id,
		// 而路径 :id 在既有语义里是 user_id。单体也是这么传的
		// (见 userclient.DeleteUserProfile),这里保持一致。
		UserProfileId: req.Id,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.Empty{}, nil
}
