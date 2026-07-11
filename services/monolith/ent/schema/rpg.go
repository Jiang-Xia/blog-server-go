// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Rpg 实体，对应 MySQL 表 x_rpg（结构对齐 Nest TypeORM）。
type Rpg struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Rpg) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg"},
	}
}

// Mixin 注入 Nest 公共时间戳字段（TimestampMixin）。
func (Rpg) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Rpg) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("RPG记录ID"),
		field.Int("uid").StorageKey("uid").Comment("关联user表id，一对一").Unique(),
		field.Int("exp").StorageKey("exp").Comment("当前经验值").Default(0),
		field.Int("level").StorageKey("level").Comment("当前等级").Default(1),
		field.Int("lifeValue").StorageKey("lifeValue").Comment("当前生命值，最大100").Default(100),
		field.Time("lastSignDate").StorageKey("lastSignDate").Comment("最后签到日期").Optional().Nillable(),
		field.Int("totalSignDays").StorageKey("totalSignDays").Comment("累计签到天数").Default(0),
		field.Int("consecutiveSignDays").StorageKey("consecutiveSignDays").Comment("连续签到天数，断签重置").Default(0),
		field.Time("banStartTime").StorageKey("banStartTime").Comment("禁言开始时间").Optional().Nillable(),
		field.Time("banEndTime").StorageKey("banEndTime").Comment("禁言结束时间").Optional().Nillable(),
		field.Int("sensitiveHitsCount").StorageKey("sensitiveHitsCount").Comment("累计敏感词命中次数").Default(0),
		field.Int("zeroLifeCount").StorageKey("zeroLifeCount").Comment("生命值归零累计次数").Default(0),
		field.Int("lotteryTickets").StorageKey("lotteryTickets").Comment("抽奖券数量").Default(0),
		field.Int("reputation").StorageKey("reputation").Comment("作者声望").Default(0),
		field.Int("lotteryPityCounter").StorageKey("lotteryPityCounter").Comment("史诗抽奖保底计数").Default(0),
		field.Text("effectJson").StorageKey("effectJson").Comment("扩展字段：活动加成快照、临时状态等").Optional().Nillable(),
		field.Int("lotteryLegendaryPityCounter").StorageKey("lotteryLegendaryPityCounter").Comment("传说抽奖保底计数").Default(0),
	}
}
