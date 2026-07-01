// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Role 实体，对应 MySQL 表 x_role（结构对齐 Nest TypeORM）。
type Role struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_role"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("roleName").StorageKey("roleName").Comment("角色名"),
		field.String("roleDesc").StorageKey("roleDesc").Comment("角色描述"),
		field.Int("id").StorageKey("id").Comment("主键 ID"),
	}
}
