// 从现网 MySQL information_schema 反向生成 Ent schema（entimport 替代方案）。
// 用法：go run scripts/gen_ent_schema.go
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/jsonkey"
	"github.com/Jiang-Xia/blog-server-go/scripts/nestcomment"
	_ "github.com/go-sql-driver/mysql"
)

type column struct {
	Name      string
	EntType   string
	Optional  bool
	Nillable  bool
	Unique    bool
	Default   string
	DBComment string
}

type tableInfo struct {
	physical      string
	logical       string
	domain        string
	columns       []column
	hasCreateTime bool
	hasUpdateTime bool
	hasIsDelete   bool
	hasVersion    bool
}

var domainTables = map[string]string{
	"user": "user", "role": "user", "privilege": "user", "dept": "user", "menu": "user",
	"role_users_user": "user", "role_data_scope": "user", "role_menus_menu": "user", "role_privileges_privilege": "user",
	"article": "blog", "category": "blog", "tag": "blog", "article_tags_tag": "blog",
	"comment": "blog", "reply": "blog", "like": "blog", "collect": "blog", "msgboard": "blog",
	"link": "blog", "resources": "blog", "my_file": "blog", "file": "blog", "site_notification": "blog",
	"operation_log": "blog", "scheduled_task": "blog", "scheduled_task_log": "blog",
	"sensitive_word": "blog", "sensitive_word_hit": "blog",
	"knowledge_chunk": "blog", "rag_index_job": "blog", "rag_query_log": "blog",
	"pay_order": "rpg",
}

func main() {
	cfg, err := config.MustLoad("configs/monolith.yaml")
	if err != nil {
		panic(err)
	}
	prefix := cfg.MySQL.TablePrefixOrDefault()

	nestRoot := strings.TrimSpace(os.Getenv("NEST_ROOT"))
	if nestRoot == "" {
		nestRoot = filepath.Join("..", "blog-server")
	}
	comments, err := nestcomment.Load(nestRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: load nest entity comments: %v\n", err)
		comments = nestcomment.Index{}
	} else {
		fmt.Printf("loaded field comments from Nest entities under %s\n", nestRoot)
	}

	adminCfg := cfg.MySQL
	adminCfg.Database = ""
	db, err := sql.Open("mysql", adminCfg.FormatDSN())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}

	readSchema := cfg.MySQL.Database
	applyPrefix := false
	var tableCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME LIKE ?`,
		cfg.MySQL.Database, prefix+"%").Scan(&tableCount)
	if tableCount == 0 {
		readSchema = cfg.MySQL.SchemaSourceDatabase
		if readSchema == "" {
			readSchema = "x_my_blog"
		}
		applyPrefix = true
		fmt.Printf("read schema from %s (Nest TypeORM), emit tables with prefix %q for %s\n", readSchema, prefix, cfg.MySQL.Database)
	}

	query := `SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA, COLUMN_COMMENT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?`
	args := []any{readSchema}
	if !applyPrefix {
		query += ` AND TABLE_NAME LIKE ?`
		args = append(args, prefix+"%")
	}
	query += ` ORDER BY TABLE_NAME, ORDINAL_POSITION`

	rows, err := db.Query(query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	tables := map[string]*tableInfo{}
	order := []string{}

	for rows.Next() {
		var tableName, colName, dataType, nullable, colKey, extra, colComment string
		var colDefault sql.NullString
		if err := rows.Scan(&tableName, &colName, &dataType, &nullable, &colKey, &colDefault, &extra, &colComment); err != nil {
			panic(err)
		}
		logical := tableName
		physical := tableName
		if applyPrefix {
			physical = prefix + logical
		} else {
			logical = strings.TrimPrefix(tableName, prefix)
		}
		t, ok := tables[physical]
		if !ok {
			domain := domainTables[logical]
			if domain == "" {
				domain = "rpg"
			}
			t = &tableInfo{physical: physical, logical: logical, domain: domain}
			tables[physical] = t
			order = append(order, physical)
		}
		if colName == "createTime" {
			t.hasCreateTime = true
		}
		if colName == "updateTime" {
			t.hasUpdateTime = true
		}
		if colName == "isDelete" {
			t.hasIsDelete = true
		}
		if colName == "version" {
			t.hasVersion = true
		}
		if isMixinField(colName) {
			continue
		}
		t.columns = append(t.columns, mapColumn(colName, dataType, nullable, colKey, colDefault, extra, colComment))
	}

	base := filepath.Join("services", "monolith", "ent", "schema")
	_ = os.MkdirAll(base, 0o755)
	writeMixin(filepath.Join(base, "mixin.go"))

	for _, physical := range order {
		t := tables[physical]
		out := filepath.Join(base, t.logical+".go")
		content := renderSchema(t.physical, t.logical, toGoTypeName(t.logical), chooseMixin(t), t.columns, comments)
		if err := os.WriteFile(out, []byte(content), 0o644); err != nil {
			panic(err)
		}
		fmt.Println("generated", out)
	}
}

// mixinOverrides 本地库与生产 Nest 表结构不一致时强制指定 mixin（避免误用 TimestampMixin）。
var mixinOverrides = map[string]string{
	"rpg_user_buff":              "CreateTimeMixin", // Nest/生产仅 createTime
	"rpg_user_achievement":       "CreateTimeMixin",
	"rpg_user_lottery_record":    "CreateTimeMixin",
	"rpg_article_tip":            "CreateTimeMixin",
	"rpg_user_social_log":        "CreateTimeMixin",
	"rpg_leaderboard_snapshot":   "CreateTimeMixin",
}

func chooseMixin(t *tableInfo) string {
	if m, ok := mixinOverrides[t.logical]; ok {
		return m
	}
	if !t.hasCreateTime && !t.hasUpdateTime {
		return ""
	}
	if t.hasIsDelete || t.hasVersion {
		return "TimeMixin"
	}
	if t.hasCreateTime && t.hasUpdateTime {
		return "TimestampMixin"
	}
	if t.hasCreateTime {
		return "CreateTimeMixin"
	}
	return ""
}

func writeMixin(path string) {
	content := `// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// TimeMixin 对应 Nest TypeORM 公共字段：createTime / updateTime / isDelete / version。
type TimeMixin struct{ mixin.Schema }

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
		field.Time("updateTime").StorageKey("updateTime").Comment("更新时间").Default(time.Now).UpdateDefault(time.Now),
		field.Bool("isDelete").StorageKey("isDelete").Comment("软删除标记").Default(false),
		field.Int("version").StorageKey("version").Comment("乐观锁版本号").Default(0),
	}
}

// TimestampMixin 仅 createTime/updateTime（RPG 等 Nest 表无 isDelete/version）。
type TimestampMixin struct{ mixin.Schema }

func (TimestampMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
		field.Time("updateTime").StorageKey("updateTime").Comment("更新时间").Default(time.Now).UpdateDefault(time.Now),
	}
}

// CreateTimeMixin 仅 createTime（部分 Nest 表无 updateTime/isDelete）。
type CreateTimeMixin struct{ mixin.Schema }

func (CreateTimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
	}
}
`
	_ = os.WriteFile(path, []byte(content), 0o644)
}

func renderSchema(tableName, logicalTable, goName, mixinName string, cols []column, comments nestcomment.Index) string {
	var b strings.Builder
	b.WriteString("// 由 scripts/gen_ent_schema.go 生成，请勿手改。\n")
	b.WriteString("package schema\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"entgo.io/ent\"\n")
	b.WriteString("\t\"entgo.io/ent/dialect/entsql\"\n")
	b.WriteString("\t\"entgo.io/ent/schema\"\n")
	b.WriteString("\t\"entgo.io/ent/schema/field\"\n")
	b.WriteString(")\n\n")
	b.WriteString(fmt.Sprintf("// %s 实体，对应 MySQL 表 %s（结构对齐 Nest TypeORM）。\n", goName, tableName))
	b.WriteString(fmt.Sprintf("type %s struct{ ent.Schema }\n\n", goName))
	b.WriteString(fmt.Sprintf("// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。\n"))
	b.WriteString(fmt.Sprintf("func (%s) Annotations() []schema.Annotation {\n", goName))
	b.WriteString("\treturn []schema.Annotation{\n")
	b.WriteString(fmt.Sprintf("\t\tentsql.Annotation{Table: %q},\n", tableName))
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	if mixinName != "" {
		b.WriteString(fmt.Sprintf("// Mixin 注入 Nest 公共时间戳字段（%s）。\n", mixinName))
		b.WriteString(fmt.Sprintf("func (%s) Mixin() []ent.Mixin {\n", goName))
		b.WriteString(fmt.Sprintf("\treturn []ent.Mixin{%s{}}\n", mixinName))
		b.WriteString("}\n\n")
	}
	b.WriteString(fmt.Sprintf("// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。\n"))
	b.WriteString(fmt.Sprintf("func (%s) Fields() []ent.Field {\n", goName))
	b.WriteString("\treturn []ent.Field{\n")
	for _, c := range cols {
		fieldComment := resolveFieldComment(comments, logicalTable, c)
		if c.EntType == "JSON" {
			line := fmt.Sprintf("\t\tfield.JSON(%q, map[string]interface{}{}).StorageKey(%q)", c.Name, c.Name)
			line += structTagJSON(c.Name)
			line += commentSuffix(fieldComment)
			if c.Optional {
				line += ".Optional()"
			}
			line += ","
			b.WriteString(line + "\n")
			continue
		}
		line := fmt.Sprintf("\t\tfield.%s(%q).StorageKey(%q)", c.EntType, c.Name, c.Name)
		line += structTagJSON(c.Name)
		line += commentSuffix(fieldComment)
		if c.Unique {
			line += ".Unique()"
		}
		if c.Default != "" {
			line += ".Default(" + c.Default + ")"
		}
		if c.Optional {
			line += ".Optional()"
			if c.Nillable {
				line += ".Nillable()"
			}
		}
		line += ","
		b.WriteString(line + "\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return b.String()
}

func isMixinField(name string) bool {
	switch name {
	case "createTime", "updateTime", "isDelete", "version":
		return true
	default:
		return false
	}
}

func mapColumn(name, dataType, nullable, colKey string, colDefault sql.NullString, extra, dbComment string) column {
	c := column{
		Name:      name,
		Optional:  nullable == "YES",
		Nillable:  nullable == "YES",
		Unique:    colKey == "UNI",
		DBComment: strings.TrimSpace(dbComment),
	}
	switch {
	case strings.HasPrefix(dataType, "int") || dataType == "bigint" || dataType == "tinyint" || dataType == "smallint" || dataType == "mediumint":
		c.EntType = "Int"
	case dataType == "float" || dataType == "double" || strings.HasPrefix(dataType, "decimal"):
		c.EntType = "Float"
	case dataType == "datetime" || dataType == "timestamp" || dataType == "date":
		c.EntType = "Time"
	case dataType == "json":
		c.EntType = "JSON"
		c.Default = "map[string]interface{}{}"
	case dataType == "text" || dataType == "mediumtext" || dataType == "longtext":
		c.EntType = "Text"
	default:
		c.EntType = "String"
	}
	if name == "id" {
		c.Optional = false
		c.Nillable = false
	}
	if colDefault.Valid && !c.Unique {
		def := colDefault.String
		switch {
		case def == "NULL", strings.HasPrefix(def, "CURRENT_TIMESTAMP"):
		case c.EntType == "Int" || c.EntType == "Float":
			c.Default = def
		case c.EntType == "Bool":
			if def == "0" {
				c.Default = "false"
			} else if def == "1" {
				c.Default = "true"
			}
		case c.EntType == "String" || c.EntType == "Text":
			c.Default = fmt.Sprintf("%q", def)
		}
	}
	return c
}

func resolveFieldComment(comments nestcomment.Index, logicalTable string, c column) string {
	if msg := comments.Lookup(logicalTable, c.Name); msg != "" {
		return msg
	}
	return c.DBComment
}

func commentSuffix(comment string) string {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return ""
	}
	return fmt.Sprintf(`.Comment(%q)`, comment)
}

// structTagJSON 为 snake_case 字段生成 Nest 对齐的小驼峰 json tag（保留 0 值，不用 omitempty）。
func structTagJSON(fieldName string) string {
	if !strings.Contains(fieldName, "_") {
		return ""
	}
	return fmt.Sprintf(`.StructTag(`+"`json:\"%s\"`"+`)`, jsonkey.SnakeToCamel(fieldName))
}

func toGoTypeName(table string) string {
	parts := splitWords(table)
	for i, p := range parts {
		parts[i] = capitalize(p)
	}
	return strings.Join(parts, "")
}

func splitWords(s string) []string {
	var parts []string
	var buf strings.Builder
	for i, r := range s {
		if r == '_' {
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
			continue
		}
		if i > 0 && unicode.IsUpper(r) {
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
		}
		buf.WriteRune(r)
	}
	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
