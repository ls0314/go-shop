package repository

import (
	"demo-shop/services/user/internal/model"

	"gorm.io/gorm"
)

// UserRepo 用户表数据层实例。
// 当前只承载绑定域需要的主键查询,用户 CRUD 随用户域迁入时再补。
type UserRepo struct {
	DB *gorm.DB
}

// NewUserRepo 创建用户表数据层实例
func NewUserRepo(conn *gorm.DB) *UserRepo {
	return &UserRepo{DB: conn}
}

// WithTx 切换数据库事务实例
func (ur *UserRepo) WithTx(tx *gorm.DB) *UserRepo {
	return &UserRepo{DB: tx}
}

// GetUserById 根据ID查询用户
func (ur *UserRepo) GetUserById(id int64) (*model.SysUser, error) {
	var user model.SysUser
	if err := ur.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
