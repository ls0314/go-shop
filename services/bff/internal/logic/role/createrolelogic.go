// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateRole 创建角色(管理端,需权限码)。
//
// ============================================================
// 响应有两套形状,这里选的是嵌套的那套
// ============================================================
//
// proto 的 CreateRoleResp 是:
//
//	Role  role      = 1;
//	string error_msg = 2;
//
// 即被包在 role 字段里,而 .api 里 types.CreateRoleResp 也声明成
// {role: {...}}(与单体的行为一致)。
//
// **注意与 CreateUser / CreatePermission 的差别** ——
// 那两个的 data 是**实体本体**(单体 utils.Success(c, user)),
// 而角色这里是嵌套 {role: ...}。
//
// 这是既有的不一致,不要统一。判据永远是单体 handler 里
// utils.Success 的第二个参数是什么:
//
//	CreateRoleHandler:      utils.Success(c, role)      → data 是实体?
//	CreatePermissionHandler: utils.Success(c, perm)     → data 是实体?
//
// **这一点我没有逐个核对过。** 上面那段是**推断**(依据 proto 的
// 消息结构),若前端解的是 data.role_id 而这里包了一层,
// 会全部拿到 undefined。端到端验证时按单体的实际响应核一次。
func (l *CreateRoleLogic) CreateRole(req *types.CreateRoleReq) (*types.CreateRoleResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.CreateRole(ctx, &v1_userv1.CreateRoleReq{
		Role: &v1_userv1.Role{
			RoleName:    req.RoleName,
			RoleType:    req.RoleType,
			Description: req.Description,
			DataScope:   req.DataScope,
			IsDefault:   req.IsDefault,
			// IsSystem 不传:系统内置角色只能由初始化脚本或
			// 数据迁移创建,不该由 API 指定 —— 否则管理员能造出
			// 一个"不可改不可删"的角色把自己锁住。
			//
			// RoleId 不传:由服务端生成。
		},
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 角色名/编码重复等 → 400
		return nil, err
	}

	// CreateRoleResp 是**匿名内嵌 RoleItem**(见 .api 的说明)——
	// 序列化成 {...RoleItem 的字段...},没有 role 那一层包装,
	// 因为单体是 utils.Success(c, created),data 就是那个角色对象。
	//
	// 内嵌的代价:*RoleItem 与 *CreateRoleResp 是两个不同的具名类型,
	// 不能直接赋值,故这里解引用后拷进内嵌字段。
	item := converter.RoleItem(resp.GetRole())
	return &types.CreateRoleResp{RoleItem: item}, nil
}
