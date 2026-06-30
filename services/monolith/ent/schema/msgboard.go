// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Msgboard 实体，对应 MySQL 表 x_msgboard（结构对齐 Nest TypeORM）。
type Msgboard struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Msgboard) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_msgboard"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Msgboard) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Msgboard) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("自身id"),
		field.String("name").StorageKey("name").Comment("昵称"),
		field.String("eamil").StorageKey("eamil").Comment("邮箱"),
		field.String("address").StorageKey("address").Comment("个人主页地址"),
		field.String("comment").StorageKey("comment").Comment("评论内容"),
		field.String("avatar").StorageKey("avatar").Comment("头像"),
		field.String("location").StorageKey("location").Comment("位置"),
		field.String("system").StorageKey("system").Comment("系统"),
		field.String("browser").StorageKey("browser").Comment("浏览器版本"),
		field.String("respondent").StorageKey("respondent").Comment("回复人名称").Optional().Nillable(),
		field.String("imgUrl").StorageKey("imgUrl").Comment("图片地址").Optional().Nillable(),
		field.String("ip").StorageKey("ip").Comment("ip地址").Optional().Nillable(),
		field.Int("pId").StorageKey("pId").Comment("父级id 0为一级评论").Default(0),
		field.Int("replyId").StorageKey("replyId").Comment("回复的评论的id").Optional().Nillable(),
		field.String("status").StorageKey("status").Comment("审核状态").Default("approved"),
	}
}
