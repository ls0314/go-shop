package svc

import (
	"demo-shop/services/user/internal/config"
	"demo-shop/services/user/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Redis

	PermRepo *repository.PermissionRepo
	RoleRepo *repository.RoleRepo
	MenuRepo *repository.MenuRepo
	DeptRepo *repository.DeptRepo
}

func NewServiceContext(config config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(config.DataSource), &gorm.Config{})
	if err != nil {
		panic("连接user_db失败：" + err.Error())
	}

	return &ServiceContext{
		Config: config,
		DB:     db,
		Redis:  redis.MustNewRedis(config.Redis.RedisConf),

		PermRepo: repository.NewPermissionRepo(db),
		RoleRepo: repository.NewRoleRepo(db),
		MenuRepo: repository.NewMenuRepo(db),
		DeptRepo: repository.NewDeptRepo(db),
	}
}
