// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserQuestProgress 实体，对应 MySQL 表 x_rpg_user_quest_progress（结构对齐 Nest TypeORM）。
type RpgUserQuestProgress struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserQuestProgress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_quest_progress"},
	}
}

// Mixin 注入 Nest 公共时间戳字段（TimestampMixin）。
func (RpgUserQuestProgress) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserQuestProgress) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("用户ID"),
		field.String("questCode").StorageKey("questCode").Comment("关联 Quest.code"),
		field.Int("progress").StorageKey("progress").Comment("当前进度").Default(0),
		field.Int("completed").StorageKey("completed").Comment("是否已完成").Default(0),
		field.Int("claimed").StorageKey("claimed").Comment("奖励是否已领取").Default(0),
		field.Time("questDate").StorageKey("questDate").Comment("任务日期（每日重置）"),
	}
}
