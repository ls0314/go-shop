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

type CreateDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDeptLogic {
	return &CreateDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateDept 创建部门(管理端)。
//
// ============================================================
// ParentId 传 0 表示根部门
// ============================================================
//
// .api 里 parent_id 是 optional 的 int64,不传时为 0 ——
// 而 0 正好可以表达"根"(顶层部门没有父节点)。
//
// 故这里**不做区分**:不传与传 0 都是"建一个根部门"。
// 这个语义要服务端认可(把 0 当成"无父节点"而不是"父节点ID=0")。
//
// 若实际需要"必须指定父节点",那应当把 parent_id 设为必填
// 并在服务端校验;BFF 不替它决定。
//
// ============================================================
// SortOrder 的默认值由服务端决定
// ============================================================
//
// 不传时为 0(排最前)。BFF **不兜底成"最大值+1"** ——
// 那需要先查一次(多一次 RPC,且是 TOCTOU)。
//
// 而 user-service 在 UpdateDept 换父节点时会自己算
// `GetMaxSortId(新父)+1` 来"挪到末尾" —— 说明"末尾"这个
// 语义是它管的。创建时是否也该如此,由它决定;
// 若前端需要"新建的排在最后",应当显式传 sort_order。
func (l *CreateDeptLogic) CreateDept(req *types.CreateDeptReq) (*types.CreateDeptResp, error) {
	ctx, cancel := rpc.CtxWithTimeout(l.ctx)
	defer cancel()

	resp, grpcErr := l.svcCtx.RBACRPC.CreateDept(ctx, &v1_userv1.CreateDeptReq{
		Dept: &v1_userv1.Dept{
			ParentId:  req.ParentId,
			DeptName:  req.DeptName,
			DeptType:  req.DeptType,
			LeaderId:  req.LeaderId,
			SortOrder: req.SortOrder,
			Status:    req.Status,
			// DeptId / CreatedAt / Children 不传
		},
	})

	kind, err := rpc.Classify(resp.GetErrorMsg(), grpcErr)
	switch kind {
	case rpc.ResultInfra:
		return nil, err
	case rpc.ResultBiz:
		// 同父下重名 / 父部门不存在 → 400
		return nil, err
	}

	// 响应是**裸对象**(单体: utils.Success(c, created))
	return &types.CreateDeptResp{
		DeptItem: converter.DeptItem(resp.GetDept()),
	}, nil
}
