package deptservicelogic

import (
	"context"
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/mitchellh/mapstructure"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateDeptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDeptLogic {
	return &UpdateDeptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDeptLogic) UpdateDept(in *v1_userv1.UpdateDeptReq) (*v1_userv1.UpdateDeptResp, error) {
	olderDept, err := l.svcCtx.DeptRepo.GetDeptById(in.DeptId)
	if err != nil {
		return &v1_userv1.UpdateDeptResp{ErrorMsg: model.DeptNotExist.Error()}, nil
	}

	// 局部更新:用 FieldUpdate 列表覆盖旧对象,只覆盖传入的字段
	newDept := *olderDept
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newDept,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(converter.FieldUpdatesToMap(in.Updates)); err != nil {
		return &v1_userv1.UpdateDeptResp{ErrorMsg: err.Error()}, nil
	}

	// 改名了要保证同一父节点下不重名
	if newDept.DeptName != olderDept.DeptName {
		existing, _ := l.svcCtx.DeptRepo.GetDeptByUK(newDept.DeptName, newDept.ParentId)
		if existing != nil {
			return &v1_userv1.UpdateDeptResp{ErrorMsg: model.DeptExist.Error()}, nil
		}
	}

	// 换了父节点则排到新父节点末尾
	if newDept.ParentId != olderDept.ParentId {
		sortId, err := l.svcCtx.DeptRepo.GetMaxSortId(newDept.ParentId)
		if err != nil {
			return nil, err
		}
		newDept.SortOrder = sortId
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return l.svcCtx.DeptRepo.WithTx(tx).UpdateDept(&newDept)
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.UpdateDeptResp{Dept: converter.ToProtoDept(&newDept)}, nil
}
