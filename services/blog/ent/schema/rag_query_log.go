// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RagQueryLog 实体，对应 MySQL 表 x_rag_query_log（结构对齐 Nest TypeORM）。
type RagQueryLog struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RagQueryLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rag_query_log"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RagQueryLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid"),
		field.String("question").StorageKey("question"),
		field.String("answer_preview").StorageKey("answer_preview").Optional().Nillable().StructTag(`json:"answerPreview"`),
		field.JSON("citations_json", []map[string]interface{}{}).StorageKey("citations_json").Optional().StructTag(`json:"citationsJson"`),
		field.Int("latency_ms").StorageKey("latency_ms").Default(0).StructTag(`json:"latencyMs"`),
		field.String("status").StorageKey("status").Default("success"),
		field.Time("create_at").StorageKey("create_at").StructTag(`json:"createAt"`),
	}
}
