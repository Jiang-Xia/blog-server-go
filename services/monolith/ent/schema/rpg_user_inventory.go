// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserInventory 实体，对应 MySQL 表 x_rpg_user_inventory（结构对齐 Nest TypeORM）。
type RpgUserInventory struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserInventory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_inventory"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserInventory) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("用户ID"),
		field.String("itemCode").StorageKey("itemCode").Comment("关联 rpg_item_config.code"),
		field.Int("quantity").StorageKey("quantity").Comment("数量（碎片可堆叠，称号/头像框为1）").Default(1),
		field.Text("effectJson").StorageKey("effectJson").Comment("实例级扩展（强化、绑定属性等）").Optional().Nillable(),
		field.Time("acquiredAt").StorageKey("acquiredAt").Comment("获得时间"),
		field.String("source").StorageKey("source").Comment("来源: level_up/lottery/quest/admin等").Default("system"),
	}
}
