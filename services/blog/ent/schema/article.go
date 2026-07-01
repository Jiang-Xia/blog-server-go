// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Article 实体，对应 MySQL 表 x_article（结构对齐 Nest TypeORM）。
type Article struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (Article) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_article"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (Article) Mixin() []ent.Mixin {
	return []ent.Mixin{TimeMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (Article) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("主键id"),
		field.Text("uTime").StorageKey("uTime").Comment("文章的更新时间"),
		field.Int("uid").StorageKey("uid").Comment("用户id"),
		field.Int("deptId").StorageKey("deptId").Comment("所属机构ID").Optional().Nillable(),
		field.Int("topping").StorageKey("topping").Default(0),
		field.Text("title").StorageKey("title").Comment("文章标题"),
		field.Text("cover").StorageKey("cover").Comment("封面图"),
		field.Int("likes").StorageKey("likes").Comment("喜欢/点赞数").Default(0),
		field.Int("views").StorageKey("views").Comment("阅读量").Default(0),
		field.Int("articleExp").StorageKey("articleExp").Comment("文章经验").Default(0),
		field.Int("articleLevel").StorageKey("articleLevel").Comment("文章等级").Default(1),
		field.Int("reputationGained").StorageKey("reputationGained").Comment("该文贡献作者声望").Default(0),
		field.Int("isMasterpiece").StorageKey("isMasterpiece").Comment("神作标记").Default(0),
		field.Int("tipTotal").StorageKey("tipTotal").Comment("累计被打赏碎片").Default(0),
		field.String("status").StorageKey("status").Comment("文章状态: draft-草稿, publish-已发布, scheduled-定时发布").Default("publish"),
		field.Text("description").StorageKey("description").Comment("文章描述"),
		field.Text("contentHtml").StorageKey("contentHtml").Comment("文章html"),
		field.Int("useArticles").StorageKey("useArticles").Optional().Nillable(),
		field.String("articles").StorageKey("articles").Optional().Nillable(),
		field.Text("content").StorageKey("content").Comment("文章内容"),
		field.Time("scheduledPublishAt").StorageKey("scheduledPublishAt").Comment("定时发布时间").Optional().Nillable(),
	}
}
