package svc

import (
	"demo-shop/services/product/internal/config"
	"demo-shop/services/product/internal/infra/es"
	"demo-shop/services/product/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Redis
	// ES 商品搜索。未配置或连不上时为 nil,调用方须判空降级。
	ES *es.ESClient

	ProductRepo      *repository.ProductRepo
	CategoryRepo     *repository.CategoryRepo
	InventoryLogRepo *repository.InventoryLogRepo
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败：" + err.Error())
	}

	// ES 为可选依赖:Addresses 留空 = 不启用,此时 ES 为 nil,
	// 商品搜索自动降级为直查数据库(getProductSpuListByDB 那条路径)。
	// 一旦配置了地址却连不上,则 fail-fast —— 说明配置有问题,静默降级会掩盖它。
	var esClient *es.ESClient
	if len(c.ES.Addresses) > 0 {
		client, err := es.NewESClient(c.ES.Addresses)
		if err != nil {
			panic("连接 ES 失败：" + err.Error())
		}
		esClient = client
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  redis.MustNewRedis(c.Redis.RedisConf),
		ES:     esClient,

		ProductRepo:      repository.NewProductRepo(db),
		CategoryRepo:     repository.NewCategoryRepo(db),
		InventoryLogRepo: repository.NewInventoryLogRepo(db),
	}
}
