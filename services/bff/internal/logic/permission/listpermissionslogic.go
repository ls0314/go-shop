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

type ListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListPermissions 分页查询权限点(管理端)。
//
// 响应是 {list, total},不是 proto 的 {items, total} ——
// 单体 permission_handler.go 是:
//
//	utils.Success(c, gin.H{
//		"list":  permList,
//		"total": total,
//	})
//
// 12 个列表接口全部如此(见 bff.api 里的统一说明)。
//
// 过滤参数是 permission_type(api / menu / button),不是名字模糊查 ——
// 单体就是这么做的(penissionType := c.Query("permissionType"))。
//
// **注意单体那个查询参数名是 camelCase**:c.Query("permissionType")。
// 而 .api 里我按其它接口的一致风格声明成了 snake_case 的
// `form:"permission_type"`。**这是一处可能不一致的地方** ——
// 需要对着单体 handler 核一次前端传的到底是哪个名字。
//
// 若前端传的是 permissionType,而 BFF 只认 permission_type,
// 那么"按类型筛选"会静默失效(不报错,只是筛不出来)——
// 这类 bug 最难发现。故端到端验证时要专门试这一条。
func (l *ListPermissionsLogic) ListPermissions(req *types.ListPermissionsReq) (*types.ListPermissionsResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListPermissions(ctx, &v1_userv1.ListPermissionsReq{
		Page:           int32(page),
		PageSize:       int32(pageSize),
		PermissionType: req.PermissionType,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListPermissionsResp{
		List:  converter.PermissionItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
