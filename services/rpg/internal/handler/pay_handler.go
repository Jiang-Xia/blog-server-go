// Package handler 支付 HTTP 端点，路径对齐 Nest PayController。
package handler

import (
	"context"
	"fmt"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	paysvc "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/pay/service"
	"github.com/cloudwego/hertz/pkg/app"
)

// PayHandler /pay/* 路由。
type PayHandler struct {
	svc *paysvc.PayService
}

// NewPayHandler 构造 PayHandler。
func NewPayHandler(svc *paysvc.PayService) *PayHandler {
	return &PayHandler{svc: svc}
}

func (h *PayHandler) TradeCreate(ctx context.Context, c *app.RequestContext) {
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
	data, err := h.svc.CCreateOrder(ctx, dto)
	handleAdminResult(ctx, c, data, err)
}

func (h *PayHandler) TradeQuery(ctx context.Context, c *app.RequestContext) {
	outTradeNo := string(c.Query("out_trade_no"))
	if outTradeNo == "" {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, "商户订单号不能为空"))
		return
	}
	data, err := h.svc.CQueryOrder(ctx, paysvc.TradeQueryDTO{OutTradeNo: outTradeNo})
	handleAdminResult(ctx, c, data, err)
}

func (h *PayHandler) TradeRefund(ctx context.Context, c *app.RequestContext) {
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
	dto := paysvc.TradeRefundDTO{
		OutTradeNo:   outTradeNo,
		RefundAmount: strField(body, "refund_amount"),
		RefundReason: strField(body, "refund_reason"),
		OutRequestNo: strField(body, "out_request_no"),
	}
	data, err := h.svc.CRefundOrder(ctx, dto)
	handleAdminResult(ctx, c, data, err)
}

func (h *PayHandler) TradeClose(ctx context.Context, c *app.RequestContext) {
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
	data, err := h.svc.CCloseOrder(ctx, paysvc.TradeCloseDTO{OutTradeNo: outTradeNo})
	handleAdminResult(ctx, c, data, err)
}

func (h *PayHandler) GetOpenID(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	data, err := h.svc.GetOpenIDByCode(ctx, paysvc.GetOpenIDDTO{Code: strField(body, "code")})
	handleAdminResult(ctx, c, data, err)
}

func (h *PayHandler) H5OpenMini(ctx context.Context, c *app.RequestContext) {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	dto := paysvc.H5OpenMiniDTO{
		Type:    strField(body, "type"),
		AppID:   strField(body, "appId"),
		Page:    strField(body, "page"),
		Version: strField(body, "version"),
	}
	if q, ok := body["query"].(map[string]interface{}); ok {
		dto.Query = map[string]string{}
		for k, v := range q {
			dto.Query[k] = fmt.Sprint(v)
		}
	}
	data, err := h.svc.BuildH5OpenMiniURL(ctx, dto)
	handleAdminResult(ctx, c, data, err)
}

// Notice 支付宝异步通知；成功返回纯文本 success。
func (h *PayHandler) Notice(ctx context.Context, c *app.RequestContext) {
	postData := map[string]string{}
	c.Request.PostArgs().VisitAll(func(k, v []byte) {
		postData[string(k)] = string(v)
	})
	if len(postData) == 0 {
		var body map[string]string
		if err := c.Bind(&body); err == nil {
			postData = body
		}
	}
	ok, _ := h.svc.HandleAlipayNotify(ctx, postData)
	if ok {
		c.String(200, "success")
		return
	}
	c.String(200, "failure")
}
