package converter

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoScope 把 model 转成 proto。
func ToProtoScope(s *model.SysScope) *v1_userv1.Scope {
	if s == nil {
		return nil
	}
	return &v1_userv1.Scope{
		ScopeId:        s.ScopeId,
		RoleId:         s.RoleId,
		ResourceType:   s.ResourceType,
		FieldName:      s.FieldName,
		ConditionType:  s.ConditionType,
		ConditionValue: s.ConditionValue,
		Description:    s.Description,
		CreatedAt:      timestamppb.New(s.CreatedAt),
	}
}

// FromProtoScope 把 proto 入参转成 model。created_at 由数据库生成,不从入参取。
func FromProtoScope(p *v1_userv1.Scope) *model.SysScope {
	if p == nil {
		return &model.SysScope{}
	}
	return &model.SysScope{
		ScopeId:        p.GetScopeId(),
		RoleId:         p.GetRoleId(),
		ResourceType:   p.GetResourceType(),
		FieldName:      p.GetFieldName(),
		ConditionType:  p.GetConditionType(),
		ConditionValue: p.GetConditionValue(),
		Description:    p.GetDescription(),
	}
}
