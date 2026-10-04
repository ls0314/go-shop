package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"demo-shop/services/trade/internal/config"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zeromicro/go-zero/core/conf"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	configFile    = flag.String("f", "etc/trade.yaml", "the config file")
	migrationsURL = flag.String("migrations", "file://migrations", "migrations 目录")
	verifyOnly    = flag.Bool("verify", false, "只校验 schema 版本与表是否就绪,不执行迁移")
	counts        = flag.Bool("counts", false, "打印本服务各表的行数(割接对账用),不执行迁移")
)

// expectedTables 本服务的表清单。
//
// 显式列出而不是"查 information_schema 数表个数":迁移跑成功但漏建一张表
// (或建错库)是数个数查不出来的。清单与 migrations/*.up.sql 一一对应。
var expectedTables = []string{
	"user_cart_item",      // 000001
	"user_order_master",   // 000002
	"user_order_detail",   // 000002
	"user_order_log",      // 000002
	"user_payment_record", // 000003
	"sys_outbox_message",  // 000004
}

// requiredIndexes 订单域正确性依赖的索引。
//
// 这几条缺了**不会报错**,只会静默失去对应的保证,故纳入就绪校验:
//   - uk_idempotent_key:重复下单靠它兜底(服务端先查后插,并发下查不到也不能保证唯一)
//   - uk_pay_trade_no:支付回调幂等靠它(同一渠道交易号只能落一次)
//   - idx_order_expire:超时扫描的性能依据(没有它每轮全表扫)
var requiredIndexes = []string{
	"uk_order_no",
	"uk_idempotent_key",
	"uk_pay_no",
	"uk_pay_trade_no",
	"idx_order_expire",
	"uk_user_sku",
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := ensureDatabase(c.DataSource); err != nil {
		log.Fatalf("准备 trade_db 失败: %v", err)
	}

	db, err := sql.Open("pgx", c.DataSource)
	if err != nil {
		log.Fatalf("打开 trade_db 失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("trade_db 不可达: %v", err)
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

	log.Println("trade_db 迁移完成")

	// 迁移后立即校验:迁完不等于迁对(漏表/漏索引都可能在 Up 返回 nil 的情况下发生)
	if err := verifySchema(db); err != nil {
		log.Fatalf("迁移后校验失败: %v", err)
	}
	log.Println("trade_db schema 校验通过")
}

// ensureDatabase 确保目标库存在,不存在则创建。
//
// 与 marketing-service 的 cmd/migrate 同一实现:迁到不存在的库时,
// 原生报错是 "database ... does not exist",既看不出该建什么库,
// 也不知道用什么参数建。
//
// 建库语句不能在目标库里执行(会报"正在被访问"),故先连 maintenance 库 postgres。
// PostgreSQL 的 CREATE DATABASE 不支持 IF NOT EXISTS,故先查 pg_database 判存在性。
func ensureDatabase(dataSource string) error {
	dbName, adminDSN, err := splitDatabase(dataSource)
	if err != nil {
		return err
	}

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("连接 maintenance 库失败: %w", err)
	}
	defer admin.Close()

	if err := admin.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL 不可达(检查 DataSource 的地址/账号): %w", err)
	}

	var exists bool
	if err := admin.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbName).Scan(&exists); err != nil {
		return fmt.Errorf("查询 pg_database 失败: %w", err)
	}
	if exists {
		log.Printf("[bootstrap] 库 %s 已存在", dbName)
		return nil
	}

	// 库名来自配置而非用户输入,且下面已加双引号;CREATE DATABASE 不支持参数占位
	if _, err := admin.Exec(`CREATE DATABASE "` + dbName + `"`); err != nil {
		return fmt.Errorf("创建库 %s 失败: %w", dbName, err)
	}
	log.Printf("[bootstrap] 已创建库 %s", dbName)
	return nil
}

// splitDatabase 从 DSN 里取出库名,并把 DSN 改成连 maintenance 库 postgres
func splitDatabase(dataSource string) (dbName, adminDSN string, err error) {
	u, err := url.Parse(dataSource)
	if err != nil {
		return "", "", fmt.Errorf("DataSource 不是合法 URL: %w", err)
	}
	dbName = strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return "", "", fmt.Errorf("DataSource 里没有库名: %s", dataSource)
	}
	u.Path = "/postgres"
	return dbName, u.String(), nil
}

// printCounts 打印本服务各表的行数(割接对账用)。
// 对源库与目标库各跑一次即可比对:
//
//	go run ./cmd/migrate -f etc/trade.yaml -counts     # 目标库 trade_db
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

// missingRelations 返回在清单中缺席的表/索引名
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
