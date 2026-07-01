// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Privilege 实体，对应 MySQL 表 x_privilege（结构对齐 Nest TypeORM）。
type Privilege struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Privilege) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_privilege"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Privilege) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Privilege) Fields() []ent.Field {
	return []ent.Field{
		field.String("privilegeName").StorageKey("privilegeName").Comment("权限名称"),
		field.String("privilegeCode").StorageKey("privilegeCode").Comment("权限识别码"),
		field.String("privilegePage").StorageKey("privilegePage").Comment("所属页面(菜单id)"),
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("isVisible").StorageKey("isVisible").Comment("是否可见").Default(1),
		field.String("pathPattern").StorageKey("pathPattern").Comment("路径模式，如 /api/users/:id"),
		field.String("httpMethod").StorageKey("httpMethod").Comment("HTTP方法，*表示全部"),
		field.Int("isPublic").StorageKey("isPublic").Comment("是否公开接口").Default(0),
		field.Int("requireOwnership").StorageKey("requireOwnership").Comment("是否需要检查资源所有权").Default(0),
		field.String("description").StorageKey("description").Comment("描述").Optional().Nillable(),
	}
}
