// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgItemConfig 实体，对应 MySQL 表 x_rpg_item_config（结构对齐 Nest TypeORM）。
type RpgItemConfig struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgItemConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_item_config"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgItemConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgItemConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("ID"),
		field.String("code").StorageKey("code").Comment("全局唯一物品编码").Unique(),
		field.String("name").StorageKey("name").Comment("显示名称"),
		field.Int("sort").StorageKey("sort").Comment("排序权重").Default(10),
		field.Int("active").StorageKey("active").Comment("是否启用").Default(1),
		field.Int("isHidden").StorageKey("isHidden").Comment("隐藏成就：1是0否").Default(0),
		field.Text("effectJson").StorageKey("effectJson").Comment("类型相关扩展配置").Optional().Nillable(),
		field.String("itemType").StorageKey("itemType").Comment("物品类型: title/avatar_frame/pet/equipment/achievement/buff/currency/consumable"),
		field.String("description").StorageKey("description").Comment("描述"),
		field.String("category").StorageKey("category").Comment("分类（成就: creation/social/sign…）").Default(""),
		field.String("icon").StorageKey("icon").Comment("图标ID").Default("default"),
		field.String("rarity").StorageKey("rarity").Comment("稀有度: common/rare/epic/legendary").Default("common"),
	}
}
