// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserLoadout 实体，对应 MySQL 表 x_rpg_user_loadout（结构对齐 Nest TypeORM）。
type RpgUserLoadout struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserLoadout) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_loadout"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserLoadout) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("用户ID，一人一行").Unique(),
		field.String("titleCode").StorageKey("titleCode").Comment("当前称号 → rpg_item_config.code").Optional().Nillable(),
		field.String("avatarFrameCode").StorageKey("avatarFrameCode").Comment("当前头像框 → rpg_item_config.code").Optional().Nillable(),
		field.Int("petId").StorageKey("petId").Comment("当前出战宠物 → rpg_user_pet.id").Optional().Nillable(),
		field.Text("effectJson").StorageKey("effectJson").Comment("扩展：多套预设、临时外观等").Optional().Nillable(),
	}
}
