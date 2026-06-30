// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Comment 实体，对应 MySQL 表 x_comment（结构对齐 Nest TypeORM）。
type Comment struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Comment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_comment"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Comment) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").Comment("主键 ID"),
		field.String("content").StorageKey("content").Comment("评论内容"),
		field.Int("uid").StorageKey("uid"),
		field.Int("userId").StorageKey("userId").Optional().Nillable(),
		field.Int("articleId").StorageKey("articleId").Optional().Nillable(),
		field.String("status").StorageKey("status").Comment("审核状态").Default("approved"),
	}
}
