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

type UpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateUser 局部更新用户(管理端)。
//
// ============================================================
// 与 UpdateUserInfo 的区别
// ============================================================
//
//	PUT /api/v1/admin/user/:id        ← 本接口。改 sys_user 表
//	                                     (username/phone/email/status/password)
//	PUT /api/v1/admin/user/info/:id   ← 改 user_profile 表
//	                                     (nickname/real_name/gender/avatar_url/birthdate)
//
// 两张表、两组字段、两个 RPC。**不要合并** ——
// 单体的 handler 也是分开的,前端调的是两个接口。
//
// ============================================================
// 本次**不含归属校验**
// ============================================================
//
// 与 GetUser 同理:这是管理端接口,靠权限码授权(下一步做)。
// 在 BFF 加"只能改自己"会让管理员改不了别人。
//
// ============================================================
// 改字段的语义:指针 nil = 不更新
// ============================================================
//
// .api 里 UpdateUserReq 的五个字段都是指针,故:
//
//	传了 "status": "disabled"  → 更新 status
//	没传 status                → 保持原值
//	传了 "status": ""          → **也更新**,设为空串
//
// 最后一种与前两种的区别只有指针类型能给 —— 这是用指针而不是
// 用非指针 + optional 的原因(go-zero 无法区分"没传"与"传了零值")。
//
// ============================================================
// password 的语义
// ============================================================
//
// 传了就是把口令改成新值(明文透传,服务侧 bcrypt)。
// 不传就不动口令。
//
// **这里不做强度校验** —— 口令规则属于 user-service 的领域知识
// (它知道历史的强度要求与提示文案)。BFF 自己加一套规则,
// 两边必然漂移,而漂移的表现是"这里放过了、那里拒绝"。
//
// ============================================================
// 至少改一个字段
// ============================================================
//
// 全部为 nil 时**不调下游**,直接回 400(见 errNothingToUpdate 的说明)。
// 这与"客户端传了不认识的字段"是同一种情况 ——
// go-zero 的 httpx.Parse 会静默丢弃未声明的字段(不报未知字段错),
// 于是也走到这里。
func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (*types.UpdateUserResp, error) {
	// 字段名用 snake_case,与 sys_user 的列名、model.SysUser 的
	// json tag 一致(单体传 map 时键名就是 JSON 字段名,
	// 由 toProtoFieldValue 直接当 field 用)。
	updates := compactFields(
		strField("username", req.Username),
		strField("phone", req.Phone),
		strField("email", req.Email),
		strField("status", req.Status),
		// 注意 field 名是 "password" 而不是 "password_hash":
		// 单体客户端传的 map 键来自 JSON tag(model.SysUser 的
		// PasswordHash 字段 tag 是 json:"password"),
		// 故 field 名就是 "password"。见 createuserlogic.go 的说明。
		strField("password", req.Password),
	)

	if len(updates) == 0 {
		return nil, errNothingToUpdate
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.UpdateUser(ctx, &v1_userv1.UpdateUserReq{
		UserId:  req.Id,
		Updates: updates,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	// proto 注释:UpdateUserResp.user 返回**合并后的完整对象**,
	// 调用方无需再查一次。故直接映射,不做二次查询。
	return toUpdateUserResp(resp.GetUser()), nil
}
