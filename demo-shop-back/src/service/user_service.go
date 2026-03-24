package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/middleware"
	"demo-shop-back/src/model"
	"regexp"
	"time"
)
import "golang.org/x/crypto/bcrypt"

func ValidatePassword(password string) bool {

	if len(password) < 8 {
		return false
	}

	hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[^\w\s]`).MatchString(password)

	return hasLetter && hasNumber && hasSymbol
}

func ValidatePhone(phone string) bool {
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`).MatchString(phone)
	return reg
}

func Register(req model.RegisterRequest) error {
	//log.Printf("Register called with phone=%s, username=%s", req.Phone, req.Username)
	if !ValidatePassword(req.Password) {
		return model.RegPasswordInvalid
	}

	if !ValidatePhone(req.Phone) {
		return model.PhoneMalformed
	}

	database := db.GetDB()

	var exist model.SysUser

	if err := database.Where("phone = ?", req.Phone).First(&exist).Error; err == nil {
		return model.PhoneExist
	}

	if err := database.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		return model.UsernameExist
	}

	if err := database.Where("email = ?", req.Email).First(&exist).Error; err == nil {
		return model.EmailExist
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx := db.DB.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()
	RegSql := `INSERT INTO sys_user
	(username,password_hash,email,phone,status,created_at,updated_at) 
	VALUES (?,?,?,?,'active',?,?)` // 第一次创建时间就是第一次更新时间 更新时间初始化

	if err := tx.Exec(RegSql,
		req.Username,
		string(passwordHash),
		req.Email,
		req.Phone,
		now,
		now,
	).Error; err != nil {
		tx.Rollback()
		return err
	}

	UserInfoSql := `INSERT INTO user_profile
	(nickname,created_at,updated_at)
	VALUES (?,?,?)`

	if err := tx.Exec(
		UserInfoSql,
		req.Nickname,
		now,
		now,
	).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func recordLoginLog(userID int64, ip string, device string, status string, reason string) {

	db.DB.Exec(`
	INSERT INTO user_login_log
	(user_id, login_ip, login_device, login_status, failure_reason)
	VALUES (?, ?, ?, ?, ?)
	`, userID, ip, device, status, reason)
}

func Login(req model.LoginRequest, ip string, device string) (*model.LoginResponse, error) {

	var user model.UserLoginInfo

	UserLoginSql := `SELECT user_id, username, password_hash 
					FROM sys_user 
					WHERE username = ?`

	err := db.DB.Raw(UserLoginSql, req.Username).Scan(&user).Error

	if err != nil || user.UserID == 0 {
		recordLoginLog(0, ip, device, "fail", "user not found")
		return nil, model.LoginPasswordInvalid
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, model.LoginPasswordInvalid
	}

	jwtService := middleware.GetJWTService()
	accessToken, err := jwtService.GenerateAccessToken(user.UserID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwtService.GenerateRefreshToken(user.UserID, user.Username)
	if err != nil {
		return nil, err
	}

	LoginInfoSql := `UPDATE sys_user
					SET last_login_time = NOW(),
					last_login_ip = ?
					WHERE user_id = ?`

	db.DB.Exec(LoginInfoSql, ip, user.UserID)

	recordLoginLog(user.UserID, ip, device, "success", "")

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
