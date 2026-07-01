// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserGuildMember 实体，对应 MySQL 表 x_rpg_user_guild_member（结构对齐 Nest TypeORM）。
type RpgUserGuildMember struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserGuildMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_guild_member"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserGuildMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("guildId").StorageKey("guildId").Comment("公会ID → rpg_guild.id"),
		field.Int("uid").StorageKey("uid").Comment("成员用户ID，一人一会").Unique(),
		field.Time("joinTime").StorageKey("joinTime").Comment("加入时间"),
		field.String("role").StorageKey("role").Comment("角色: leader/officer/member").Default("member"),
	}
}
