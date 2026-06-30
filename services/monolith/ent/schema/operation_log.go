// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// OperationLog 实体，对应 MySQL 表 x_operation_log（结构对齐 Nest TypeORM）。
type OperationLog struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (OperationLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_operation_log"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (OperationLog) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (OperationLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("操作日志ID"),
		field.Int("userId").StorageKey("userId").Comment("操作人ID"),
		field.String("username").StorageKey("username").Comment("操作人用户名").Default(""),
		field.String("module").StorageKey("module").Comment("操作模块，如 article, user, role"),
		field.String("action").StorageKey("action").Comment("操作类型，如 create, update, delete"),
		field.String("method").StorageKey("method").Comment("HTTP方法：POST/PUT/PATCH/DELETE"),
		field.String("path").StorageKey("path").Comment("请求路径"),
		field.String("description").StorageKey("description").Comment("操作描述").Default(""),
		field.String("ip").StorageKey("ip").Comment("操作人IP").Default(""),
		field.Text("requestBody").StorageKey("requestBody").Comment("请求体摘要（脱敏后）").Optional().Nillable(),
		field.Int("statusCode").StorageKey("statusCode").Comment("响应状态码").Default(200),
	}
}
