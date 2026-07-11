// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// TimeMixin 对应 Nest TypeORM 公共字段：createTime / updateTime / isDelete / version。
type TimeMixin struct{ mixin.Schema }

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
		field.Time("updateTime").StorageKey("updateTime").Comment("更新时间").Default(time.Now).UpdateDefault(time.Now),
		field.Bool("isDelete").StorageKey("isDelete").Comment("软删除标记").Default(false),
		field.Int("version").StorageKey("version").Comment("乐观锁版本号").Default(0),
	}
}

// TimestampMixin 仅 createTime/updateTime（RPG 等 Nest 表无 isDelete/version）。
type TimestampMixin struct{ mixin.Schema }

func (TimestampMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
		field.Time("updateTime").StorageKey("updateTime").Comment("更新时间").Default(time.Now).UpdateDefault(time.Now),
	}
}

// CreateTimeMixin 仅 createTime（部分 Nest 表无 updateTime/isDelete）。
type CreateTimeMixin struct{ mixin.Schema }

func (CreateTimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createTime").StorageKey("createTime").Comment("创建时间").Default(time.Now).Immutable(),
	}
}
