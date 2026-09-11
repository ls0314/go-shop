package db

import (
	"database/sql"
	"demo-shop-back/src/config"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const defaultMigrationsURL = "file://db//migrations"

func RunMigrationsWith(dbConfig config.DatabaseConfig, sourceURL string) error {

	// 构建数据库连接字符串（不带数据库名，用于创建数据库）
	dsnWithoutDB := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Sslmode,
	)

	// 先尝试创建数据库（如果不存在）
	if err := createDatabaseIfNotExists(dsnWithoutDB, dbConfig.Dbname); err != nil {
		log.Printf("创建数据库失败或数据库已存在: %v", err)
	}

	// 等待数据库创建完成
	time.Sleep(1 * time.Second)

	// 初始化数据库连接
	if err := InitDBWith(dbConfig); err != nil {
		return fmt.Errorf("初始化数据库连接失败: %v", err)
	}

	// 获取现有的数据库连接
	sqlDB, err := GetDB().DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %v", err)
	}

	// 创建 migrate 实例
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("创建 migrate driver 失败: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("创建 migrate 实例失败: %v", err)
	}

	// 执行迁移
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("执行迁移失败: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("数据库已是最新版本，无需迁移")
	} else {
		log.Println("数据库迁移成功")
	}

	return nil
}

// RunMigrations 执行数据库迁移
func RunMigrations() error {
	return RunMigrationsWith(config.GlobalConfig.Database, defaultMigrationsURL)
}

// createDatabaseIfNotExists 创建数据库（如果不存在）
func createDatabaseIfNotExists(dsn string, dbName string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// 检查数据库是否已存在
	var exists int
	row := db.QueryRow(fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", dbName))
	err = row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 如果数据库不存在，则创建
	// 标识符加双引号:库名含连字符/大写等特殊字符时,裸标识符会直接语法错误
	if err == sql.ErrNoRows {
		_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s" WITH ENCODING 'UTF8' LC_COLLATE='en_US.utf8' LC_CTYPE='en_US.utf8'`, dbName))
		if err != nil {
			return err
		}
		log.Printf("数据库 %s 创建成功", dbName)
	} else {
		log.Printf("数据库 %s 已存在", dbName)
	}

	return nil
}

// RollbackMigrations 回滚最后一次迁移
func RollbackMigrations() error {
	// 确保数据库连接已初始化
	if DB == nil {
		if err := InitDB(); err != nil {
			return fmt.Errorf("初始化数据库连接失败: %v", err)
		}
	}

	// 获取现有的数据库连接
	sqlDB, err := GetDB().DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %v", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("创建 migrate driver 失败: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("创建 migrate 实例失败: %v", err)
	}

	// 回滚一步
	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("回滚迁移失败: %v", err)
	}

	log.Println("数据库回滚成功")
	return nil
}
