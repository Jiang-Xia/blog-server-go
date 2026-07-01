// order_status 支付订单状态与渠道常量，对齐 Nest PayOrderStatus。
package constants

// 订单状态，对齐 Nest PayOrderStatus。
const (
	OrderStatusPending  = "PENDING"
	OrderStatusPaid     = "PAID"
	OrderStatusRefunded = "REFUNDED"
	OrderStatusClosed   = "CLOSED"
	OrderStatusFailed   = "FAILED"
)

// 支付渠道，对齐 Nest PayChannel。
const (
	ChannelAlipay = "alipay"
	ChannelWechat = "wechat"
)
