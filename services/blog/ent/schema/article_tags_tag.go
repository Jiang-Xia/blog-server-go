// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ArticleTagsTag 实体，对应 MySQL 表 x_article_tags_tag（结构对齐 Nest TypeORM）。
type ArticleTagsTag struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (ArticleTagsTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_article_tags_tag"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (ArticleTagsTag) Fields() []ent.Field {
	return []ent.Field{
		field.Int("articleId").StorageKey("articleId"),
		field.String("tagId").StorageKey("tagId"),
	}
}
