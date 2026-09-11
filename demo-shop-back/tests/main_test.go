package tests

import (
	"demo-shop-back/db"
	"demo-shop-back/src/config"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	cfg := config.DatabaseConfig{
		Host:     getenv("TEST_PG_HOST", "localhost"),
		Port:     getenv("TEST_PG_PORT", "5432"),
		User:     getenv("TEST_PG_USER", "postgres"),
		Password: getenv("TEST_PG_PASSWORD", "postgres"),
		Dbname:   "demo_shop_test",
		Sslmode:  "disable",
	}

	migrationsURL, err := filepath.Abs("../db/migrations")
	if err != nil {
		log.Fatalf("定位迁移目录失败：%v", err)
	}

	sourceURL := "file://" + filepath.ToSlash(migrationsURL)

	if err := db.RunMigrationsWith(cfg, sourceURL); err != nil {
		log.Fatalf("测试库引导失败：%v", err)
	}

	if os.Getenv("TEST_KEEP_DATA") == "" {
		if err := truncateAllTables(); err != nil {
			log.Fatalf("清空测试数据失败: %v", err)
		}
		log.Println("[test] 测试库已清空,本轮从零态开始")
	} else {
		log.Println("[test] TEST_KEEP_DATA=1,保留上轮数据(用于事后取证)")
	}

	os.Exit(m.Run())
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func truncateAllTables() error {
	var tables []string
	if err := db.DB.Raw(
		`SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename <> 'schema_migrations'`,
	).Scan(&tables).Error; err != nil {
		return fmt.Errorf("查询表清单失败: %w", err)
	}
	if len(tables) == 0 {
		return nil
	}
	quoted := make([]string, 0, len(tables))
	for _, tb := range tables {
		quoted = append(quoted, `"`+tb+`"`) // 标识符加引号,防未来表名带大小写/特殊字符
	}
	stmt := "TRUNCATE TABLE " + strings.Join(quoted, ", ") + " RESTART IDENTITY CASCADE"
	if err := db.DB.Exec(stmt).Error; err != nil {
		return fmt.Errorf("清空业务表失败: %w", err)
	}
	return nil
}
