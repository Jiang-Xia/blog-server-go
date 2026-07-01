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
