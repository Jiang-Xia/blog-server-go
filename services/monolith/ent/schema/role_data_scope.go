// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RoleDataScope 实体，对应 MySQL 表 x_role_data_scope（结构对齐 Nest TypeORM）。
type RoleDataScope struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RoleDataScope) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_role_data_scope"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RoleDataScope) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RoleDataScope) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("roleId").StorageKey("roleId").Comment("角色ID"),
		field.String("resourceType").StorageKey("resourceType").Comment("资源类型").Default("article"),
		field.String("scopeType").StorageKey("scopeType").Comment("数据范围类型"),
		field.JSON("deptIds", map[string]interface{}{}).StorageKey("deptIds").Comment("CUSTOM 时指定的部门 ID 列表").Optional(),
	}
}
