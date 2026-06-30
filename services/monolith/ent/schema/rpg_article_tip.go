// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgArticleTip 实体，对应 MySQL 表 x_rpg_article_tip（结构对齐 Nest TypeORM）。
type RpgArticleTip struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgArticleTip) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_article_tip"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgArticleTip) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgArticleTip) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("打赏者UID"),
		field.Int("articleId").StorageKey("articleId").Comment("文章ID"),
		field.Int("authorUid").StorageKey("authorUid").Comment("作者UID"),
		field.Int("amount").StorageKey("amount").Comment("打赏碎片数"),
	}
}
