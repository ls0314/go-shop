package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"demo-shop/services/product/internal/config"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zeromicro/go-zero/core/conf"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	configFile    = flag.String("f", "etc/product.yaml", "the config file")
	migrationsURL = flag.String("migrations", "file://migrations", "migrations 目录")
	verifyOnly    = flag.Bool("verify", false, "只校验 schema 版本与表是否就绪,不执行迁移")
	counts        = flag.Bool("counts", false, "打印本服务各表的行数(割接对账用),不执行迁移")
)

// expectedTables 本服务的表清单。
//
// 为什么要显式列出而不是"查 information_schema 数表个数":
// 迁移跑成功但漏建一张表(或建错库),数个数是查不出来的。
// 清单与 migrations/*.up.sql 一一对应,新增迁移时必须同步更新。
var expectedTables = []string{
	"sys_category",          // 000001
	"sys_product_spu",       // 000002
	"sys_product_sku",       // 000002
	"sys_product_spu_image", // 000002
	"sys_product_stock_log", // 000002
}

// requiredIndexes 幂等闸与唯一约束。库存四操作的正确性依赖它们,
// 缺了不会报错、只会静默重复扣减,故纳入就绪校验。
var requiredIndexes = []string{
	"uk_stock_log_order", // (order_id, sku_id, change_type) 三元组唯一索引
	"idx_sku_code",       // sku_code 部分唯一索引
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	db, err := sql.Open("pgx", c.DataSource)
	if err != nil {
		log.Fatalf("打开 product_db 失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("product_db 不可达: %v", err)
	}

	if *verifyOnly {
		if err := verifySchema(db); err != nil {
			log.Fatalf("schema 校验失败: %v", err)
		}
		return
	}

	if *counts {
		if err := printCounts(db); err != nil {
			log.Fatalf("统计行数失败: %v", err)
		}
		return
	}

	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		log.Fatalf("创建 migrate driver 失败: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(*migrationsURL, "postgres", driver)
	if err != nil {
		log.Fatalf("创建 migrate 实例失败: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("执行迁移失败: %v", err)
	}

	log.Println("product_db 迁移完成")

	// 迁移后立即校验:迁完不等于迁对(漏表/漏索引都可能在 Up 返回 nil 的情况下发生)
	if err := verifySchema(db); err != nil {
		log.Fatalf("迁移后校验失败: %v", err)
	}
	log.Println("product_db schema 校验通过")
}

// printCounts 打印本服务各表的行数。
//
// 这是**割接对账**的工具:数据割接(源库 → 本库)后要能一眼比对行数,
// 否则"搬完了"只是感觉。对源库与目标库各跑一次即可比对:
//
//	go run ./cmd/migrate -f etc/product.yaml -counts   # 目标库 product_db
//	go run ./cmd/migrate -f etc/<源库配置>.yaml -counts  # 源库 demo_shop
func printCounts(db *sql.DB) error {
	log.Printf("[counts] 本服务表清单:")
	for _, tb := range expectedTables {
		var n int64
		if err := db.QueryRow("SELECT COUNT(*) FROM " + tb).Scan(&n); err != nil {
			return fmt.Errorf("统计 %s 失败: %w", tb, err)
		}
		log.Printf("[counts]   %-24s %d 行", tb, n)
	}
	return nil
}

// verifySchema 校验 schema 版本与表/索引是否就绪。
func verifySchema(db *sql.DB) error {
	// golang-migrate 的版本账本:表不存在说明一次都没迁过
	var version int64
	var dirty bool
	err := db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty)
	switch {
	case err == sql.ErrNoRows:
		return fmt.Errorf("schema_migrations 为空:从未执行过迁移")
	case err != nil:
		return fmt.Errorf("读取 schema 版本失败(表可能不存在): %w", err)
	}
	if dirty {
		return fmt.Errorf("schema_migrations 处于 dirty 状态(version=%d):上一次迁移中途失败,需人工修复", version)
	}

	missingTables, err := missingRelations(db, "BASE TABLE", expectedTables)
	if err != nil {
		return err
	}
	missingIndexes, err := missingRelations(db, "INDEX", requiredIndexes)
	if err != nil {
		return err
	}

	log.Printf("[verify] schema 版本=%d", version)
	log.Printf("[verify] 表 %d/%d 就绪", len(expectedTables)-len(missingTables), len(expectedTables))
	log.Printf("[verify] 关键索引 %d/%d 就绪", len(requiredIndexes)-len(missingIndexes), len(requiredIndexes))

	if len(missingTables) > 0 || len(missingIndexes) > 0 {
		var parts []string
		if len(missingTables) > 0 {
			sort.Strings(missingTables)
			parts = append(parts, "缺表: "+strings.Join(missingTables, ", "))
		}
		if len(missingIndexes) > 0 {
			sort.Strings(missingIndexes)
			parts = append(parts, "缺索引: "+strings.Join(missingIndexes, ", "))
		}
		return fmt.Errorf("%s", strings.Join(parts, "; "))
	}
	return nil
}

// missingRelations 返回在 sys 表/索引清单中缺席的名字(同时兼容 index 与 constraint 两种落库形态)
func missingRelations(db *sql.DB, relKind string, expected []string) ([]string, error) {
	rows, err := db.Query(
		`SELECT c.relname
		   FROM pg_class c
		   JOIN pg_namespace n ON n.oid = c.relnamespace
		  WHERE n.nspname = current_schema() AND c.relkind = ANY($1)`,
		pgRelKinds(relKind),
	)
	if err != nil {
		return nil, fmt.Errorf("查询 %s 清单失败: %w", relKind, err)
	}
	defer rows.Close()

	existing := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		existing[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// UNIQUE INDEX 在 pg_class 里是 'i',但用 ALTER TABLE ADD CONSTRAINT UNIQUE 建的
	// 会同时有 'i'(索引)与 'c'(约束)。两种形态都算"存在",避免误报。
	missing := make([]string, 0)
	for _, name := range expected {
		if _, ok := existing[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing, nil
}

// pgRelKinds 把可读的名字翻成 pg_class.relkind 取值
func pgRelKinds(kind string) []string {
	switch kind {
	case "INDEX":
		return []string{"i", "I"}
	default: // BASE TABLE
		return []string{"r", "p"}
	}
}
