package converter

import (
	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProtoUser 把 model 转成 proto 快照。
// 不含 password_hash / failed_attempts / lock_until 等凭据与内部状态字段。
func ToProtoUser(u *model.SysUser) *v1_userv1.User {
	if u == nil {
		return nil
	}
	return &v1_userv1.User{
		UserId:    u.UserID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		Status:    u.Status,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

// ToProtoUserProfile 把 model 转成 proto 快照。
func ToProtoUserProfile(p *model.UserProfile) *v1_userv1.UserProfile {
	if p == nil {
		return nil
	}
	var birthdate *timestamppb.Timestamp
	if p.Birthdate != nil {
		birthdate = timestamppb.New(*p.Birthdate)
	}
	return &v1_userv1.UserProfile{
		UserInfoId: p.UserInfoID,
		UserId:     p.UserId,
		Nickname:   p.Nickname,
		RealName:   p.RealName,
		Gender:     p.Gender,
		AvatarUrl:  p.AvatarURL,
		Birthdate:  birthdate,
		CreatedAt:  timestamppb.New(p.CreatedAt),
		UpdatedAt:  timestamppb.New(p.UpdatedAt),
	}
}
