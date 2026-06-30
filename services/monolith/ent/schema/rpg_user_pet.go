// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserPet 实体，对应 MySQL 表 x_rpg_user_pet（结构对齐 Nest TypeORM）。
type RpgUserPet struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserPet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_pet"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgUserPet) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserPet) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("所属用户ID"),
		field.String("petCode").StorageKey("petCode").Comment("宠物模板编码 → rpg_item_config.code (item_type=pet)"),
		field.Int("level").StorageKey("level").Comment("宠物等级").Default(1),
		field.Int("exp").StorageKey("exp").Comment("宠物经验").Default(0),
		field.Text("effectJson").StorageKey("effectJson").Comment("个体属性、技能、进化分支等").Optional().Nillable(),
		field.String("nickname").StorageKey("nickname").Comment("宠物昵称").Default(""),
	}
}
