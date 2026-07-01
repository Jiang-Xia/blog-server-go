// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RpgActivity 实体，对应 MySQL 表 x_rpg_activity（结构对齐 Nest TypeORM）。
type RpgActivity struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (RpgActivity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_rpg_activity"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (RpgActivity) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (RpgActivity) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("自身id"),
		field.String("code").StorageKey("code").Comment("活动编码").Unique(),
		field.String("name").StorageKey("name").Comment("活动名称"),
		field.Time("startTime").StorageKey("startTime").Comment("开始时间"),
		field.Time("endTime").StorageKey("endTime").Comment("结束时间"),
		field.Float("expBuffRate").StorageKey("expBuffRate").Comment("经验加成倍率（1.0=无加成）").Default(1),
		field.Int("active").StorageKey("active").Comment("是否启用").Default(1),
		field.Text("effectJson").StorageKey("effectJson").Comment("活动编码").Optional().Nillable(),
		field.String("description").StorageKey("description").Comment("活动描述").Default(""),
		field.String("activityType").StorageKey("activityType").Comment("活动类型: season/event/festival").Default("event"),
		field.String("posterUrl").StorageKey("posterUrl").Comment("赛季海报URL").Default(""),
	}
}
