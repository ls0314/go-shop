package converter

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
)

func ToProtoPermission(p *model.SysPermission) *v1_userv1.Permission {
	if p == nil {
		return nil
	}
	return &v1_userv1.Permission{
		PermissionId:   p.PermissionID,
		PermissionCode: p.PermissionCode,
		PermissionName: p.PermissionName,
		PermissionType: p.PermissionType,
		RequestMethod:  p.RequestMethod,
		ApiPath:        p.ApiPath,
		Description:    p.Description,
		IsSystem:       p.IsSystem,
	}
}
