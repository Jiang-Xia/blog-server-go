// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Link 实体，对应 MySQL 表 x_link（结构对齐 Nest TypeORM）。
type Link struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Link) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_link"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Link) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Link) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("自身id"),
		field.String("icon").StorageKey("icon").Comment("图标链接"),
		field.String("url").StorageKey("url").Comment("网址"),
		field.String("title").StorageKey("title").Comment("标题"),
		field.String("desp").StorageKey("desp").Comment("个人签名"),
		field.Int("agreed").StorageKey("agreed").Comment("是否已经同意申请").Default(0),
		field.String("lastCheckStatus").StorageKey("lastCheckStatus").Comment("友链健康状态").Default("unchecked"),
		field.Time("lastCheckTime").StorageKey("lastCheckTime").Comment("最后检测时间").Optional().Nillable(),
	}
}
