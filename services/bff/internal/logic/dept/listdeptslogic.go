// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package dept

import (
	"context"

	v1_userv1 "demo-shop/api/gen/user/v1"
	"demo-shop/services/bff/internal/converter"
	"demo-shop/services/bff/internal/infra/rpc"
	"demo-shop/services/bff/internal/svc"
	"demo-shop/services/bff/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDeptsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDeptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDeptsLogic {
	return &ListDeptsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ListDepts 分页查询部门(**扁平列表**)。
//
// 响应是 {list, total}(单体 department_handler.go 手工包了
// gin.H{"list": ...}),不是 proto 的 {items, total}。
//
// 过滤参数是 deptType(camelCase)—— 单体用 c.Query("deptType")。
//
// ============================================================
// 列表里 children 为空,要树请用 /tree/:userId
// ============================================================
//
// proto 的 Dept.children 注释:"递归,部门树查询时填充"。
// 故这里的每个元素 children 都是 []。
//
// 部门树那条接口是 GET /admin/dept/tree/:userId ——
// **路径参数是 :userId**,语义是"取这个用户所属的部门树",
// 不是"取以某部门为根的子树"。名字容易误读。
func (l *ListDeptsLogic) ListDepts(req *types.ListDeptsReq) (*types.ListDeptsResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.ListDepts(ctx, &v1_userv1.ListDeptsReq{
		Page:     int32(page),
		PageSize: int32(pageSize),
		DeptType: req.DeptType,
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		return nil, err
	}

	return &types.ListDeptsResp{
		List:  converter.DeptItems(resp.GetItems()),
		Total: resp.GetTotal(),
	}, nil
}
