// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Collect 实体，对应 MySQL 表 x_collect（结构对齐 Nest TypeORM）。
type Collect struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Collect) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_collect"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Collect) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Collect) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("收藏用户id"),
		field.Int("articleId").StorageKey("articleId").Comment("收藏文章id"),
	}
}
