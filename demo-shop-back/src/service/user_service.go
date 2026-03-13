package service

import (
	"demo-shop-back/db"
	"demo-shop-back/src/model"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/google/uuid"
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
	log.Printf("Register called with phone=%s, username=%s", req.Phone, req.Username)
	if !ValidatePassword(req.Password) {
		return errors.New("密码必须同时包含字母数字标点符号且>=8位")
	}

	if !ValidatePhone(req.Phone) {
		return errors.New("手机号格式错误")
	}

	database := db.GetDB()

	var exist model.SysUser

	if err := database.Where("phone = ?", req.Phone).First(&exist).Error; err == nil {
		return errors.New("手机号已注册")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userID := uuid.New()

	tx := db.DB.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	if err := tx.Exec(
		`INSERT INTO sys_user
	(user_id,username,password_hash,email,phone,status,created_at,updated_at)
	VALUES (?,?,?,?,?,'active',?,?)`,
		userID,
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

	if err := tx.Exec(
		`INSERT INTO user_profile
	(user_info_id,nickname,created_at,updated_at)
	VALUES (?,?,?,?)`,
		userID,
		req.Nickname,
		now,
		now,
	).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
