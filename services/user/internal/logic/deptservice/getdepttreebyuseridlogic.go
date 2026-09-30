package deptservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptTreeByUserIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeptTreeByUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptTreeByUserIdLogic {
	return &GetDeptTreeByUserIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDeptTreeByUserIdLogic) GetDeptTreeByUserId(in *v1_userv1.GetDeptTreeByUserIdReq) (*v1_userv1.GetDeptTreeByUserIdResp, error) {
	// 用户 -> 所属部门ID -> 部门,再组装成树
	deptIds, err := l.svcCtx.DeptRepo.ListDeptIdsByUserId(in.UserId)
	if err != nil {
		return nil, err
	}
	// 无部门归属视为业务失败(与菜单树返回空数组的语义不同)
	if len(deptIds) == 0 {
		return &v1_userv1.GetDeptTreeByUserIdResp{ErrorMsg: model.DeptNotExist.Error()}, nil
	}

	deptList, err := l.svcCtx.DeptRepo.ListDeptByIds(deptIds)
	if err != nil {
		return nil, err
	}

	return &v1_userv1.GetDeptTreeByUserIdResp{
		Items: converter.ToProtoDeptTree(MakeTree(deptList)),
	}, nil
}
