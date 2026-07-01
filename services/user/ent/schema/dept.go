// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Dept 实体，对应 MySQL 表 x_dept（结构对齐 Nest TypeORM）。
type Dept struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Dept) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_dept"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Dept) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Dept) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键 ID"),
		field.String("deptName").StorageKey("deptName").Comment("部门名称"),
		field.String("deptCode").StorageKey("deptCode").Comment("部门编码").Unique(),
		field.Int("parentId").StorageKey("parentId").Comment("父级部门ID").Default(0),
		field.String("leaderId").StorageKey("leaderId").Comment("部门负责人ID").Optional().Nillable(),
		field.String("leaderName").StorageKey("leaderName").Comment("部门负责人姓名").Optional().Nillable(),
		field.Int("orderNum").StorageKey("orderNum").Comment("部门排序").Default(0),
		field.Int("status").StorageKey("status").Comment("部门状态").Default(1),
		field.String("remark").StorageKey("remark").Comment("部门描述").Optional().Nillable(),
	}
}
