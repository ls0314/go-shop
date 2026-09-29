package main

import (
	"database/sql"
	"flag"
	"log"

	"demo-shop/services/user/internal/config"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zeromicro/go-zero/core/conf"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	configFile    = flag.String("f", "etc/user.yaml", "the config file")
	migrationsURL = flag.String("migrations", "file://migrations", "migrations 目录")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	db, err := sql.Open("pgx", c.DataSource)
	if err != nil {
		log.Fatalf("打开 user_db 失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("user_db 不可达: %v", err)
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

	log.Println("user_db 迁移完成")
}
