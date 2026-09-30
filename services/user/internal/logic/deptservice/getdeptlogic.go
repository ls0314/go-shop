package deptservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptLogic {
	return &GetDeptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDeptLogic) GetDept(in *v1_userv1.GetDeptReq) (*v1_userv1.GetDeptResp, error) {
	dept, err := l.svcCtx.DeptRepo.GetDeptById(in.DeptId)
	if err != nil {
		// 业务失败:error 返回 nil,原因放 error_msg
		return &v1_userv1.GetDeptResp{ErrorMsg: model.DeptNotExist.Error()}, nil
	}

	return &v1_userv1.GetDeptResp{Dept: converter.ToProtoDept(dept)}, nil
}
