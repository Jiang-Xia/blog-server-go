// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// SiteNotification 实体，对应 MySQL 表 x_site_notification（结构对齐 Nest TypeORM）。
type SiteNotification struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (SiteNotification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_site_notification"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (SiteNotification) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (SiteNotification) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("接收通知的用户 uid（文章作者等）"),
		field.String("type").StorageKey("type").Comment("通知类型，如 comment_on_article"),
		field.Text("payload").StorageKey("payload").Comment("JSON 字符串，含 articleId / articleTitle / commentId 等展示字段"),
		field.Int("read").StorageKey("read").Default(0),
	}
}
