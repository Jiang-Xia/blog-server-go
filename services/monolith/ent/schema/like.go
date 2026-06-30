// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Like 实体，对应 MySQL 表 x_like（结构对齐 Nest TypeORM）。
type Like struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Like) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_like"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Like) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").Comment("主键 ID"),
		field.Int("articleId").StorageKey("articleId"),
		field.Int("uid").StorageKey("uid").Default(-999),
		field.String("ip").StorageKey("ip").Default(""),
		field.String("status").StorageKey("status"),
	}
}
