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

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateUser 管理员建号。
//
// ============================================================
// 与 Register(自助注册)的区别
// ============================================================
//
// 单体 RegisterHandler 的注释写明了注册会额外做:
//
//	"校验手机号格式、查重用户名/手机号/邮箱,并在同一事务内
//	 创建 user_profile(nickname 为空时回落为用户名)"
//
// 而这条**不做这些** —— 它是管理员建号,不建档案、不做注册级查重。
// 所以两条路由的服务端行为不同,BFF 不能合并,也不该在这里
// "顺手"加上注册的校验。
//
// ============================================================
// password 的字段名陷阱
// ============================================================
//
// proto 的 CreateUserReq 有两个字段:
//
//	User     *User   // 用户快照
//	Password string  // 明文口令,服务侧做 bcrypt 哈希后入库
//
// 而单体的 model.SysUser 里,哈希字段的 json tag 是 **"password"**:
//
//	PasswordHash string `gorm:"column:password_hash" json:"password"`
//
// 于是单体 handler 收 JSON 时:
//
//	var req model.SysUser
//	c.ShouldBindJSON(&req)            // 客户端传 {"password": "明文"}
//	u.userRPC.CreateUser(&req, req.PasswordHash)   // ← 命名与语义不符
//
// 即:客户端传的**明文口令**被绑定进了一个叫 PasswordHash 的字段,
// 然后当成明文传给了 RPC(password 参数,服务侧再哈希)。
//
// 字段名误导,但**行为是对的**(到服务端时确实是明文,由它哈希)。
// 故 .api 里 types.CreateUserReq.Password 收明文,与既有行为一致。
//
// **不要**因为看到 "PasswordHash" 就在 BFF 这边哈希一次 ——
// 那会导致服务端对已哈希的值再哈希,用户永远登不进来。
//
// ============================================================
// status 的可选性
// ============================================================
//
// .api 里 status 是 optional。不传时传空串给下游,由 user-service
// 决定默认值(proto 注释:active / disabled)。
//
// BFF **不在这里补默认值**:默认值是"新用户应当是启用还是禁用"
// 的业务决策,属于服务端的领域知识。若 BFF 兜一个 "active",
// 而服务端后来改成默认 disabled,两边就不一致了。
func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (*types.CreateUserResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.CreateUser(ctx, &v1_userv1.CreateUserReq{
		User: &v1_userv1.User{
			Username: req.Username,
			Email:    req.Email,
			Phone:    req.Phone,
			Status:   req.Status,
			// UserId 与 CreatedAt/UpdatedAt 不传:
			// ID 由服务端生成(雪花或自增),时间由 DB 填。
		},
		// 明文,服务侧 bcrypt(见上方说明)
		Password: req.Password,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 用户名/手机号已存在等 → 400
		return nil, err
	}

	// data 是 proto User 本体(单体: utils.Success(c, user))
	return toCreateUserResp(resp.GetUser()), nil
}
