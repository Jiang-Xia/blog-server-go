// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgLotteryPool 实体，对应 MySQL 表 x_rpg_lottery_pool（结构对齐 Nest TypeORM）。
type RpgLotteryPool struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgLotteryPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_lottery_pool"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgLotteryPool) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("自身id"),
		field.String("itemCode").StorageKey("itemCode").Comment("关联奖品 → rpg_item_config.code"),
		field.Float("probability").StorageKey("probability").Comment("概率权重(0-1)"),
		field.Int("active").StorageKey("active").Comment("是否启用").Default(1),
		field.Int("sort").StorageKey("sort").Comment("排序").Default(10),
		field.Text("effectJson").StorageKey("effectJson").Comment("关联奖品 → rpg_item_config.code").Optional().Nillable(),
		field.String("rarity").StorageKey("rarity").Comment("展示稀有度: common/rare/epic/legendary"),
	}
}
