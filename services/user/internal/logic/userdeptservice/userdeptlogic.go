package userdeptservicelogic

import (
	"context"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AssignUserDeptsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignUserDeptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserDeptsLogic {
	return &AssignUserDeptsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AssignUserDepts 全量替换用户的部门绑定:先清空再写入。
// primary_index 是 dept_ids 中的下标,越界或为负时无主部门。
func (l *AssignUserDeptsLogic) AssignUserDepts(in *v1_userv1.AssignUserDeptsReq) (*v1_userv1.AssignUserDeptsResp, error) {
	if _, err := l.svcCtx.UserRepo.GetUserById(in.UserId); err != nil {
		return &v1_userv1.AssignUserDeptsResp{ErrorMsg: model.UserNotExist.Error()}, nil
	}

	for _, deptId := range in.DeptIds {
		if _, err := l.svcCtx.DeptRepo.GetDeptById(deptId); err != nil {
			return &v1_userv1.AssignUserDeptsResp{ErrorMsg: model.DeptNotExist.Error()}, nil
		}
	}

	// 下标越界视为无主部门,避免写入一条 is_primary 全为 false 之外的脏数据
	primaryIndex := in.PrimaryIndex
	if primaryIndex < 0 || int(primaryIndex) >= len(in.DeptIds) {
		primaryIndex = -1
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		repo := l.svcCtx.UserDeptRepo.WithTx(tx)
		if err := repo.DeleteByUserId(in.UserId); err != nil {
			return err
		}
		return repo.CreateUserDept(in.UserId, in.DeptIds, primaryIndex)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.AssignUserDeptsResp{}, nil
}

type ListUserDeptsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserDeptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserDeptsLogic {
	return &ListUserDeptsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserDeptsLogic) ListUserDepts(in *v1_userv1.ListUserDeptsReq) (*v1_userv1.ListUserDeptsResp, error) {
	deptIds, err := l.svcCtx.UserDeptRepo.ListDeptIdsBySingleUserId(in.UserId)
	if err != nil {
		return nil, err
	}

	primaryDeptId, err := l.svcCtx.UserDeptRepo.GetPrimaryDeptId(in.UserId)
	if err != nil {
		return nil, err
	}

	if len(deptIds) == 0 {
		return &v1_userv1.ListUserDeptsResp{
			Items:         []*v1_userv1.Dept{},
			PrimaryDeptId: primaryDeptId,
		}, nil
	}

	depts, err := l.svcCtx.DeptRepo.ListDeptByIds(deptIds)
	if err != nil {
		return nil, err
	}

	return &v1_userv1.ListUserDeptsResp{
		Items:         converter.ToProtoDeptTree(depts),
		Total:         int64(len(depts)),
		PrimaryDeptId: primaryDeptId,
	}, nil
}

type ClearUserDeptsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearUserDeptsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearUserDeptsLogic {
	return &ClearUserDeptsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearUserDeptsLogic) ClearUserDepts(in *v1_userv1.ClearUserDeptsReq) (*v1_userv1.ClearUserDeptsResp, error) {
	if _, err := l.svcCtx.UserRepo.GetUserById(in.UserId); err != nil {
		return &v1_userv1.ClearUserDeptsResp{ErrorMsg: model.UserNotExist.Error()}, nil
	}

	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.UserDeptRepo.WithTx(tx).DeleteByUserId(in.UserId)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.ClearUserDeptsResp{}, nil
}
