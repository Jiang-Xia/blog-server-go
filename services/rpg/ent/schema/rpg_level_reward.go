// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgLevelReward 实体，对应 MySQL 表 x_rpg_level_reward（结构对齐 Nest TypeORM）。
type RpgLevelReward struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgLevelReward) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_level_reward"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgLevelReward) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("level").StorageKey("level").Comment("触发等级").Unique(),
		field.String("avatarFrame").StorageKey("avatarFrame").Comment("头像框物品 code").Default(""),
		field.String("title").StorageKey("title").Comment("称号物品 code").Default(""),
		field.Int("currencyReward").StorageKey("currencyReward").Comment("钻石奖励").Default(0),
		field.Int("active").StorageKey("active").Comment("是否启用").Default(1),
		field.Int("sort").StorageKey("sort").Comment("排序").Default(10),
	}
}
