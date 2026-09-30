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

// FromProtoPerm 把 proto 入参转成 model。created_at 由数据库生成,不从入参取。
func FromProtoPerm(p *v1_userv1.Permission) *model.SysPermission {
	if p == nil {
		return &model.SysPermission{}
	}
	return &model.SysPermission{
		PermissionCode: p.GetPermissionCode(),
		PermissionName: p.GetPermissionName(),
		PermissionType: p.GetPermissionType(),
		RequestMethod:  p.GetRequestMethod(),
		ApiPath:        p.GetApiPath(),
		Description:    p.GetDescription(),
		IsSystem:       p.GetIsSystem(),
	}
}
