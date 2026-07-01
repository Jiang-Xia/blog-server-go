// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Category 实体，对应 MySQL 表 x_category（结构对齐 Nest TypeORM）。
type Category struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Category) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_category"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid"),
		field.String("label").StorageKey("label"),
		field.String("value").StorageKey("value").Comment("testNums: string[];"),
		field.String("color").StorageKey("color"),
		field.Time("create_at").StorageKey("create_at").Comment("创建时间"),
		field.Time("update_at").StorageKey("update_at").Comment("更新时间"),
	}
}
