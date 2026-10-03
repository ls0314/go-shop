package svc

import (
	"demo-shop/services/marketing/internal/config"
	"demo-shop/services/marketing/internal/infra/gate"
	"demo-shop/services/marketing/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ServiceContext 依赖注入容器。与 user/product 两服务同构:DB / Redis / 自有 repo。
//
// Redis 只承载券闸门(DS-A-19):
//   - coupon:stock:{templateId}   模板剩余可领量
//   - coupon:limit:{templateId}   该模板的每人限领额度
//   - coupon:ucnt:{tid}:{uid}     某用户在该模板下已领数
//
// 闸门是**加速器不是正确性来源**:Redis 不可用时 CouponGate.Enabled() 为 false,
// 整条闸门旁路,由 DB 的行锁 + 条件更新两道防线独立保证不超发。
type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Redis

	CouponGate     *gate.CouponGate
	CouponRepo     *repository.CouponRepo
	UserCouponRepo *repository.UserCouponRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		panic("连接 marketing_db 失败：" + err.Error())
	}

	rdb := redis.MustNewRedis(c.Redis.RedisConf)

	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  rdb,

		CouponGate:     gate.NewCouponGate(rdb),
		CouponRepo:     repository.NewCouponRepo(db),
		UserCouponRepo: repository.NewUserCouponRepo(db),
	}
}
