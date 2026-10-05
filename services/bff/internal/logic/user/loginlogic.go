// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"
	"demo-shop/services/bff/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Login 用户登录。
//
// ============================================================
// client_ip / device 从 context 取,而不是从 *http.Request
// ============================================================
//
// user-service 的 LoginReq 要这两个字段用于写登录日志,而它们只在
// HTTP 层有。但 goctl 生成的 logic 构造函数只收 (ctx, svcCtx),
// 拿不到 *http.Request。
//
// 故由 RequestMeta 中间件把它们放进 context
// (见 middleware/requestmetamiddleware.go),本 logic 从中读取。
//
// **为什么不让 logic 持有 r**:那样 handler 要用一个非生成的
// 构造函数,而 handler 是生成物 —— 重跑 goctl 会退回两参版本,
// 于是"登录日志里没有 IP"这个缺陷会静默回来。走 context 则
// 生成物一行都不用动。
//
// 本 logic 依赖 RequestMeta 中间件已生效;它在 bff.api 的全部
// 11 个 @server 块里都配了,故正常路径一定满足。
//
// ============================================================
// 凭据错误为什么回 401 而不是 400
// ============================================================
//
// 单体的 LoginHandler 对凭据错误**特判回 401**:
//
//	c.JSON(401, gin.H{"code": 401, "message": resp.ErrorMsg, "data": nil})
//
// 注释写明理由:"前端 axios 拦截器据此判断是否需要重新登录"。
//
// 所以这一条不能走 response.Failure 的 400 默认分支 ——
// 用 rpc.WrapUnauthorized 挂上哨兵,让 Failure 判到 401。
//
// **怎么知道是"凭据错误"而不是别的业务失败?** 登录这条路径上
// user-service 只会因凭据返回 error_msg,故"ResultBiz 一律当凭据错误"。
//
// 若将来它在这个接口上加了别的业务失败(如账号被封禁),需要在这里
// 按文案细分 —— 那是一次明确的契约变更,不能靠"顺手放宽"解决。
func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.Login(ctx, &v1_userv1.LoginReq{
		Username: req.Username,
		Phone:    req.Phone,
		Password: req.Password,

		ClientIp: utils.ClientIPFrom(l.ctx),
		Device:   utils.DeviceFrom(l.ctx),
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		// 未建连/超时 → response.Failure 判成 503 或 500。
		//
		// **不在这里重试** —— 重试是上游(网关)的事,
		// BFF 同步重试只会把自己的超时叠加到上游的等待上。
		return nil, err
	case rpc.ResultBiz:
		// 凭据错误 → 401
		return nil, rpc.WrapUnauthorized(err)
	}

	// 显式映射,不直接返回 resp:
	// *v1_userv1.LoginResp 与 *types.LoginResp 是两个不同的具名 struct,
	// Go 不允许互相赋值(哪怕字段完全相同)。且 types.* 是对前端的契约,
	// 让 proto 直接流出去意味着"改了 proto 字段名"直接变成"前端挂了"。
	return &types.LoginResp{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}
