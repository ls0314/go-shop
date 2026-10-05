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

type GetSelfPermsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSelfPermsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSelfPermsLogic {
	return &GetSelfPermsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetSelfPerms 取当前登录用户的全部权限码(按钮级权限展示用)。
//
// ============================================================
// 取数走 RBACService,不是 UserService
// ============================================================
//
// 两个 gRPC service 由**同一个 user-service 进程**提供、同一个 etcd
// key 发现,故用的是同一条连接(见 svc/servicecontext.go)—— 但
// 方法名与响应类型不同:
//
//	h.svcCtx.UserRPC.ListSelfPermCodes   ← 在 UserService 上
//	h.svcCtx.RBACRPC.ListPermCodesByUserId ← 在 RBACService 上
//
// 单体用的是前者(userclient.ListSelfPermCodes),**这里保持一致** ——
// 两者语义有细微差别:ListSelfPermCodes 是"我自己的权限码"
// (面向当前登录用户),而 ListPermCodesByUserId 是可以查任意用户
// 的判权数据源。用后者需要额外做归属校验,而前者由服务端按
// 传入的 userId 直接取,不多一层语义。
//
// ============================================================
// 响应键名是 perms,不是 perm_codes
// ============================================================
//
// proto 的字段叫 perm_codes,而单体 handler 手工重命名了:
//
//	if codes == nil { codes = []string{} }
//	utils.Success(c, gin.H{"perms": codes})
//
// 前端解的是 res.data.perms。**且空值必须是 [] 而不是 null** ——
// 那句 `if codes == nil` 就是为此:前端有 `perms.length` 这类直接
// 取属性的写法,null 会抛错。
func (l *GetSelfPermsLogic) GetSelfPerms(req *types.Empty) (*types.GetSelfPermsResp, error) {
	userId, ok := middleware.UserID(l.ctx)
	if !ok {
		return nil, errNoIdentity
	}

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.UserRPC.ListSelfPermCodes(ctx, &v1_userv1.ListSelfPermCodesReq{
		UserId: userId,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	codes := resp.GetPermCodes()
	// 保证是 [] 而不是 null(见上方注释)。
	// 无任何权限时 proto 返回空切片,但显式兜一次能防住
	// "服务端返回 nil 切片"的边界 —— 那时 len() 仍是 0,
	// 但 JSON 会变成 null。
	if codes == nil {
		codes = []string{}
	}

	return &types.GetSelfPermsResp{Perms: codes}, nil
}
