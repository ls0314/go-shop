package svc

import (
	"demo-shop/services/user/internal/config"
	"demo-shop/services/user/internal/repository"
	"demo-shop/services/user/internal/utils"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Redis
	JWT    *utils.JWTIssuer

	PermRepo  *repository.PermissionRepo
	RoleRepo  *repository.RoleRepo
	MenuRepo  *repository.MenuRepo
	DeptRepo  *repository.DeptRepo
	ScopeRepo *repository.ScopeRepo

	RolePermRepo *repository.RolePermRepo
	RoleMenuRepo *repository.RoleMenuRepo
	MenuPermRepo *repository.MenuPermRepo
	UserRoleRepo *repository.UserRoleRepo
	UserDeptRepo *repository.UserDeptRepo

	UserRepo        *repository.UserRepo
	UserProfileRepo *repository.UserProfileRepo

	// AddressRepo 收货地址。表在 C1 就按归属建到了 user_db
	// (migrations/000006),但读写路径一直留在单体 —— 本次搬过来
	// (DS-A-25 §4.5.2 第 1 条:address → user-service)。
	AddressRepo *repository.AddressRepo
}

func NewServiceContext(config config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(config.DataSource), &gorm.Config{})
	if err != nil {
		panic("连接user_db失败：" + err.Error())
	}

	issuer, err := utils.NewJWTIssuer(config.JwtPrivateKeyPath)
	if err != nil {
		panic("加载 JWT 私钥失败: " + err.Error())
	}

	return &ServiceContext{
		Config: config,
		DB:     db,
		Redis:  redis.MustNewRedis(config.Redis.RedisConf),
		JWT:    issuer,

		PermRepo:  repository.NewPermissionRepo(db),
		RoleRepo:  repository.NewRoleRepo(db),
		MenuRepo:  repository.NewMenuRepo(db),
		DeptRepo:  repository.NewDeptRepo(db),
		ScopeRepo: repository.NewScopeRepo(db),

		RolePermRepo: repository.NewRolePermRepo(db),
		RoleMenuRepo: repository.NewRoleMenuRepo(db),
		MenuPermRepo: repository.NewMenuPermRepo(db),
		UserRoleRepo: repository.NewUserRoleRepo(db),
		UserDeptRepo: repository.NewUserDeptRepo(db),

		UserRepo:        repository.NewUserRepo(db),
		UserProfileRepo: repository.NewUserProfileRepo(db),

		AddressRepo: repository.NewAddressRepo(db),
	}
}
