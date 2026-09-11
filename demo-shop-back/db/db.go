package db

import (
	"demo-shop-back/src/config"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接实例
var DB *gorm.DB

func InitDBWith(dbConfig config.DatabaseConfig) error {

	// 获取数据库连接字符串
	dsn := dbConfig.GetDSN()

	// 配置 GORM 日志
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到错误
			Colorful:                  true,        // 彩色打印
		},
	)

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: newLogger,
	}

	// 打开数据库连接
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %v", err)
	}

	// 获取底层的 sql.DB 实例来设置连接池参数
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(80)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(80)           // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	log.Println("数据库连接成功")
	return nil
}

// InitDB 初始化数据库连接
func InitDB() error {
	return InitDBWith(config.GlobalConfig.Database)
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("获取底层数据库连接失败: %v", err)
		}
		return sqlDB.Close()
	}
	return nil
}
