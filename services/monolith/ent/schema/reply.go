// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Reply 实体，对应 MySQL 表 x_reply（结构对齐 Nest TypeORM）。
type Reply struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Reply) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_reply"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Reply) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Reply) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").Comment("主键 ID"),
		field.String("parentId").StorageKey("parentId").Comment("评论id(父级id)"),
		field.String("replyUid").StorageKey("replyUid").Comment("回复目标id"),
		field.String("content").StorageKey("content").Comment("回复内容"),
		field.Int("uid").StorageKey("uid").Comment("回复用户id"),
		field.String("status").StorageKey("status").Comment("审核状态").Default("approved"),
	}
}
