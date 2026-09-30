package deptservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDeptsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDeptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDeptsLogic {
	return &ListDeptsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListDeptsLogic) ListDepts(in *v1_userv1.ListDeptsReq) (*v1_userv1.ListDeptsResp, error) {
	page := int(in.Page)
	if page <= 0 {
		page = 1
	}

	pageSize := int(in.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	deptList, total, err := l.svcCtx.DeptRepo.GetDeptList(page, pageSize, in.DeptType)
	if err != nil {
		return nil, err
	}

	items := make([]*v1_userv1.Dept, 0, len(deptList))
	for i := range deptList {
		items = append(items, converter.ToProtoDept(&deptList[i]))
	}

	return &v1_userv1.ListDeptsResp{
		Items: items,
		Total: total,
	}, nil
}
