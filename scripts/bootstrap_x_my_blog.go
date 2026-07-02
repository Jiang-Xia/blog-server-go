package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/nestenv"
	_ "github.com/go-sql-driver/mysql"
)

func loadMySQLConfig() (config.MySQLConfig, string, error) {
	envPath := flag.String("env", "", "deploy/pm2/env.production 路径（优先于 CONFIG_PATH/yaml）")
	sourceDB := flag.String("source", "", "源库名，默认 myblog 或 BOOTSTRAP_SOURCE_DB")
	flag.Parse()

	src := strings.TrimSpace(*sourceDB)
	if src == "" {
		src = strings.TrimSpace(os.Getenv("BOOTSTRAP_SOURCE_DB"))
	}
	if src == "" {
		src = "myblog"
	}

	if strings.TrimSpace(*envPath) != "" {
		cfg, err := nestenv.ConfigFromFile(*envPath)
		if err != nil {
			return config.MySQLConfig{}, "", err
		}
		return cfg.MySQL, src, nil
	}

	path := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	if path == "" {
		path = "configs/monolith.yaml"
	}
	cfg, err := config.MustLoad(path)
	if err != nil {
		return config.MySQLConfig{}, "", err
	}
	if src == "myblog" && strings.TrimSpace(cfg.MySQL.SchemaSourceDatabase) != "" {
		src = cfg.MySQL.SchemaSourceDatabase
	}
	return cfg.MySQL, src, nil
}

func main() {
	mysqlCfg, sourceDB, err := loadMySQLConfig()
	if err != nil {
		panic(err)
	}
	prefix := mysqlCfg.TablePrefixOrDefault()
	targetDB := mysqlCfg.Database
	if targetDB == "" {
		panic("target database empty in config")
	}
	if sourceDB == targetDB {
		panic(fmt.Sprintf("source and target database must differ (source=%s target=%s)", sourceDB, targetDB))
	}

	dsn := mysqlCfg.FormatDSN()
	base := mysqlCfg
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
		fmt.Fprintf(os.Stderr, "run deploy/sql/prod/001_create_x_my_blog.sql as MySQL root, then retry\n")
	}

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
		if strings.HasPrefix(name, prefix) {
			continue
		}
		tables = append(tables, name)
	}
	if len(tables) == 0 {
		panic(fmt.Sprintf("source database %s has no tables; ensure Nest myblog is initialized", sourceDB))
	}

	for _, t := range tables {
		dest := prefix + t
		q := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s` LIKE `%s`.`%s`", targetDB, dest, sourceDB, t)
		if _, err := db.Exec(q); err != nil {
			panic(fmt.Errorf("clone %s -> %s: %w", t, dest, err))
		}
		fmt.Printf("cloned %s.%s -> %s.%s\n", sourceDB, t, targetDB, dest)
	}

	check, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer check.Close()
	if err := check.Ping(); err != nil {
		panic(fmt.Errorf("ping target %s: %w", targetDB, err))
	}
	fmt.Printf("bootstrap done: %d tables in %s with prefix %q (source %s unchanged)\n", len(tables), targetDB, prefix, sourceDB)
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
	return fmt.Errorf("database %s not accessible; run deploy/sql/prod/001_create_x_my_blog.sql as MySQL admin", targetDB)
}
