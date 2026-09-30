package deptservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CreateDeptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDeptLogic {
	return &CreateDeptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDeptLogic) CreateDept(in *v1_userv1.CreateDeptReq) (*v1_userv1.CreateDeptResp, error) {
	dept := converter.FromProtoDept(in.Dept)

	// 同一父节点下部门名唯一
	existing, _ := l.svcCtx.DeptRepo.GetDeptByUK(dept.DeptName, dept.ParentId)
	if existing != nil {
		return &v1_userv1.CreateDeptResp{ErrorMsg: model.DeptExist.Error()}, nil
	}

	// 未指定排序时排到该父节点末尾
	if dept.SortOrder == 0 {
		sortId, err := l.svcCtx.DeptRepo.GetMaxSortId(dept.ParentId)
		if err != nil {
			return nil, err
		}
		dept.SortOrder = sortId
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.DeptRepo.WithTx(tx).CreateDept(dept)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.CreateDeptResp{Dept: converter.ToProtoDept(dept)}, nil
}
