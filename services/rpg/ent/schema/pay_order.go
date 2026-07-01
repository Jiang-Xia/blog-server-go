// 由 scripts/gen_ent_schema.go 生成，请勿手改。
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// PayOrder 实体，对应 MySQL 表 x_pay_order（结构对齐 Nest TypeORM）。
type PayOrder struct{ ent.Schema }

// Annotations 指定 Ent 映射的真实表名（含 x_ 前缀）。
func (PayOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "x_pay_order"},
	}
}

// Mixin 注入 TypeORM 公共时间戳与软删除字段。
func (PayOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{TimestampMixin{}}
}

// Fields 定义表列，StorageKey 保持与 Nest camelCase 列名一致。
func (PayOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("id").Comment("订单ID"),
		field.String("outTradeNo").StorageKey("outTradeNo").Comment("商户订单号").Unique(),
		field.String("tradeNo").StorageKey("tradeNo").Comment("第三方交易号").Default(""),
		field.String("subject").StorageKey("subject").Comment("订单标题").Default(""),
		field.Float("totalAmount").StorageKey("totalAmount").Comment("订单总金额（元）").Default(0.00),
		field.String("buyerOpenId").StorageKey("buyerOpenId").Comment("买家标识（openid等）").Default(""),
		field.String("status").StorageKey("status").Comment("订单状态").Default("PENDING"),
		field.Float("refundAmount").StorageKey("refundAmount").Comment("退款金额（元）").Default(0.00),
		field.String("channel").StorageKey("channel").Comment("支付渠道").Default("alipay"),
		field.JSON("extendParams", map[string]interface{}{}).StorageKey("extendParams").Comment("扩展参数").Optional(),
	}
}
