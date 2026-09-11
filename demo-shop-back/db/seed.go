package db

import (
	"database/sql"
	"fmt"
	"os"
)

// RunSeedFile 在已初始化的数据库连接上执行种子 SQL 文件。
//
// 种子与迁移分离（工程规范）：迁移链只负责 DDL 与系统数据；
// 演示数据（测试账号、测试角色，见 db/seeds/seed.sql）只服务于
// 开发实例（demo_shop），生产与测试实例不应加载。
//
// 要求：先通过 InitDBWith / RunMigrationsWith 建立连接。
// 文件内多语句由 lib/pq 以简单查询协议一次性执行（与迁移文件同机制）。
func RunSeedFile(seedPath string) error {
	content, err := os.ReadFile(seedPath)
	if err != nil {
		return fmt.Errorf("读取种子文件失败: %w", err)
	}

	var sqlDB *sql.DB
	sqlDB, err = GetDB().DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	if _, err := sqlDB.Exec(string(content)); err != nil {
		return fmt.Errorf("执行种子数据失败: %w", err)
	}
	return nil
}
