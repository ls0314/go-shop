package converter

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoRole 把 model 转成 proto 快照。
func ToProtoRole(r *model.SysRole) *v1_userv1.Role {
	if r == nil {
		return nil
	}
	return &v1_userv1.Role{
		RoleId:      r.RoleId,
		RoleName:    r.RoleName,
		RoleType:    r.RoleType,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		IsDefault:   r.IsDefault,
		DataScope:   r.DataScope,
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}
}

// FromProtoRole 把 proto 入参转成 model。created_at 由数据库生成,不从入参取。
func FromProtoRole(r *v1_userv1.Role) *model.SysRole {
	if r == nil {
		return &model.SysRole{}
	}
	return &model.SysRole{
		RoleId:      r.GetRoleId(),
		RoleName:    r.GetRoleName(),
		RoleType:    r.GetRoleType(),
		Description: r.GetDescription(),
		IsSystem:    r.GetIsSystem(),
		IsDefault:   r.GetIsDefault(),
		DataScope:   r.GetDataScope(),
	}
}

// FieldUpdatesToMap 把 proto 的 FieldUpdate 列表转成 map,供 mapstructure 覆盖。
func FieldUpdatesToMap(updates []*v1_userv1.FieldUpdate) map[string]interface{} {
	out := make(map[string]interface{}, len(updates))
	for _, u := range updates {
		switch v := u.GetValue().GetValue().(type) {
		case *v1_userv1.FieldValue_StringValue:
			out[u.GetField()] = v.StringValue
		case *v1_userv1.FieldValue_Int64Value:
			out[u.GetField()] = v.Int64Value
		case *v1_userv1.FieldValue_BoolValue:
			out[u.GetField()] = v.BoolValue
		}
	}
	return out
}
