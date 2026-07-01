// Package service 管理端支付订单业务。
package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	payconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/constants"
	payrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"go.uber.org/zap"
)

// RechargeManualFulfiller 管理端手工补钻标记接口（由 RPG recharge 实现）。
type RechargeManualFulfiller interface {
	// MarkManualFulfillmentAndNotify 标记订单已手工补钻并推送 WS 通知。
	MarkManualFulfillmentAndNotify(ctx context.Context, outTradeNo string) (interface{}, error)
}

// OrderListDTO 管理端订单列表查询。
type OrderListDTO struct {
	Page        int    `json:"page"`
	PageSize    int    `json:"pageSize"`
	OutTradeNo  string `json:"outTradeNo"`
	Status      string `json:"status"`
	Channel     string `json:"channel"`
	Subject     string `json:"subject"`
	BizType     string `json:"bizType"`
	RechargeUID string `json:"rechargeUid"`
}

// PayOrderService 管理端订单 CRUD 与退款关单。
type PayOrderService struct {
	repo      *payrepo.PayOrderRepo
	pay       *PayService
	recharge  RechargeManualFulfiller
	log       *zap.Logger
}

// NewPayOrderService 构造 PayOrderService。
func NewPayOrderService(repo *payrepo.PayOrderRepo, pay *PayService, recharge RechargeManualFulfiller, log *zap.Logger) *PayOrderService {
	return &PayOrderService{repo: repo, pay: pay, recharge: recharge, log: log}
}

// CreateOrder 创建订单：支付宝 + 入库 + 轮询。
func (s *PayOrderService) CreateOrder(ctx context.Context, dto TradeCreateDTO) (interface{}, error) {
	outTradeNo := dto.OutTradeNo
	if outTradeNo == "" {
		outTradeNo = s.generateTradeNo()
		dto.OutTradeNo = outTradeNo
	}

	tradeResult, err := s.pay.CallAlipayTradeCreate(ctx, dto)
	if err != nil {
		s.log.Error("admin create alipay failed", zap.Error(err))
		return map[string]interface{}{
			"alipaySuccess": false,
			"localSuccess":  false,
			"message":       "支付宝创建交易失败",
		}, nil
	}

	tradeNo, _ := tradeResult["tradeNo"].(string)
	localSuccess := false
	existing, err := s.repo.FindByOutTradeNo(ctx, outTradeNo)
	if ent.IsNotFound(err) {
		amount, _ := parseFloat(dto.TotalAmount)
		_, createErr := s.repo.Create(ctx, payrepo.CreateInput{
			OutTradeNo:   outTradeNo,
			TradeNo:      tradeNo,
			Subject:      dto.Subject,
			TotalAmount:  amount,
			BuyerOpenID:  dto.BuyerOpenID,
			Status:       payconst.OrderStatusPending,
			Channel:      payconst.ChannelAlipay,
			ExtendParams: dto.ExtendParams,
		})
		localSuccess = createErr == nil
	} else if err == nil {
		localSuccess = true
		_ = existing
	}

	if localSuccess {
		s.pay.StartPolling(outTradeNo)
		return map[string]interface{}{
			"outTradeNo": outTradeNo,
			"message":    "创建订单成功",
		}, nil
	}
	return map[string]interface{}{
		"alipaySuccess": true,
		"localSuccess":  false,
		"message":       "支付宝交易已创建，但本地入库失败",
	}, nil
}

// GetOrderList 分页查询订单。
func (s *PayOrderService) GetOrderList(ctx context.Context, dto OrderListDTO) (interface{}, error) {
	bizType := dto.BizType
	if bizType == "" && dto.RechargeUID != "" {
		bizType = payconst.PAY_BIZ_RPG_RECHARGE
	}
	rows, total, err := s.repo.List(ctx, payrepo.PayOrderListFilter{
		Page:        dto.Page,
		PageSize:    dto.PageSize,
		OutTradeNo:  dto.OutTradeNo,
		Status:      dto.Status,
		Channel:     dto.Channel,
		Subject:     dto.Subject,
		BizType:     bizType,
		RechargeUID: dto.RechargeUID,
	})
	if err != nil {
		return nil, err
	}
	list := make([]interface{}, 0, len(rows))
	for _, o := range rows {
		list = append(list, formatPayOrderForAdminList(o))
	}
	page, pageSize := dto.Page, dto.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return map[string]interface{}{
		"list": list,
		"pagination": map[string]interface{}{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	}, nil
}

func formatPayOrderForAdminList(order *ent.PayOrder) map[string]interface{} {
	isRpg := bizTypeOf(order) == payconst.PAY_BIZ_RPG_RECHARGE
	orderSource := "external"
	if isRpg {
		orderSource = payconst.PAY_BIZ_RPG_RECHARGE
	}
	var rechargeInfo interface{}
	if isRpg && order.ExtendParams != nil {
		rechargeInfo = map[string]interface{}{
			"uid":       toIntFromAny(order.ExtendParams["uid"]),
			"diamonds":  toIntFromAny(order.ExtendParams["diamonds"]),
			"fulfilled": toBool(order.ExtendParams["fulfilled"]),
		}
	}
	return map[string]interface{}{
		"id":           order.ID,
		"outTradeNo":   order.OutTradeNo,
		"tradeNo":      order.TradeNo,
		"subject":      order.Subject,
		"totalAmount":  order.TotalAmount,
		"buyerOpenId":  order.BuyerOpenId,
		"status":       order.Status,
		"refundAmount": order.RefundAmount,
		"channel":      order.Channel,
		"extendParams": order.ExtendParams,
		"createTime":   order.CreateTime,
		"updateTime":   order.UpdateTime,
		"orderSource":  orderSource,
		"rechargeInfo": rechargeInfo,
	}
}

// MarkRpgRechargeManuallyFulfilled 管理端手工补钻后标记 fulfilled。
func (s *PayOrderService) MarkRpgRechargeManuallyFulfilled(ctx context.Context, outTradeNo string) (interface{}, error) {
	if s.recharge == nil {
		return nil, errcode.WithMessage(errcode.InternalError, "充值服务未就绪")
	}
	return s.recharge.MarkManualFulfillmentAndNotify(ctx, outTradeNo)
}

// RefundOrder 退款（支持部分退款）。
func (s *PayOrderService) RefundOrder(ctx context.Context, outTradeNo, refundAmount, refundReason string) (interface{}, error) {
	order, err := s.repo.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "订单不存在")
		}
		return nil, err
	}
	if order.Status != payconst.OrderStatusPaid {
		return nil, errcode.WithMessage(errcode.InvalidParam, "只有已支付的订单才能退款")
	}
	requestRefund, _ := parseFloat(refundAmount)
	alreadyRefunded := order.RefundAmount
	remaining := order.TotalAmount - alreadyRefunded
	if requestRefund <= 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "退款金额必须大于 0")
	}
	if requestRefund > remaining+0.001 {
		return nil, errcode.WithMessage(errcode.InvalidParam,
			fmt.Sprintf("退款金额不能超过剩余可退金额，当前最多可退 ¥%.2f", remaining))
	}

	outRequestNo := fmt.Sprintf("%s_R%d%d", outTradeNo, time.Now().UnixMilli(), rand.Intn(1000))
	refundResult, err := s.pay.TradeRefund(ctx, TradeRefundDTO{
		OutTradeNo:   outTradeNo,
		RefundAmount: refundAmount,
		RefundReason: defaultRefundReason(refundReason),
		OutRequestNo: outRequestNo,
	})
	if err != nil {
		return map[string]interface{}{
			"alipaySuccess": false,
			"localSuccess":  false,
			"message":       "退款失败，请稍后重试",
		}, nil
	}
	if fc, _ := refundResult["fundChange"].(string); fc != "Y" {
		return map[string]interface{}{
			"alipaySuccess": false,
			"alipayResult":  refundResult,
			"localSuccess":  false,
			"message":       "退款未成功，支付宝侧未确认资金变动",
		}, nil
	}

	order.RefundAmount = math.Round((alreadyRefunded+requestRefund)*100) / 100
	if order.RefundAmount >= order.TotalAmount {
		order.Status = payconst.OrderStatusRefunded
	}
	if _, err := s.repo.Save(ctx, order); err != nil {
		return map[string]interface{}{
			"alipaySuccess": true,
			"localSuccess":  false,
			"message":       "支付宝退款成功，但本地状态更新失败",
		}, nil
	}
	msg := fmt.Sprintf("部分退款成功，已退 ¥%.2f，剩余可退 ¥%.2f",
		order.RefundAmount, order.TotalAmount-order.RefundAmount)
	if order.Status == payconst.OrderStatusRefunded {
		msg = "全额退款成功"
	}
	return map[string]interface{}{
		"outTradeNo":        outTradeNo,
		"refundAmount":      requestRefund,
		"totalRefundAmount": order.RefundAmount,
		"remaining":         order.TotalAmount - order.RefundAmount,
		"message":           msg,
	}, nil
}

func defaultRefundReason(reason string) string {
	if reason == "" {
		return "用户申请退款"
	}
	return reason
}

// CloseOrder 关单。
func (s *PayOrderService) CloseOrder(ctx context.Context, outTradeNo string) (interface{}, error) {
	order, err := s.repo.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "订单不存在")
		}
		return nil, err
	}
	if order.Status != payconst.OrderStatusPending && order.Status != payconst.OrderStatusFailed {
		return nil, errcode.WithMessage(errcode.InvalidParam, "只有待支付或失败的订单才能关单")
	}
	if _, err := s.pay.TradeClose(ctx, TradeCloseDTO{OutTradeNo: outTradeNo}); err != nil {
		return map[string]interface{}{
			"alipaySuccess": false,
			"localSuccess":  false,
			"message":       "关单失败，支付宝侧未确认",
		}, nil
	}
	order.Status = payconst.OrderStatusClosed
	if _, err := s.repo.Save(ctx, order); err != nil {
		return map[string]interface{}{
			"alipaySuccess": true,
			"localSuccess":  false,
			"message":       "支付宝关单成功，但本地状态更新失败",
		}, nil
	}
	s.pay.StopPolling(outTradeNo)
	return map[string]interface{}{"outTradeNo": outTradeNo, "message": "关单成功"}, nil
}

// QueryAndUpdateOrder 主动查询并更新本地。
func (s *PayOrderService) QueryAndUpdateOrder(ctx context.Context, outTradeNo string) (interface{}, error) {
	order, err := s.repo.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.InvalidParam, "订单不存在")
		}
		return nil, err
	}
	queryResult, err := s.pay.TradeQuery(ctx, TradeQueryDTO{OutTradeNo: outTradeNo})
	if err != nil {
		return map[string]interface{}{
			"alipaySuccess": false,
			"localSuccess":  false,
			"message":       "查询支付宝订单状态失败",
		}, nil
	}
	tradeStatus, _ := queryResult["tradeStatus"].(string)
	newStatus := mapAlipayStatus(tradeStatus)
	if newStatus != "" && newStatus != order.Status {
		order.Status = newStatus
		if tn, ok := queryResult["tradeNo"].(string); ok && tn != "" {
			order.TradeNo = tn
		}
		if _, err := s.repo.Save(ctx, order); err != nil {
			return map[string]interface{}{
				"alipaySuccess": true,
				"localSuccess":  false,
				"message":       "支付宝查询成功，但本地状态更新失败",
			}, nil
		}
	}
	return map[string]interface{}{
		"outTradeNo": outTradeNo,
		"status":     order.Status,
		"tradeNo":    order.TradeNo,
		"message":    "查询成功",
	}, nil
}

// DeleteOrders 批量删除本地订单。
func (s *PayOrderService) DeleteOrders(ctx context.Context, ids []int) (interface{}, error) {
	if len(ids) == 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "请选择要删除的订单")
	}
	unique := dedupePositiveIDs(ids)
	for _, id := range unique {
		order, err := s.repo.FindByID(ctx, id)
		if err != nil {
			continue
		}
		if order.Status == payconst.OrderStatusPaid {
			return nil, errcode.WithMessage(errcode.InvalidParam, "已支付订单请先退款后再删除")
		}
		if order.Status == payconst.OrderStatusPending {
			s.pay.StopPolling(order.OutTradeNo)
		}
	}
	deleted, err := s.repo.DeleteByIDs(ctx, unique)
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("成功删除 %d 条订单", deleted)
	if deleted == 0 {
		msg = "未删除任何订单"
	}
	return map[string]interface{}{"deleted": deleted, "message": msg}, nil
}

func (s *PayOrderService) generateTradeNo() string {
	now := time.Now()
	return fmt.Sprintf("%04d%02d%02d%02d%02d%02d%04d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		rand.Intn(10000))
}

func dedupePositiveIDs(ids []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func toIntFromAny(v interface{}) interface{} {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return nil
	}
}

func toBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	default:
		return false
	}
}
