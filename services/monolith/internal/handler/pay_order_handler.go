// Package handler 支付订单管理 HTTP 端点，路径对齐 Nest PayOrderController。
package handler

import (
	"context"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	paysvc "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/service"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/cloudwego/hertz/pkg/app"
)

// PayOrderHandler /pay/order/* 路由（需 JWT）。
type PayOrderHandler struct {
	svc *paysvc.PayOrderService
	jwt *auth.JWTService
}

// NewPayOrderHandler 构造 PayOrderHandler。
func NewPayOrderHandler(svc *paysvc.PayOrderService, jwt *auth.JWTService) *PayOrderHandler {
	return &PayOrderHandler{svc: svc, jwt: jwt}
}

func (h *PayOrderHandler) Create(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	dto := paysvc.TradeCreateDTO{
		OutTradeNo:  strField(body, "out_trade_no"),
		Subject:     strField(body, "subject"),
		TotalAmount: strField(body, "total_amount"),
		BuyerOpenID: strField(body, "buyer_open_id"),
	}
	if ep, ok := body["extend_params"].(map[string]interface{}); ok {
		dto.ExtendParams = ep
	}
	data, err := h.svc.CreateOrder(ctx, dto)
	handleAdminResult(ctx, c, data, err)
}

func (h *PayOrderHandler) List(ctx context.Context, c *app.RequestContext) {
	dto := paysvc.OrderListDTO{
		Page:        queryInt(c, "page", 1),
		PageSize:    queryInt(c, "pageSize", 20),
		OutTradeNo:  string(c.Query("outTradeNo")),
		Status:      string(c.Query("status")),
		Channel:     string(c.Query("channel")),
		Subject:     string(c.Query("subject")),
		BizType:     string(c.Query("bizType")),
		RechargeUID: string(c.Query("rechargeUid")),
	}
	data, err := h.svc.GetOrderList(ctx, dto)
	handleAdminResult(ctx, c, data, err)
}

func (h *PayOrderHandler) Refund(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	outTradeNo := strField(body, "out_trade_no")
	if outTradeNo == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "商户订单号不能为空"))
		return
	}
	data, err := h.svc.RefundOrder(ctx, outTradeNo, strField(body, "refund_amount"), strField(body, "refund_reason"))
	handleAdminResult(ctx, c, data, err)
}

func (h *PayOrderHandler) Close(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	outTradeNo := strField(body, "out_trade_no")
	if outTradeNo == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "商户订单号不能为空"))
		return
	}
	data, err := h.svc.CloseOrder(ctx, outTradeNo)
	handleAdminResult(ctx, c, data, err)
}

func (h *PayOrderHandler) Query(ctx context.Context, c *app.RequestContext) {
	outTradeNo := string(c.Query("out_trade_no"))
	if outTradeNo == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "商户订单号不能为空"))
		return
	}
	data, err := h.svc.QueryAndUpdateOrder(ctx, outTradeNo)
	handleAdminResult(ctx, c, data, err)
}

func (h *PayOrderHandler) Delete(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	ids := parseIDList(body["ids"])
	data, err := h.svc.DeleteOrders(ctx, ids)
	handleAdminResult(ctx, c, data, err)
}

func (h *PayOrderHandler) MarkRechargeFulfilled(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	outTradeNo := strField(body, "out_trade_no")
	if outTradeNo == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "商户订单号不能为空"))
		return
	}
	data, err := h.svc.MarkRpgRechargeManuallyFulfilled(ctx, outTradeNo)
	handleAdminResult(ctx, c, data, err)
}

func queryInt(c *app.RequestContext, key string, def int) int {
	v := string(c.Query(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseIDList(raw interface{}) []int {
	switch v := raw.(type) {
	case []interface{}:
		out := make([]int, 0, len(v))
		for _, item := range v {
			if n, ok := toInt(item); ok {
				out = append(out, n)
			}
		}
		return out
	case []int:
		return v
	default:
		return nil
	}
}
