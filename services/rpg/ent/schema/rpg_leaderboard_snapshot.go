// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgLeaderboardSnapshot 实体，对应 MySQL 表 x_rpg_leaderboard_snapshot（结构对齐 Nest TypeORM）。
type RpgLeaderboardSnapshot struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgLeaderboardSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_leaderboard_snapshot"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgLeaderboardSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgLeaderboardSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("用户UID"),
		field.Int("score").StorageKey("score").Comment("分数").Default(0),
		field.Int("rank").StorageKey("rank").Comment("排名").Default(0),
		field.String("periodType").StorageKey("periodType").Comment("week/month/season"),
		field.String("periodKey").StorageKey("periodKey").Comment("周期标识"),
		field.String("scoreType").StorageKey("scoreType").Comment("exp/reputation/currency/level"),
	}
}
