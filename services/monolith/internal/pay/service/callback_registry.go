// callback_registry 支付订单 PAID 后业务回调注册表（如 RPG 充值发钻）。
package service

import (
	"context"
	"sync"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

// PayOrderPaidHandler 支付订单变为 PAID 后的业务回调。
type PayOrderPaidHandler func(ctx context.Context, order *ent.PayOrder) error

var (
	paidHandlers   []PayOrderPaidHandler
	paidHandlersMu sync.RWMutex
)

// RegisterPayPaidCallback 注册 PAID 回调（如 RPG 充值发钻）。
func RegisterPayPaidCallback(h PayOrderPaidHandler) {
	paidHandlersMu.Lock()
	defer paidHandlersMu.Unlock()
	paidHandlers = append(paidHandlers, h)
}

// InvokePayPaidCallbacks 依次调用已注册的 PAID 回调。
func InvokePayPaidCallbacks(ctx context.Context, order *ent.PayOrder) {
	paidHandlersMu.RLock()
	handlers := append([]PayOrderPaidHandler(nil), paidHandlers...)
	paidHandlersMu.RUnlock()
	for _, h := range handlers {
		_ = h(ctx, order)
	}
}
