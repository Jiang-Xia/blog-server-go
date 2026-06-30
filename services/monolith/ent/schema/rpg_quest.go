// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgQuest 实体，对应 MySQL 表 x_rpg_quest（结构对齐 Nest TypeORM）。
type RpgQuest struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgQuest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_quest"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgQuest) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("自身id"),
		field.String("code").StorageKey("code").Comment("任务编码").Unique(),
		field.String("name").StorageKey("name").Comment("任务名称"),
		field.String("description").StorageKey("description").Comment("任务描述"),
		field.String("type").StorageKey("type").Comment("类型: daily/weekly").Default("daily"),
		field.String("targetAction").StorageKey("targetAction").Comment("目标行为: sign_in/comment/article/like/collect/msgboard"),
		field.Int("targetCount").StorageKey("targetCount").Comment("目标次数").Default(1),
		field.Int("expReward").StorageKey("expReward").Comment("经验奖励").Default(10),
		field.Int("hpReward").StorageKey("hpReward").Comment("完成恢复HP").Default(0),
		field.Int("currencyReward").StorageKey("currencyReward").Comment("完成奖励通用货币(钻石)").Default(0),
		field.Int("active").StorageKey("active").Comment("是否启用").Default(1),
		field.Int("sort").StorageKey("sort").Comment("排序权重").Default(10),
		field.Text("effectJson").StorageKey("effectJson").Comment("任务编码").Optional().Nillable(),
		field.String("questSubtype").StorageKey("questSubtype").Comment("daily/bounty/special").Default("daily"),
	}
}
