// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// File 实体，对应 MySQL 表 x_file（结构对齐 Nest TypeORM）。
type File struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (File) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_file"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").Comment("主键 ID"),
		field.String("pid").StorageKey("pid").Default("0"),
		field.Int("isFolder").StorageKey("isFolder").Default(0),
		field.String("originalname").StorageKey("originalname"),
		field.String("filename").StorageKey("filename"),
		field.String("type").StorageKey("type"),
		field.Int("size").StorageKey("size"),
		field.String("url").StorageKey("url"),
		field.Time("create_at").StorageKey("create_at").Comment("创建时间"),
	}
}
