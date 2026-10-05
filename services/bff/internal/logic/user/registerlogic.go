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

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Register 自助注册。
//
// ============================================================
// 与 CreateUser(管理端建号)的区别
// ============================================================
//
// 单体 RegisterHandler 的注释写明了:
//
//	"注册会校验手机号格式、查重用户名/手机号/邮箱,并在同一事务内
//	 创建 user_profile(nickname 为空时回落为用户名)"
//
// 而 CreateUser 是管理员建号,**不走这些校验**。两条路由的服务端
// 逻辑不同,故 BFF 也不能把它们合并 —— 这里只是原样转发,
// 校验在 user-service 里(它是唯一知道"用户名是否已存在"的一方)。
//
// 所以 BFF **不做**任何字段级校验(如手机号正则):那会让同一套
// 规则存在两处,而两处漂移的表现是"BFF 放过了、服务端拒绝"或反之。
// 参数格式校验归 user-service,它是权威。
//
// ============================================================
// 为什么错误一律 400
// ============================================================
//
// 单体的 RegisterHandler:
//
//	utils.Fail(c, 400, errMsg)   // "口令/手机号格式不合法、用户名或
//	                             //  手机号已存在等均属业务失败,用 400"
//
// 与登录不同 —— 注册没有"凭据错误"这一说(还没有凭据),
// 故全部业务失败都是 400,不需要 WrapUnauthorized。
func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.RegisterResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.Register(ctx, &v1_userv1.RegisterReq{
		Username: req.Username,
		Password: req.Password,
		Phone:    req.Phone,
		Email:    req.Email,
		Nickname: req.Nickname,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 用户名已存在 / 手机号格式不合法 / 口令强度不足 → 400
		return nil, err
	}

	// resp.User 是 proto 的 User(RegisterResp.user)。
	//
	// 单体在这一层是 utils.Success(c, toModelUser(resp.User)) ——
	// 即 data 是**一个 User 对象**(不是 {user: {...}} 包装)。
	// types.RegisterResp 与 proto User 的字段集一致,故逐个映射。
	u := resp.GetUser()
	return &types.RegisterResp{
		UserId:    u.GetUserId(),
		Username:  u.GetUsername(),
		Email:     u.GetEmail(),
		Phone:     u.GetPhone(),
		Status:    u.GetStatus(),
		CreatedAt: formatTimestamp(u.GetCreatedAt()),
		UpdatedAt: formatTimestamp(u.GetUpdatedAt()),
	}, nil
}
