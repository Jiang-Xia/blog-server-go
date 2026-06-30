// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RoleMenusMenu 实体，对应 MySQL 表 x_role_menus_menu（结构对齐 Nest TypeORM）。
type RoleMenusMenu struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RoleMenusMenu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_role_menus_menu"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RoleMenusMenu) Fields() []ent.Field {
	return []ent.Field{
		field.Int("roleId").StorageKey("roleId"),
		field.String("menuId").StorageKey("menuId"),
	}
}
