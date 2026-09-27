package service

import (
	"context"
	"demo-shop-back/db"
	"demo-shop-back/src/infra"
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// PermissionService 权限表服务层实例
type PermissionService struct {
	PermRepo     *repository.PermissionRepo // 权限表数据层实例
	UserRoleRepo *repository.UserRoleRepo
	RolePermRepo *repository.RolePermRepo
	db           *gorm.DB
}

// NewPermissionService 创建权限表服务层实例
// 接收值：permRepo - 权限表数据层实例
// 返回值：*PermissionService - 权限表服务层实例指针
func NewPermissionService() *PermissionService {
	return &PermissionService{
		PermRepo:     repository.NewPermissionRepo(db.DB),
		UserRoleRepo: repository.NewUserRoleRepo(db.DB),
		RolePermRepo: repository.NewRolePermRepo(db.DB),
		db:           db.DB,
	}
}

// CreatePermission 创建权限
// 接收值：perm - 权限结构体
// 返回值：error - 错误信息
func (p *PermissionService) CreatePermission(perm *model.SysPermission) error {
	// 根据传入权限编码联合判断权限是否存在
	existing, _ := p.PermRepo.GetPermByCodeUk(perm.ApiPath, perm.RequestMethod, perm.PermissionCode)
	if existing != nil {
		return model.PermissionExist
	}

	err := p.db.Transaction(func(tx *gorm.DB) error {
		permTxRepo := p.PermRepo.WithTx(tx)

		if err := permTxRepo.CreatePerm(perm); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	// 权限映射变更：版本号+1，使所有 api:perm 缓存即时失效
	if cache := infra.GetCache(); cache != nil {
		_, _ = cache.Incr(context.Background(), "api:perm:version")
	}
	return nil

}

// GetPermission 查询权限信息（根据权限ID）
// 接收值：
//
//	id - 所查询权限唯一标识
//
// 返回值：
//
//	*model.SysPermission - 权限对象指针
//	error - 错误信息
func (p *PermissionService) GetPermission(id int64) (*model.SysPermission, error) {
	//调用数据层查询权限信息
	perm, err := p.PermRepo.GetPermByID(id)
	if err != nil {
		return nil, model.MenuNotExist
	}
	return perm, nil

}

// GetPermissionList 分页查询权限信息（可根据权限类型查询）
// 接收值：
//
//	page - 页数,
//	pageSize - 页大小
//	permType - 权限类型
//
// 返回值:
//
//	[]model.SysPermission - 分页权限信息列表
//	int64 - 权限总数
//	error - 错误信息
func (p *PermissionService) GetPermissionList(page, pageSize int, permType string) ([]model.SysPermission, int64, error) {
	// 防止参数越界
	if page <= 0 {
		page = 1
	}
	// 防参数越界:<=0 用默认 10;>100 封顶 100(而非压成 10,避免大 pageSize 反而返回最少)
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	// 调用数据层返回分页权限信息
	return p.PermRepo.GetPermList(page, pageSize, permType)
}

// UpdatePermission 更新权限信息
// 接收值：
//
//	permID - 更新权限唯一标识
//	updatePerm - 更新权限部分信息
//
// 返回值：
//
//	error - 错误信息
func (p *PermissionService) UpdatePermission(permID int64, updatePerm map[string]interface{}) error {
	// 查询所更新权限是否存在
	olderPerm, err := p.PermRepo.GetPermByID(permID)
	if err != nil {
		return model.PermissionNotExist
	}
	// 禁止修改系统权限信息
	if olderPerm.IsSystem {
		return model.PermissionIsSystem
	}
	// 用原权限信息初始化更新权限信息
	newPerm := *olderPerm
	// 配置mapstructure，用updatePerm覆盖更新权限信息（只覆盖传入字段）
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &newPerm,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	if err := decoder.Decode(updatePerm); err != nil {
		return err
	}

	// 保证更新的权限名未被使用
	if newPerm.PermissionCode != olderPerm.PermissionCode {
		return model.PermissionCodeNotAlter
	}

	// 开启事务
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	permTxRepo := p.PermRepo.WithTx(tx)
	// 调用数据层更新权限
	if err := permTxRepo.UpdatePerm(&newPerm); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	err = tx.Commit().Error

	if err != nil {
		return err
	}
	// 权限映射变更：版本号+1，使所有 api:perm 缓存即时失效
	if cache := infra.GetCache(); cache != nil {
		_, _ = cache.Incr(context.Background(), "api:perm:version")
	}
	return nil

}

// DeletePermission 删除权限
// 接收值：id - 待删除权限唯一标识
// 返回值：error - 错误信息
func (p *PermissionService) DeletePermission(id int64) error {
	// 判断待删除权限是否存在
	existing, _ := p.PermRepo.GetPermByID(id)
	if existing == nil {
		return model.PermissionNotExist
	}
	// 判断待删除权限是否是系统权限
	if existing.IsSystem {
		return model.PermissionIsSystem
	}
	// 判断待删除权限是否还存在角色关联
	hasRel, err := p.PermRepo.CheckRoleRelPerm(id)
	if err != nil {
		return err
	}
	if hasRel {
		return model.PermissionHasRel
	}

	// 开启事务
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建事务实例
	permTxRepo := p.PermRepo.WithTx(tx)
	// 调用数据层删除权限
	if err := permTxRepo.DeletePerm(id); err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	err = tx.Commit().Error

	if err != nil {
		return err
	}
	// 权限映射变更：版本号+1，使所有 api:perm 缓存即时失效
	if cache := infra.GetCache(); cache != nil {
		_, _ = cache.Incr(context.Background(), "api:perm:version")
	}
	return nil
}
