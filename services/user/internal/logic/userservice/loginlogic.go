package userservicelogic

import (
	"context"
	"errors"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login 校验凭据并签发令牌。username 与 phone 二选一。
func (l *LoginLogic) Login(in *v1_userv1.LoginReq) (*v1_userv1.LoginResp, error) {
	var (
		user *model.SysUser
		err  error
	)
	switch {
	case in.Username != "":
		user, err = l.svcCtx.UserRepo.GetUserByName(in.Username)
	case in.Phone != "":
		user, err = l.svcCtx.UserRepo.GetUserByPhone(in.Phone)
	default:
		// 两者都空时不能继续:下方会解引用 user
		return &v1_userv1.LoginResp{ErrorMsg: model.UserNotExist.Error()}, nil
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.recordLoginLog(0, in.ClientIp, in.Device, "fail", "user not found")
			return &v1_userv1.LoginResp{ErrorMsg: model.UserNotExist.Error()}, nil
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		l.recordLoginLog(user.UserID, in.ClientIp, in.Device, "fail", "password mismatch")
		return &v1_userv1.LoginResp{ErrorMsg: model.LoginPasswordInvalid.Error()}, nil
	}

	accessToken, err := l.svcCtx.JWT.GenerateAccessToken(user.UserID, user.Username)
	if err != nil {
		return nil, err
	}
	refreshToken, err := l.svcCtx.JWT.GenerateRefreshToken(user.UserID, user.Username)
	if err != nil {
		return nil, err
	}

	// 签发成功后才记成功日志:否则签发失败会留下一条"登录成功"的误导记录
	l.recordLoginLog(user.UserID, in.ClientIp, in.Device, "success", "")

	return &v1_userv1.LoginResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// recordLoginLog 写登录日志。日志是旁路审计,写入失败不影响登录结果。
func (l *LoginLogic) recordLoginLog(userID int64, ip, device, status, reason string) {
	err := l.svcCtx.UserRepo.InsertLoginLog(&model.UserLoginLog{
		UserId:        userID,
		LoginIp:       ip,
		LoginDevice:   device,
		LoginStatus:   status,
		FailureReason: reason,
	})
	if err != nil {
		l.Errorf("写入登录日志失败: userId=%d status=%s err=%v", userID, status, err)
	}
}
