// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserBuff 实体，对应 MySQL 表 x_rpg_user_buff（结构对齐 Nest TypeORM）。
type RpgUserBuff struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserBuff) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_buff"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgUserBuff) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserBuff) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("uid").StorageKey("uid").Comment("用户ID"),
		field.String("buffCode").StorageKey("buffCode").Comment("Buff编码"),
		field.String("buffType").StorageKey("buffType").Comment("Buff类型: exp_boost / hp_regen / ban_reduction / shield / lucky"),
		field.String("name").StorageKey("name").Comment("Buff名称"),
		field.String("description").StorageKey("description").Comment("Buff描述"),
		field.Float("value").StorageKey("value").Comment("效果值（如1.5=50%加成）"),
		field.Time("expireAt").StorageKey("expireAt").Comment("过期时间"),
		field.Int("remainingUses").StorageKey("remainingUses").Comment("剩余使用次数，-1=不限").Default(1),
		field.Int("isActive").StorageKey("isActive").Comment("是否激活").Default(1),
		field.String("sourceType").StorageKey("sourceType").Comment("来源类型").Optional().Nillable(),
		field.Int("sourceId").StorageKey("sourceId").Comment("来源ID").Optional().Nillable(),
		field.Text("effectJson").StorageKey("effectJson").Comment("运行时快照扩展").Optional().Nillable(),
		field.String("triggerMode").StorageKey("triggerMode").Comment("auto/manual/passive").Default("auto"),
	}
}
