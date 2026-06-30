// 由 scripts/gen_ent_schema.go 生成；sensitive_word_hit 仅含 createTime。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// SensitiveWordHit 实体，对应 MySQL 表 x_sensitive_word_hit（结构对齐 Nest TypeORM）。
type SensitiveWordHit struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (SensitiveWordHit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_sensitive_word_hit"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (SensitiveWordHit) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("命中记录ID"),
		field.String("sourceType").StorageKey("sourceType").Comment("来源类型：comment/msgboard"),
		field.String("sourceId").StorageKey("sourceId").Comment("来源ID（评论uuid/留言id）"),
		field.Text("content").StorageKey("content").Comment("原始内容"),
		field.String("hitWords").StorageKey("hitWords").Comment("命中的敏感词（逗号分隔）"),
		field.Int("uid").StorageKey("uid").Comment("用户ID（留言时可为null）").Optional().Nillable(),
		field.String("ip").StorageKey("ip").Comment("IP地址").Optional().Nillable(),
		field.String("status").StorageKey("status").Comment("审核状态").Default("pending"),
		field.Int("reviewerId").StorageKey("reviewerId").Comment("审核人ID").Optional().Nillable(),
		field.Time("reviewTime").StorageKey("reviewTime").Comment("审核时间").Optional().Nillable(),
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
	}
}
