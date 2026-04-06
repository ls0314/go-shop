package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"

	"github.com/mitchellh/mapstructure"
)

// PermissionService 权限表服务层实例
type PermissionService struct {
	PermRepo *repository.PermissionRepo // 权限表数据层实例
}

// NewPermissionService 创建权限表服务层实例
// 接收值：permRepo - 权限表数据层实例
// 返回值：*PermissionService - 权限表服务层实例指针
func NewPermissionService(permRepo *repository.PermissionRepo) *PermissionService {
	return &PermissionService{PermRepo: permRepo}
}

// CreatePermission 创建权限
// 接收值：perm - 权限结构体
// 返回值：error - 错误信息
func (s *PermissionService) CreatePermission(perm *model.SysPermission) error {
	// 根据传入权限名判断权限是否存在
	existing, _ := s.PermRepo.GetPermByCode(perm.PermissionCode)
	if existing != nil {
		return model.PermissionExist
	}
	//  调用数据层创建权限
	return s.PermRepo.CreatePerm(perm)
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
func (s *PermissionService) GetPermission(id int64) (*model.SysPermission, error) {
	//调用数据层查询权限信息
	perm, err := s.PermRepo.GetPermByID(id)
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
//	[]model.SysMenu - 分页权限信息列表
//	int64 - 权限总数
//	error - 错误信息
func (s *PermissionService) GetPermissionList(page, pageSize int, permType string) ([]model.SysPermission, int64, error) {
	// 防止参数越界
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	// 调用数据层返回分页权限信息
	return s.PermRepo.GetPermList(page, pageSize, permType)
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
func (s *PermissionService) UpdatePermission(permID int64, updatePerm map[string]interface{}) error {
	// 查询所更新权限是否存在
	olderPerm, err := s.PermRepo.GetPermByID(permID)
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
		existing, _ := s.PermRepo.GetPermByCode(newPerm.PermissionCode)
		if existing != nil {
			return model.PermissionExist
		}
	}
	// 调用数据层更新权限部分信息
	return s.PermRepo.UpdatePerm(&newPerm)
}

// DeletePermission 删除权限
// 接收值：id - 待删除权限唯一标识
// 返回值：error - 错误信息
func (s *PermissionService) DeletePermission(id int64) error {
	// 判断待删除权限是否存在
	existing, _ := s.PermRepo.GetPermByID(id)
	if existing == nil {
		return model.PermissionNotExist
	}
	// 判断待删除权限是否是系统权限
	if existing.IsSystem {
		return model.PermissionIsSystem
	}
	// 判断待删除权限是否还存在角色关联
	hasRel, err := s.PermRepo.CheckRoleRelPerm(id)
	if err != nil {
		return err
	}
	if hasRel {
		return model.PermissionHasRel
	}
	// 调用数据层删除权限
	return s.PermRepo.DeletePerm(id)
}
