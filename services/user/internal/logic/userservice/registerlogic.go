package userservicelogic

import (
	"context"
	"errors"
	"regexp"

	"demo-shop/api/gen/user/v1"
	"demo-shop/services/user/internal/converter"
	"demo-shop/services/user/internal/model"
	"demo-shop/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	reLetter = regexp.MustCompile(`[A-Za-z]`)
	reNumber = regexp.MustCompile(`[0-9]`)
	reSymbol = regexp.MustCompile(`[^\w\s]`)
	rePhone  = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// ValidatePassword 口令需同时含字母、数字、特殊符号且长度 >= 8。
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	return reLetter.MatchString(password) && reNumber.MatchString(password) && reSymbol.MatchString(password)
}

// ValidatePhone 中国大陆手机号。
func ValidatePhone(phone string) bool {
	return rePhone.MatchString(phone)
}

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 注册用户并创建默认档案,两步在同一事务内完成。
func (l *RegisterLogic) Register(in *v1_userv1.RegisterReq) (*v1_userv1.RegisterResp, error) {
	if !ValidatePassword(in.Password) {
		return &v1_userv1.RegisterResp{ErrorMsg: model.RegPasswordInvalid.Error()}, nil
	}
	if !ValidatePhone(in.Phone) {
		return &v1_userv1.RegisterResp{ErrorMsg: model.PhoneMalformed.Error()}, nil
	}

	// 三项唯一性检查。查不到是正常路径,只把"查到了"当冲突。
	if taken, err := l.existsByName(in.Username); err != nil {
		return nil, err
	} else if taken {
		return &v1_userv1.RegisterResp{ErrorMsg: model.UsernameExist.Error()}, nil
	}
	if taken, err := l.existsByPhone(in.Phone); err != nil {
		return nil, err
	} else if taken {
		return &v1_userv1.RegisterResp{ErrorMsg: model.PhoneExist.Error()}, nil
	}
	if in.Email != "" {
		if taken, err := l.existsByEmail(in.Email); err != nil {
			return nil, err
		} else if taken {
			return &v1_userv1.RegisterResp{ErrorMsg: model.EmailExist.Error()}, nil
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.SysUser{
		Username:     in.Username,
		PasswordHash: string(hash),
		Email:        in.Email,
		Phone:        in.Phone,
		Status:       "active",
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if err := l.svcCtx.UserRepo.WithTx(tx).CreateUser(user); err != nil {
			return err
		}
		nickname := in.Nickname
		if nickname == "" {
			nickname = in.Username
		}
		return l.svcCtx.UserProfileRepo.WithTx(tx).CreateProfile(&model.UserProfile{
			UserId:   user.UserID,
			Nickname: nickname,
		})
	})
	if err != nil {
		return nil, err
	}

	return &v1_userv1.RegisterResp{User: converter.ToProtoUser(user)}, nil
}

func (l *RegisterLogic) existsByName(name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	_, err := l.svcCtx.UserRepo.GetUserByName(name)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (l *RegisterLogic) existsByPhone(phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}
	_, err := l.svcCtx.UserRepo.GetUserByPhone(phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (l *RegisterLogic) existsByEmail(email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	_, err := l.svcCtx.UserRepo.GetUserByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
