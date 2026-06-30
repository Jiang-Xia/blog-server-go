// 由 scripts/gen_ent_schema.go 生成；sensitive_word 含 createTime/updateTime（无 isDelete/version）。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// SensitiveWord 实体，对应 MySQL 表 x_sensitive_word（结构对齐 Nest TypeORM）。
type SensitiveWord struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (SensitiveWord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_sensitive_word"},
	}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (SensitiveWord) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("敏感词ID"),
		field.String("word").StorageKey("word").Comment("敏感词").Unique(),
		field.String("category").StorageKey("category").Comment("分类：广告/色情/赌博/自定义等").Default("自定义"),
		field.Int("status").StorageKey("status").Comment("状态：1=启用 0=禁用").Default(1),
		field.Int("level").StorageKey("level").Comment("等级：1重/2中/3轻").Default(2),
		field.Int("hpPenalty").StorageKey("hpPenalty").Comment("扣血量").Default(20),
		field.Int("needReview").StorageKey("needReview").Comment("是否进审核：1是0否").Default(1),
		field.Int("action").StorageKey("action").Comment("1替换/2拒绝/3仅记录").Default(1),
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
		field.Time("updateTime").StorageKey("updateTime").Comment("更新时间").Default(time.Now).UpdateDefault(time.Now),
	}
}
