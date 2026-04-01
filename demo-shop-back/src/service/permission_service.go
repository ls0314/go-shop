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
	existing, _ := s.PermRepo.GetPermByCode(perm.PermissionCode)
	if existing != nil {
		return model.PermissionExist
	}

	//perm.IsSystem = false

	return s.PermRepo.CreatePerm(perm)
}

func (s *PermissionService) GetPermission(id int64) (*model.SysPermission, error) {
	perm, err := s.PermRepo.GetPermByID(id)
	if err != nil {
		return nil, model.MenuNotExist
	}
	return perm, nil

}

func (s *PermissionService) GetPermissionList(page, pageSize int, permType string) ([]model.SysPermission, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	return s.PermRepo.GetPermList(page, pageSize, permType)
}

func (s *PermissionService) UpdataPermission(perm *model.SysPermission) error {
	olderPerm, err := s.PermRepo.GetPermByID(perm.PermissionID)
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
		existing, _ := s.PermRepo.GetPermByCode(perm.PermissionCode)
		if existing != nil {
			return model.PermissionExist
		}
	}

	perm.IsSystem = olderPerm.IsSystem
	perm.CreatedAt = olderPerm.CreatedAt

	return s.PermRepo.UpdataPerm(perm)
}

func (s *PermissionService) DeletePermission(id int64) error {
	existing, _ := s.PermRepo.GetPermByID(id)
	if existing == nil {
		return model.PermissionNotExist
	}

	if existing.IsSystem {
		return model.PermissionIsSystem
	}
	hasRel, err := s.PermRepo.CheckRoleRelPerm(id)
	if err != nil {
		return err
	}
	if hasRel {
		return model.PermissionHasRel
	}
	return s.PermRepo.DeletePerm(id)
}
