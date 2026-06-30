// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Menu 实体，对应 MySQL 表 x_menu（结构对齐 Nest TypeORM）。
type Menu struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Menu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_menu"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Menu) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("id").Comment("自身id"),
		field.String("pid").StorageKey("pid").Comment("父级id").Default("0"),
		field.String("path").StorageKey("path").Comment("菜单路由路径"),
		field.String("name").StorageKey("name").Comment("菜单英文名"),
		field.Int("order").StorageKey("order").Comment("用于菜单排序").Default(1),
		field.String("icon").StorageKey("icon").Comment("用于菜单图标").Default(""),
		field.String("locale").StorageKey("locale").Comment("用于菜单本地化").Default(""),
		field.Int("requiresAuth").StorageKey("requiresAuth").Comment("菜单鉴权").Default(1),
		field.String("filePath").StorageKey("filePath").Comment("菜单路由对应前端组件路径").Default(""),
		field.String("menuCnName").StorageKey("menuCnName").Comment("菜单中文名").Default("").Optional().Nillable(),
	}
}
