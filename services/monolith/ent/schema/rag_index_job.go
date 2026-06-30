// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RagIndexJob 实体，对应 MySQL 表 x_rag_index_job（结构对齐 Nest TypeORM）。
type RagIndexJob struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RagIndexJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rag_index_job"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RagIndexJob) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("article_id").StorageKey("article_id").Default(0),
		field.String("status").StorageKey("status").Default("pending"),
		field.Int("chunk_count").StorageKey("chunk_count").Default(0),
		field.Text("error_msg").StorageKey("error_msg").Optional().Nillable(),
		field.Time("create_at").StorageKey("create_at"),
		field.Time("update_at").StorageKey("update_at"),
	}
}
