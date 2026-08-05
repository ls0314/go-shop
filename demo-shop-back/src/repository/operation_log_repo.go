package repository

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"demo-shop-back/src/model/requset"
	"time"

	"gorm.io/gorm"
)

type OperationLogRepo struct {
	db *gorm.DB
}

func NewOperationLogRepo() *OperationLogRepo {
	return &OperationLogRepo{
		db: db.DB,
	}
}

func (o *OperationLogRepo) CreateOperationLog(req *model.OperationLog) error {
	return o.db.Create(req).Error
}

func (o *OperationLogRepo) GetOperationLogList(req requset.GetOperationLogList) ([]*model.OperationLog, int64, error) {
	var resp []*model.OperationLog
	baseQuery := o.db.Model(&model.OperationLog{})

	if req.RequestMethod != "" {
		baseQuery = baseQuery.Where("request_method = ?", req.RequestMethod)
	}
	if req.Module != "" {
		baseQuery = baseQuery.Where("module = ?", req.Module)
	}
	if req.StartTime != nil {
		baseQuery = baseQuery.Where("created_at >= ?", req.StartTime)
	}

	if req.EndTime != nil {
		endOfDay := req.EndTime.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
		baseQuery = baseQuery.Where("created_at <= ?", endOfDay)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	pageSize := req.PageSize
	offset := (page - 1) * pageSize

	err := baseQuery.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&resp).Error
	if err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}
