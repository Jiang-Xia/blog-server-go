// Ent migrate：在 CI/测试库 x_my_blog 创建 user/blog/rpg 全表结构。
//
//	go run scripts/ci/migrate_schemas.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	blogmigrate "github.com/Jiang-Xia/blog-server-go/services/blog/ent/migrate"
	_ "github.com/Jiang-Xia/blog-server-go/services/blog/ent/runtime"
	rpgmigrate "github.com/Jiang-Xia/blog-server-go/services/rpg/ent/migrate"
	_ "github.com/Jiang-Xia/blog-server-go/services/rpg/ent/runtime"
	usermigrate "github.com/Jiang-Xia/blog-server-go/services/user/ent/migrate"
	_ "github.com/Jiang-Xia/blog-server-go/services/user/ent/runtime"
)

func main() {
	ctx := context.Background()
	if err := ensureDatabase(ctx); err != nil {
		fail(err)
	}
	dsn := mysqlDSN()
	for name, fn := range map[string]func(context.Context, *entsql.Driver) error{
		"user": func(ctx context.Context, drv *entsql.Driver) error {
			return usermigrate.NewSchema(drv).Create(ctx)
		},
		"blog": func(ctx context.Context, drv *entsql.Driver) error {
			return blogmigrate.NewSchema(drv).Create(ctx)
		},
		"rpg": func(ctx context.Context, drv *entsql.Driver) error {
			return rpgmigrate.NewSchema(drv).Create(ctx)
		},
	} {
		if err := runMigrate(ctx, dsn, name, fn); err != nil {
			fail(err)
		}
		fmt.Println("migrated", name)
	}
}

func ensureDatabase(ctx context.Context) error {
	db, err := sql.Open("mysql", mysqlDSNNoDB())
	if err != nil {
		return err
	}
	defer db.Close()
	name := env("CI_MYSQL_DATABASE", defaultMySQLDB)
	_, err = db.ExecContext(ctx, fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		name,
	))
	return err
}

func runMigrate(ctx context.Context, dsn, label string, create func(context.Context, *entsql.Driver) error) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("%s open: %w", label, err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("%s ping: %w", label, err)
	}
	drv := entsql.OpenDB(dialect.MySQL, db)
	return create(ctx, drv)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "migrate_schemas: %v\n", err)
	os.Exit(1)
}
