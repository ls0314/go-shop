// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package permission

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePermissionLogic {
	return &CreatePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreatePermission 创建权限点(管理端)。
//
// ============================================================
// permission_code 是判权的键,建错了会让接口判权失效
// ============================================================
//
// 权限点的三要素(proto 注释):
//
//	permission_code  唯一编码,如 "platform:coupon:create"
//	permission_type  api / menu / button
//	request_method   仅 api 类型有意义(GET / POST / PUT / DELETE)
//	api_path         仅 api 类型有意义,与 sys_permission.api_path 一致,
//	                 **带 /api/v1 前缀**
//
// 判权链路是:`c.FullPath() + Method` → 反查持这个 api_path+method 的
// 权限码 → 看用户是否有它。所以 api_path 必须与路由模板**逐字一致**
// (含 :id 这类参数名)。
//
// **BFF 不在这里校验 api_path 的格式**:
//
//	① 前缀、参数名的规范属于服务端的领域知识;
//	② 客户端传错会在判权时暴露(那条接口对谁都 403),
//	   比在创建时报错更容易定位 (因为是"某个管理员建错了")
//
// 但 BFF 也没有理由去"修正"它(比如自动补 /api/v1 前缀)——
// 那会让同一个值在两处被不同地解释。
//
// ============================================================
// is_system 不可指定
// ============================================================
//
// 与 CreateRole 同理:系统内置权限由种子数据创建,
// 不让 API 指定 —— 否则管理员能造出一个"不可改不可删"的权限点
// 把自己锁死。
func (l *CreatePermissionLogic) CreatePermission(req *types.CreatePermissionReq) (*types.CreatePermissionResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.CreatePermission(ctx, &v1_userv1.CreatePermissionReq{
		Permission: &v1_userv1.Permission{
			PermissionCode: req.PermissionCode,
			PermissionName: req.PermissionName,
			PermissionType: req.PermissionType,
			RequestMethod:  req.RequestMethod,
			ApiPath:        req.ApiPath,
			Description:    req.Description,
			// IsSystem 与 PermissionId 不传(见上方说明 / 服务端生成)
		},
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 编码重复等 → 400
		return nil, err
	}

	// 响应是**裸权限对象**(内嵌),不是 {permission: {...}} ——
	// 单体: utils.Success(c, perm)。见 getrolelogic.go 的同类说明。
	item := converter.PermissionItem(resp.GetPermission())
	return &types.CreatePermissionResp{PermissionItem: item}, nil
}
