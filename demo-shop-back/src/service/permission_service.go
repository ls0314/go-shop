package service

import (
	"demo-shop-back/src/model"
	"demo-shop-back/src/repository"
)

type PermissionService struct {
	PermRepo *repository.PermissionRepo
}

func NewPermissionService(permRepo *repository.PermissionRepo) *PermissionService {
	return &PermissionService{PermRepo: permRepo}
}

func (s *PermissionService) CreatePermission(perm *model.SysPermission) error {
	existing, _ := s.PermRepo.GetByCode(perm.PermissionCode)
	if existing != nil {
		return model.PermissionExist
	}

	//perm.IsSystem = false

	return s.PermRepo.Create(perm)
}

func (s *PermissionService) GetPermission(id int64) (*model.SysPermission, error) {
	return s.PermRepo.GetByID(id)

}

func (s *PermissionService) GetPermissionList(page, pageSize int, permType string) ([]model.SysPermission, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	return s.PermRepo.List(page, pageSize, permType)
}

func (s *PermissionService) UpdataPermission(perm *model.SysPermission) error {
	olderPerm, err := s.PermRepo.GetByID(perm.PermissionID)
	if err != nil {
		return model.PermissionNotExist
	}

	if olderPerm.IsSystem {
		return model.PermissionIsSystem
	}

	if perm.PermissionCode == "" {
		perm.PermissionCode = olderPerm.PermissionCode
	}

	if perm.PermissionCode != olderPerm.PermissionCode {
		existing, _ := s.PermRepo.GetByCode(perm.PermissionCode)
		if existing != nil {
			return model.PermissionExist
		}
	}

	perm.IsSystem = olderPerm.IsSystem
	perm.CreatedAt = olderPerm.CreatedAt

	return s.PermRepo.Updata(perm)
}

func (s *PermissionService) DeletePermission(id int64) error {
	existing, _ := s.PermRepo.GetByID(id)
	if existing == nil {
		return model.PermissionNotExist
	}

	if existing.IsSystem {
		return model.PermissionIsSystem
	}
	hasRel, err := s.PermRepo.CheckRoleRel(id)
	if err != nil {
		return err
	}
	if hasRel {
		return model.PermissionHasRel
	}
	return s.PermRepo.Delete(id)
}
