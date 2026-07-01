// rpg_recharge_link 博客 RPG 充值单与支付回调关联解析。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"

	payconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/constants"
	payrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
)

// RpgRechargeLinkMode RPG 充值单关联方式。
type RpgRechargeLinkMode string

// RPG 充值单关联模式常量。
const (
	// RpgLinkExplicitOutTradeNo 通过商户单号 out_trade_no 显式关联待支付单。
	RpgLinkExplicitOutTradeNo RpgRechargeLinkMode = "explicit_out_trade_no"
	// RpgLinkNone 未关联到 RPG 充值意向单。
	RpgLinkNone RpgRechargeLinkMode = "none"
)

// RpgRechargeLinkResult 博客 RPG 充值单关联结果。
type RpgRechargeLinkResult struct {
	OutTradeNo string
	Mode       RpgRechargeLinkMode
}

// ResolveRpgRechargeOutTradeNo 博客 RPG 充值：仅通过商户单号 out_trade_no 关联。
func ResolveRpgRechargeOutTradeNo(ctx context.Context, repo *payrepo.PayOrderRepo, outTradeNo string) (RpgRechargeLinkResult, error) {
	outTradeNo = strings.TrimSpace(outTradeNo)
	if outTradeNo == "" {
		return RpgRechargeLinkResult{Mode: RpgLinkNone}, nil
	}
	order, err := repo.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		if ent.IsNotFound(err) {
			return RpgRechargeLinkResult{Mode: RpgLinkNone}, nil
		}
		return RpgRechargeLinkResult{}, err
	}
	if bizTypeOf(order) == payconst.PAY_BIZ_RPG_RECHARGE && order.Status == payconst.OrderStatusPending {
		return RpgRechargeLinkResult{OutTradeNo: outTradeNo, Mode: RpgLinkExplicitOutTradeNo}, nil
	}
	return RpgRechargeLinkResult{Mode: RpgLinkNone}, nil
}

// FindRpgRechargeOrderForNotify notify 按 out_trade_no 查博客充值意向单。
func FindRpgRechargeOrderForNotify(ctx context.Context, repo *payrepo.PayOrderRepo, postData map[string]string) (*ent.PayOrder, RpgRechargeLinkMode, error) {
	outTradeNo := strings.TrimSpace(postData["out_trade_no"])
	if outTradeNo == "" {
		return nil, RpgLinkNone, nil
	}
	order, err := repo.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, RpgLinkNone, nil
		}
		return nil, RpgLinkNone, err
	}
	if bizTypeOf(order) == payconst.PAY_BIZ_RPG_RECHARGE {
		return order, RpgLinkExplicitOutTradeNo, nil
	}
	return nil, RpgLinkNone, nil
}

func bizTypeOf(order *ent.PayOrder) string {
	if order == nil || order.ExtendParams == nil {
		return ""
	}
	v, _ := order.ExtendParams["bizType"].(string)
	return v
}

func parsePassbackParams(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		decoded = raw
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(decoded), &m); err != nil {
		return nil
	}
	return m
}

func amountYuanOf(order *ent.PayOrder) float64 {
	if order == nil {
		return 0
	}
	if order.ExtendParams != nil {
		switch v := order.ExtendParams["amountYuan"].(type) {
		case float64:
			return v
		case json.Number:
			f, _ := v.Float64()
			return f
		}
	}
	return order.TotalAmount
}

func sameRechargeYuan(a, b float64) bool {
	return math.Abs(a-b) <= 0.001
}

func fmtAmount(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
