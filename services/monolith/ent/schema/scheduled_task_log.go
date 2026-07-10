// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ScheduledTaskLog 实体，对应 MySQL 表 x_scheduled_task_log（结构对齐 Nest TypeORM）。
type ScheduledTaskLog struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (ScheduledTaskLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_scheduled_task_log"},
	}
}

// Mixin 对齐 Nest：scheduled_task_log 仅有 createTime。
func (ScheduledTaskLog) Mixin() []ent.Mixin {
	return nil
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (ScheduledTaskLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("任务日志ID"),
		field.Time("createTime").StorageKey("createTime").Comment("创建时间"),
		field.String("taskName").StorageKey("taskName").Comment("任务名称"),
		field.String("status").StorageKey("status").Comment("执行状态：success/failed"),
		field.Time("startTime").StorageKey("startTime").Comment("任务开始时间"),
		field.Time("endTime").StorageKey("endTime").Comment("任务结束时间").Optional().Nillable(),
		field.Text("result").StorageKey("result").Comment("执行结果摘要（JSON字符串）").Optional().Nillable(),
		field.Text("errorMessage").StorageKey("errorMessage").Comment("错误信息").Optional().Nillable(),
	}
}
