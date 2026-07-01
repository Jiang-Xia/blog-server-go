// Package recharge RPG 钻石充值（pay_order 联动）。
package recharge

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	rpgnotify "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/notify"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

const bizTypeRPGRecharge = "rpg_recharge"
const rechargeSubject = "博客钻石充值"

// Service 充值业务。
type Service struct {
	repo      *rpgrepo.RpgRepo
	core      *rpgcore.RpgService
	inventory *inventory.Service
	notify    *rpgnotify.RpgNotifyService
}

// NewService 构造充值 Service。
func NewService(
	repo *rpgrepo.RpgRepo,
	core *rpgcore.RpgService,
	inventory *inventory.Service,
	notify *rpgnotify.RpgNotifyService,
) *Service {
	return &Service{repo: repo, core: core, inventory: inventory, notify: notify}
}

// CreateRecharge 创建充值意向单（返回 outTradeNo 与应付信息）。
func (s *Service) CreateRecharge(ctx context.Context, uid int, amountYuan float64) (map[string]interface{}, error) {
	if err := validateRechargeYuan(amountYuan); err != nil {
		return nil, err
	}
	amountYuan = normalizeRechargeYuan(amountYuan)
	if _, err := s.core.GetOrCreateRpg(ctx, uid); err != nil {
		return nil, err
	}
	diamonds := calcRechargeDiamonds(amountYuan)
	outTradeNo := generateOutTradeNo(uid)
	order, err := s.repo.CreatePayOrder(ctx, &ent.PayOrder{
		OutTradeNo:  outTradeNo,
		Subject:     rechargeSubject,
		TotalAmount: amountYuan,
		Status:      "PENDING",
		Channel:     "alipay",
		ExtendParams: map[string]interface{}{
			"bizType":    bizTypeRPGRecharge,
			"uid":        uid,
			"diamonds":   diamonds,
			"amountYuan": amountYuan,
			"fulfilled":  false,
		},
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"outTradeNo":  order.OutTradeNo,
		"amountYuan":  amountYuan,
		"diamonds":    diamonds,
		"subject":     order.Subject,
		"status":      order.Status,
		"payUrl":      fmt.Sprintf("/api/v1/pay/alipay?outTradeNo=%s", outTradeNo),
	}, nil
}

// GetStatus 查询充值订单状态并尝试履约发钻。
func (s *Service) GetStatus(ctx context.Context, uid int, outTradeNo string) (map[string]interface{}, error) {
	if outTradeNo == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "商户订单号不能为空")
	}
	order, err := s.repo.FindPayOrderByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return nil, errcode.WithMessage(errcode.NotFound, "充值订单不存在")
	}
	if err := assertOrderOwner(order, uid); err != nil {
		return nil, err
	}
	fulfill, _ := s.TryFulfillOrderRecord(ctx, order)
	if fulfill != nil {
		order = fulfill
	}
	balance, _ := s.inventory.GetCurrency(ctx, uid)
	return map[string]interface{}{
		"outTradeNo": outTradeNo,
		"status":     order.Status,
		"amountYuan": order.TotalAmount,
		"diamonds":   extendInt(order.ExtendParams, "diamonds"),
		"fulfilled":  extendBool(order.ExtendParams, "fulfilled"),
		"balance":    balance,
	}, nil
}

// TryFulfillOrderRecord 支付成功后幂等发钻。
func (s *Service) TryFulfillOrderRecord(ctx context.Context, order *ent.PayOrder) (*ent.PayOrder, error) {
	if order == nil || order.Status != "PAID" {
		return order, nil
	}
	if extendBool(order.ExtendParams, "fulfilled") {
		// 幂等：extendParams.fulfilled 已标记则跳过发钻。
		return order, nil
	}
	uid := extendInt(order.ExtendParams, "uid")
	diamonds := extendInt(order.ExtendParams, "diamonds")
	if uid <= 0 || diamonds <= 0 {
		return order, nil
	}
	balance, err := s.inventory.AdjustCurrency(ctx, uid, diamonds, "recharge")
	if err != nil {
		return order, err
	}
	if order.ExtendParams == nil {
		order.ExtendParams = map[string]interface{}{}
	}
	order.ExtendParams["fulfilled"] = true
	saved, err := s.repo.UpdatePayOrder(ctx, order)
	if err != nil {
		return order, err
	}
	if s.notify != nil {
		_ = s.notify.NotifyRechargeComplete(ctx, uid, rpgnotify.RechargeCompletePayload{
			OutTradeNo: order.OutTradeNo,
			Diamonds:   diamonds,
			Balance:    balance,
			AmountYuan: order.TotalAmount,
		})
	}
	return saved, nil
}

// MarkManualFulfillmentAndNotify 管理端手工标记充值已履约并发 WS 通知。
func (s *Service) MarkManualFulfillmentAndNotify(ctx context.Context, outTradeNo string) (interface{}, error) {
	order, err := s.repo.FindPayOrderByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return nil, errcode.WithMessage(errcode.NotFound, "充值订单不存在")
	}
	fulfilled, err := s.TryFulfillOrderRecord(ctx, order)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"outTradeNo": outTradeNo,
		"fulfilled":  extendBool(fulfilled.ExtendParams, "fulfilled"),
		"status":     fulfilled.Status,
	}, nil
}

func validateRechargeYuan(y float64) error {
	if y < constants.Economy.RechargeMinYuan || y > constants.Economy.RechargeMaxYuan {
		return errcode.WithMessage(errcode.InvalidParam, "充值金额超出允许范围")
	}
	return nil
}

func normalizeRechargeYuan(y float64) float64 {
	return math.Round(y*100) / 100
}

func calcRechargeDiamonds(yuan float64) int {
	return int(math.Floor(yuan * float64(constants.Economy.RechargeRate)))
}

func generateOutTradeNo(uid int) string {
	return fmt.Sprintf("RPG%d%d", uid, time.Now().UnixNano())
}

func assertOrderOwner(order *ent.PayOrder, uid int) error {
	if extendInt(order.ExtendParams, "uid") != uid {
		return errcode.WithMessage(errcode.Forbidden, "无权查看该订单")
	}
	return nil
}

func extendInt(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func extendBool(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
