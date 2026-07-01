// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserLotteryRecord 实体，对应 MySQL 表 x_rpg_user_lottery_record（结构对齐 Nest TypeORM）。
type RpgUserLotteryRecord struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserLotteryRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_lottery_record"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgUserLotteryRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserLotteryRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("用户ID"),
		field.String("poolItemCode").StorageKey("poolItemCode").Comment("中奖物品编码 → rpg_item_config.code"),
		field.String("itemName").StorageKey("itemName").Comment("奖品名称快照"),
		field.Text("effectJson").StorageKey("effectJson").Comment("奖励详情快照").Optional().Nillable(),
		field.String("rarity").StorageKey("rarity").Comment("稀有度快照"),
	}
}
