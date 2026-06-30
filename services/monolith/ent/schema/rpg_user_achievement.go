// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserAchievement 实体，对应 MySQL 表 x_rpg_user_achievement（结构对齐 Nest TypeORM）。
type RpgUserAchievement struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserAchievement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_achievement"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgUserAchievement) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserAchievement) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("用户ID"),
		field.String("achievementCode").StorageKey("achievementCode").Comment("关联 rpg_item_config.code (item_type=achievement)"),
		field.Int("progress").StorageKey("progress").Comment("当前进度").Default(0),
		field.Int("completed").StorageKey("completed").Comment("是否已完成").Default(0),
		field.Time("completedAt").StorageKey("completedAt").Comment("完成时间").Optional().Nillable(),
	}
}
