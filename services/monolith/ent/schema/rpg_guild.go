// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgGuild 实体，对应 MySQL 表 x_rpg_guild（结构对齐 Nest TypeORM）。
type RpgGuild struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgGuild) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_guild"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgGuild) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgGuild) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("自身id"),
		field.Int("leaderUid").StorageKey("leaderUid").Comment("会长用户ID"),
		field.Text("announcement").StorageKey("announcement").Comment("公会公告").Optional().Nillable(),
		field.Int("memberCount").StorageKey("memberCount").Comment("成员数（冗余字段）").Default(1),
		field.Text("effectJson").StorageKey("effectJson").Comment("公会名称").Optional().Nillable(),
		field.String("name").StorageKey("name").Comment("公会名称").Unique(),
	}
}
