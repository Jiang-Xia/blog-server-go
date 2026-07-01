// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgUserSocialLog 实体，对应 MySQL 表 x_rpg_user_social_log（结构对齐 Nest TypeORM）。
type RpgUserSocialLog struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgUserSocialLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_user_social_log"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgUserSocialLog) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgUserSocialLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.Int("fromUid").StorageKey("fromUid").Comment("发起者UID"),
		field.Int("toUid").StorageKey("toUid").Comment("目标UID"),
		field.Int("costCurrency").StorageKey("costCurrency").Comment("消耗通用货币(钻石)").Default(0),
		field.Int("hpDelta").StorageKey("hpDelta").Comment("HP变化").Default(0),
		field.String("action").StorageKey("action").Comment("cheer/egg/flower"),
	}
}
