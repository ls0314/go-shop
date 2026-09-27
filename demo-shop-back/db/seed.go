package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq" // 种子执行走 lib/pq 简单查询协议,单次 Exec 可携带多语句(与迁移文件同机制)
)

// RunSeeds 顺序执行 seedDir 下未应用过的种子文件 —— 版本账本模式(与迁移的 schema_migrations 对称)。
//
// 机制:
//   - 文件名即版本号(如 0001_demo.sql),按文件名升序执行;新增种子 = 新增文件,旧文件永不改动
//   - 每个文件一个事务:「执行 + 写 sys_seed_history 账本」原子提交 —— 中途崩溃自动回滚,
//     下次启动自动重试该文件(对比裸 Exec:半执行状态无法检测)
//   - 已记账的文件自动跳过,因此种子内容无需每条语句自备幂等保护(留 ON CONFLICT 作双保险亦可)
//   - version 是主键,多实例并发启动时后到者冲突跳过即可
//
// 约束:种子文件内部不得包含 BEGIN/COMMIT —— 事务由本函数统一管理。
//
// 执行使用独立的 lib/pq 连接(传入 DSN),不经过全局 gorm 连接:
// 多语句文件依赖 lib/pq 的简单查询协议,gorm 默认的 pgx 扩展协议不支持单次 Exec 多语句。
func RunSeeds(seedDir, dsn string) error {
	entries, err := os.ReadDir(seedDir)
	if err != nil {
		return fmt.Errorf("读取种子目录失败: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil
	}

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("打开种子连接失败: %w", err)
	}
	defer sqlDB.Close()

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")

		var applied bool
		err := sqlDB.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM sys_seed_history WHERE version = $1)`, version,
		).Scan(&applied)
		if err != nil {
			return fmt.Errorf("查询种子账本失败(确认 000014 迁移已应用): %w", err)
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(filepath.Join(seedDir, name))
		if err != nil {
			return fmt.Errorf("读取种子文件 %s 失败: %w", name, err)
		}

		tx, err := sqlDB.Begin()
		if err != nil {
			return fmt.Errorf("开启种子事务失败: %w", err)
		}
		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("执行种子 %s 失败: %w", name, err)
		}
		if _, err := tx.Exec(`INSERT INTO sys_seed_history (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("记录种子版本 %s 失败: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交种子 %s 失败: %w", name, err)
		}
	}
	return nil
}
