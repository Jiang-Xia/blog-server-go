// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// KnowledgeChunk 实体，对应 MySQL 表 x_knowledge_chunk（结构对齐 Nest TypeORM）。
type KnowledgeChunk struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (KnowledgeChunk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_knowledge_chunk"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (KnowledgeChunk) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("article_id").StorageKey("article_id"),
		field.Int("chunk_index").StorageKey("chunk_index"),
		field.String("title").StorageKey("title"),
		field.Text("content").StorageKey("content"),
		field.String("url").StorageKey("url"),
		field.String("category").StorageKey("category").Optional().Nillable(),
		field.JSON("tags", []string{}).StorageKey("tags").Optional(),
		field.JSON("embedding_json", []float64{}).StorageKey("embedding_json"),
		field.String("status").StorageKey("status").Comment("active | deleted，下架/作者禁用为 deleted 不参与检索").Default("active"),
		field.Time("indexed_at").StorageKey("indexed_at"),
		field.Time("create_at").StorageKey("create_at"),
		field.Time("update_at").StorageKey("update_at"),
		field.String("source_type").StorageKey("source_type").Default("article"),
		field.String("source_key").StorageKey("source_key").Default(""),
		field.String("heading_path").StorageKey("heading_path").Optional().Nillable(),
		field.String("content_type").StorageKey("content_type").Default("prose"),
		field.Text("search_text").StorageKey("search_text").Optional().Nillable(),
	}
}
