// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// User 实体，对应 MySQL 表 x_user（结构对齐 Nest TypeORM）。
type User struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_user"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键id"),
		field.String("status").StorageKey("status").Comment("用户状态").Default("active"),
		field.Text("password").StorageKey("password").Comment("加密后的密码"),
		field.Text("salt").StorageKey("salt").Comment("加密盐"),
		field.String("intro").StorageKey("intro").Comment("简介或者个性签名").Default(""),
		field.String("avatar").StorageKey("avatar").Comment("头像").Default(""),
		field.String("homepage").StorageKey("homepage").Comment("个人主页").Default(""),
		field.String("email").StorageKey("email").Comment("邮箱地址").Unique().Optional().Nillable(),
		field.String("githubId").StorageKey("githubId").Comment("githubId").Unique().Optional().Nillable(),
		field.String("wechatOpenId").StorageKey("wechatOpenId").Comment("微信小程序 openid").Unique().Optional().Nillable(),
		field.String("username").StorageKey("username").Comment("用户名").Unique().Optional().Nillable(),
		field.Int("deptId").StorageKey("deptId").Comment("部门ID").Optional().Nillable(),
		field.String("nickname").StorageKey("nickname").Comment("昵称"),
	}
}
