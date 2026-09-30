package deptservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteDeptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDeptLogic {
	return &DeleteDeptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDeptLogic) DeleteDept(in *v1_userv1.DeleteDeptReq) (*v1_userv1.DeleteDeptResp, error) {
	existing, _ := l.svcCtx.DeptRepo.GetDeptById(in.DeptId)
	if existing == nil {
		return &v1_userv1.DeleteDeptResp{ErrorMsg: model.DeptNotExist.Error()}, nil
	}

	// 有用户关联的部门不允许删除
	hasRel, err := l.svcCtx.DeptRepo.CheckDeptRelUser(in.DeptId)
	if err != nil {
		return nil, err
	}
	if hasRel {
		return &v1_userv1.DeleteDeptResp{ErrorMsg: model.DeptHasRel.Error()}, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.DeptRepo.WithTx(tx).DeleteDept(in.DeptId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.DeleteDeptResp{}, nil
}
