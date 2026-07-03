// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ScheduledTask 实体，对应 MySQL 表 x_scheduled_task（结构对齐 Nest TypeORM）。
type ScheduledTask struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (ScheduledTask) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_scheduled_task"},
	}
}

// Mixin 对齐 Nest：仅有 createTime/updateTime，无 isDelete/version。
func (ScheduledTask) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (ScheduledTask) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("任务ID"),
		field.String("name").StorageKey("name").Comment("任务唯一标识").Unique(),
		field.String("description").StorageKey("description").Comment("任务描述").Default(""),
		field.String("cron").StorageKey("cron").Comment("Cron表达式"),
		field.String("cronHuman").StorageKey("cronHuman").Comment("Cron可读说明，如每天 10:00").Default(""),
		field.Int("enabled").StorageKey("enabled").Comment("是否启用").Default(1),
		field.Int("logRecording").StorageKey("logRecording").Comment("是否记录执行日志").Default(1),
		field.Int("sortOrder").StorageKey("sortOrder").Comment("排序值，越小越靠前").Default(0),
	}
}
