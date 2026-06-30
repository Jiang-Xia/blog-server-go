// 初始化本地库 x_my_blog：从 Nest TypeORM 维护的源库克隆表结构，表名加 x_ 前缀（仅结构，不拷数据）。
// 用法：go run scripts/bootstrap_x_my_blog.go
// 环境变量 BOOTSTRAP_SOURCE_DB 可指定源库，默认 myblog。
package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg, err := config.MustLoad("configs/monolith.yaml")
	if err != nil {
		panic(err)
	}
	prefix := cfg.MySQL.TablePrefixOrDefault()
	targetDB := cfg.MySQL.Database
	sourceDB := strings.TrimSpace(os.Getenv("BOOTSTRAP_SOURCE_DB"))
	if sourceDB == "" {
		sourceDB = "myblog"
	}

	dsn := cfg.MySQL.FormatDSN()
	// 连接时不指定库，便于 CREATE DATABASE / 跨库 DDL。
	base := cfg.MySQL
	base.Database = ""
	db, err := sql.Open("mysql", base.FormatDSN())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("mysql ping: %w", err))
	}

	if _, err := db.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		targetDB,
	)); err != nil {
		fmt.Fprintf(os.Stderr, "warn: cannot CREATE DATABASE (need admin?): %v\n", err)
		fmt.Fprintf(os.Stderr, "run deploy/sql/local/001_create_x_my_blog.sql as MySQL root, then retry\n")
	}
	fmt.Printf("database %s ready\n", targetDB)

	if err := ensureTargetDB(db, dsn, targetDB); err != nil {
		panic(err)
	}

	rows, err := db.Query("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME", sourceDB)
	if err != nil {
		panic(fmt.Errorf("list tables from %s: %w", sourceDB, err))
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		tables = append(tables, name)
	}
	if len(tables) == 0 {
		panic(fmt.Sprintf("source database %s has no tables; run NestJS TypeORM synchronize on it first", sourceDB))
	}

	for _, t := range tables {
		dest := prefix + t
		q := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s` LIKE `%s`.`%s`", targetDB, dest, sourceDB, t)
		if _, err := db.Exec(q); err != nil {
			panic(fmt.Errorf("clone %s -> %s: %w", t, dest, err))
		}
		fmt.Printf("cloned %s.%s -> %s.%s\n", sourceDB, t, targetDB, dest)
	}

	// 验证目标库可连（与 monolith 配置一致）
	check, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer check.Close()
	if err := check.Ping(); err != nil {
		panic(fmt.Errorf("ping target %s: %w", targetDB, err))
	}
	fmt.Printf("bootstrap done: %d tables in %s with prefix %q\n", len(tables), targetDB, prefix)
}

func ensureTargetDB(admin *sql.DB, dsn, targetDB string) error {
	check, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer check.Close()
	if err := check.Ping(); err == nil {
		return nil
	}
	return fmt.Errorf("database %s not accessible; run deploy/sql/local/001_create_x_my_blog.sql as MySQL admin", targetDB)
}
