// Package service 支付宝 SDK 封装与 C 端支付流程。
package service

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	payconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/constants"
	payrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/pay/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
)

// TradeCreateDTO C 端/管理端创建交易参数（snake_case 对齐 Nest）。
type TradeCreateDTO struct {
	OutTradeNo   string                 `json:"out_trade_no"`
	Subject      string                 `json:"subject"`
	TotalAmount  string                 `json:"total_amount"`
	BuyerOpenID  string                 `json:"buyer_open_id"`
	ExtendParams map[string]interface{} `json:"extend_params"`
}

// TradeQueryDTO 交易查询参数。
type TradeQueryDTO struct {
	OutTradeNo string `json:"out_trade_no"`
}

// TradeRefundDTO 退款参数。
type TradeRefundDTO struct {
	OutTradeNo     string `json:"out_trade_no"`
	RefundAmount   string `json:"refund_amount"`
	RefundReason   string `json:"refund_reason"`
	OutRequestNo   string `json:"out_request_no"`
}

// TradeCloseDTO 关单参数。
type TradeCloseDTO struct {
	OutTradeNo string `json:"out_trade_no"`
}

// GetOpenIDDTO 小程序 code 换 openid。
type GetOpenIDDTO struct {
	Code string `json:"code"`
}

// H5OpenMiniDTO H5 拉起小程序链接。
type H5OpenMiniDTO struct {
	Type    string            `json:"type"`
	AppID   string            `json:"appId"`
	Page    string            `json:"page"`
	Query   map[string]string `json:"query"`
	Version string            `json:"version"`
}

// PayService 支付宝交易与本地订单同步。
type PayService struct {
	cfg    *config.Config
	repo   *payrepo.PayOrderRepo
	client *alipay.Client
	log    *zap.Logger

	pollMu     sync.Mutex
	pollCancel map[string]context.CancelFunc
}

// NewPayService 构造 PayService；未配置密钥时 client 为 nil（本地开发可跳过真实调用）。
func NewPayService(cfg *config.Config, repo *payrepo.PayOrderRepo, log *zap.Logger) (*PayService, error) {
	s := &PayService{
		cfg:        cfg,
		repo:       repo,
		log:        log,
		pollCancel: make(map[string]context.CancelFunc),
	}
	pc := cfg.Pay
	if strings.TrimSpace(pc.AlipayAppID) == "" || strings.TrimSpace(pc.AlipayPrivateKey) == "" {
		log.Warn("pay alipay not configured, SDK disabled")
		return s, nil
	}
	isProd := !pc.Sandbox
	var opts []alipay.OptionFunc
	if pc.Sandbox && pc.UseLegacySandboxGateway {
		opts = append(opts, alipay.WithPastSandboxGateway())
	}
	client, err := alipay.New(pc.AlipayAppID, pc.AlipayPrivateKey, isProd, opts...)
	if err != nil {
		return nil, fmt.Errorf("alipay.New: %w", err)
	}
	if pk := strings.TrimSpace(pc.AlipayPublicKey); pk != "" {
		if err := client.LoadAliPayPublicKey(pk); err != nil {
			return nil, fmt.Errorf("LoadAliPayPublicKey: %w", err)
		}
	}
	s.client = client
	return s, nil
}

func (s *PayService) notifyURL() string {
	return strings.TrimSpace(s.cfg.Pay.AlipayNotifyURL)
}

// CallAlipayTradeCreate 仅调用支付宝创建交易（不入库、不轮询）。
func (s *PayService) CallAlipayTradeCreate(ctx context.Context, dto TradeCreateDTO) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("alipay client not configured")
	}
	p := alipay.TradeCreate{
		Trade: alipay.Trade{
			OutTradeNo:  dto.OutTradeNo,
			Subject:     dto.Subject,
			TotalAmount: dto.TotalAmount,
			NotifyURL:   s.notifyURL(),
		},
		BuyerOpenId: dto.BuyerOpenID,
	}
	rsp, err := s.client.TradeCreate(ctx, p)
	if err != nil {
		return nil, err
	}
	return tradeCreateRspToMap(rsp), nil
}

// StartPolling 启动轮询（每 10 秒，最多 18 次）。
func (s *PayService) StartPolling(outTradeNo string) {
	if s.client == nil || outTradeNo == "" {
		return
	}
	s.StopPolling(outTradeNo)
	ctx, cancel := context.WithCancel(context.Background())
	s.pollMu.Lock()
	s.pollCancel[outTradeNo] = cancel
	s.pollMu.Unlock()

	go func() {
		defer s.StopPolling(outTradeNo)
		for i := 0; i < 18; i++ {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
			order, err := s.repo.FindByOutTradeNo(ctx, outTradeNo)
			if err != nil || order.Status != payconst.OrderStatusPending {
				return
			}
			rsp, err := s.client.TradeQuery(ctx, alipay.TradeQuery{OutTradeNo: outTradeNo})
			if err != nil {
				s.log.Warn("poll trade query failed", zap.String("outTradeNo", outTradeNo), zap.Error(err))
				continue
			}
			changed, err := s.syncOrderFromAlipayStatus(ctx, order, string(rsp.TradeStatus), rsp.TradeNo)
			if err != nil {
				s.log.Warn("poll sync order failed", zap.Error(err))
				continue
			}
			if changed {
				return
			}
		}
		s.log.Info("order poll max reached", zap.String("outTradeNo", outTradeNo))
	}()
}

// StopPolling 停止指定订单轮询。
func (s *PayService) StopPolling(outTradeNo string) {
	s.pollMu.Lock()
	defer s.pollMu.Unlock()
	if cancel, ok := s.pollCancel[outTradeNo]; ok {
		cancel()
		delete(s.pollCancel, outTradeNo)
	}
}

func mapAlipayStatus(tradeStatus string) string {
	switch tradeStatus {
	case "WAIT_BUYER_PAY":
		return payconst.OrderStatusPending
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return payconst.OrderStatusPaid
	case "TRADE_CLOSED":
		return payconst.OrderStatusClosed
	default:
		return ""
	}
}

func (s *PayService) syncOrderFromAlipayStatus(ctx context.Context, order *ent.PayOrder, tradeStatus, tradeNo string) (bool, error) {
	newStatus := mapAlipayStatus(tradeStatus)
	if newStatus == "" || newStatus == order.Status {
		return false, nil
	}
	order.Status = newStatus
	if tradeNo != "" {
		order.TradeNo = tradeNo
	}
	if _, err := s.repo.Save(ctx, order); err != nil {
		return false, err
	}
	s.log.Info("order status synced", zap.String("outTradeNo", order.OutTradeNo), zap.String("status", newStatus))
	if newStatus == payconst.OrderStatusPaid {
		s.StopPolling(order.OutTradeNo)
		InvokePayPaidCallbacks(ctx, order)
	}
	if newStatus == payconst.OrderStatusClosed {
		s.StopPolling(order.OutTradeNo)
	}
	return true, nil
}

// ClosePendingOrder 关闭未支付订单。
func (s *PayService) ClosePendingOrder(ctx context.Context, outTradeNo string) error {
	order, err := s.repo.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}
	if order.Status != payconst.OrderStatusPending {
		return nil
	}
	if order.TradeNo != "" {
		_, err := s.CCloseOrder(ctx, TradeCloseDTO{OutTradeNo: outTradeNo})
		return err
	}
	order.Status = payconst.OrderStatusClosed
	_, err = s.repo.Save(ctx, order)
	s.StopPolling(outTradeNo)
	return err
}

// HandleAlipayNotify 支付宝异步通知验签入库。
func (s *PayService) HandleAlipayNotify(ctx context.Context, postData map[string]string) (bool, error) {
	if s.client == nil || len(postData) == 0 {
		return false, nil
	}
	form := urlValuesFromMap(postData)
	noti, err := s.client.DecodeNotification(ctx, form)
	if err != nil {
		s.log.Warn("alipay notify verify failed", zap.Error(err))
		return false, nil
	}

	notifyOutTradeNo := noti.OutTradeNo
	if notifyOutTradeNo == "" {
		return false, nil
	}

	order, err := s.repo.FindByOutTradeNo(ctx, notifyOutTradeNo)
	directIsRpg := err == nil && bizTypeOf(order) == payconst.PAY_BIZ_RPG_RECHARGE

	if err != nil || !directIsRpg {
		if ent.IsNotFound(err) {
			order = nil
		}
		linked, mode, linkErr := FindRpgRechargeOrderForNotify(ctx, s.repo, postData)
		if linkErr != nil {
			return false, linkErr
		}
		if linked != nil {
			s.log.Info("alipay notify rpg linked",
				zap.String("mode", string(mode)),
				zap.String("notifyNo", notifyOutTradeNo),
				zap.String("localNo", linked.OutTradeNo))
			order = linked
		} else if order == nil {
			s.log.Warn("alipay notify order not found", zap.String("outTradeNo", notifyOutTradeNo))
			return true, nil
		}
	}

	if noti.TotalAmount != "" {
		notifyAmount, _ := parseFloat(noti.TotalAmount)
		orderAmount := order.TotalAmount
		if math.Abs(notifyAmount-orderAmount) > 0.01 {
			s.log.Warn("alipay notify amount mismatch",
				zap.String("outTradeNo", order.OutTradeNo),
				zap.Float64("notify", notifyAmount),
				zap.Float64("order", orderAmount))
			return false, nil
		}
	}

	_, err = s.syncOrderFromAlipayStatus(ctx, order, string(noti.TradeStatus), noti.TradeNo)
	return err == nil, err
}

// TradeQuery 管理端透传查询。
func (s *PayService) TradeQuery(ctx context.Context, dto TradeQueryDTO) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("alipay client not configured")
	}
	rsp, err := s.client.TradeQuery(ctx, alipay.TradeQuery{OutTradeNo: dto.OutTradeNo})
	if err != nil {
		return nil, err
	}
	return tradeQueryRspToMap(rsp), nil
}

// TradeRefund 管理端透传退款。
func (s *PayService) TradeRefund(ctx context.Context, dto TradeRefundDTO) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("alipay client not configured")
	}
	rsp, err := s.client.TradeRefund(ctx, alipay.TradeRefund{
		OutTradeNo:   dto.OutTradeNo,
		RefundAmount: dto.RefundAmount,
		RefundReason: dto.RefundReason,
		OutRequestNo: dto.OutRequestNo,
	})
	if err != nil {
		return nil, err
	}
	return tradeRefundRspToMap(rsp), nil
}

// TradeClose 管理端透传关单。
func (s *PayService) TradeClose(ctx context.Context, dto TradeCloseDTO) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("alipay client not configured")
	}
	rsp, err := s.client.TradeClose(ctx, alipay.TradeClose{OutTradeNo: dto.OutTradeNo})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"outTradeNo": rsp.OutTradeNo}, nil
}

// CCreateOrder C 端创建订单：支付宝 + 入库 + 轮询。
func (s *PayService) CCreateOrder(ctx context.Context, dto TradeCreateDTO) (map[string]interface{}, error) {
	link, _ := ResolveRpgRechargeOutTradeNo(ctx, s.repo, dto.OutTradeNo)
	if link.OutTradeNo != "" && dto.OutTradeNo == "" {
		dto.OutTradeNo = link.OutTradeNo
		s.log.Info("c create rpg link", zap.String("mode", string(link.Mode)), zap.String("outTradeNo", link.OutTradeNo))
	}

	outTradeNo := dto.OutTradeNo
	result := map[string]interface{}{"alipaySuccess": false, "localSuccess": false}

	if outTradeNo != "" {
		existing, err := s.repo.FindByOutTradeNo(ctx, outTradeNo)
		if err == nil && existing.Status != payconst.OrderStatusPending {
			return map[string]interface{}{
				"alipaySuccess": false,
				"localSuccess":  false,
				"message":       "订单不可支付",
			}, nil
		}
		if err == nil {
			if msg := s.validateRpgRechargeCreate(existing, &dto); msg != "" {
				return map[string]interface{}{
					"alipaySuccess": false,
					"localSuccess":  false,
					"message":       msg,
				}, nil
			}
		}
	}

	if s.client == nil {
		return result, nil
	}

	p := alipay.TradeCreate{
		Trade: alipay.Trade{
			OutTradeNo:  outTradeNo,
			Subject:     dto.Subject,
			TotalAmount: dto.TotalAmount,
			NotifyURL:   s.notifyURL(),
		},
		BuyerOpenId: dto.BuyerOpenID,
	}
	if outTradeNo != "" {
		if existing, err := s.repo.FindByOutTradeNo(ctx, outTradeNo); err == nil && bizTypeOf(existing) == payconst.PAY_BIZ_RPG_RECHARGE {
			p.PassbackParams = fmt.Sprintf(`{"biz":"%s","outTradeNo":"%s"}`, payconst.PAY_BIZ_RPG_RECHARGE, outTradeNo)
			if dto.Subject == "" {
				p.Subject = payconst.RPGRechargeSubject
			}
		}
	}

	rsp, err := s.client.TradeCreate(ctx, p)
	if err != nil {
		s.log.Error("c create alipay failed", zap.String("outTradeNo", outTradeNo), zap.Error(err))
		return result, nil
	}
	for k, v := range tradeCreateRspToMap(rsp) {
		result[k] = v
	}
	result["alipaySuccess"] = true

	if outTradeNo != "" {
		localOK := s.upsertLocalOrder(ctx, outTradeNo, dto, rsp.TradeNo)
		result["localSuccess"] = localOK
		if localOK {
			s.StartPolling(outTradeNo)
		}
	}
	return result, nil
}

func (s *PayService) validateRpgRechargeCreate(existing *ent.PayOrder, dto *TradeCreateDTO) string {
	if bizTypeOf(existing) != payconst.PAY_BIZ_RPG_RECHARGE {
		return ""
	}
	expected := amountYuanOf(existing)
	if dto.TotalAmount != "" {
		req, _ := parseFloat(dto.TotalAmount)
		if !sameRechargeYuan(req, expected) {
			return fmt.Sprintf("充值金额须为 %s 元", fmtAmount(expected))
		}
	} else {
		dto.TotalAmount = fmtAmount(expected)
	}
	if dto.Subject == "" {
		dto.Subject = payconst.RPGRechargeSubject
	}
	return ""
}

func (s *PayService) upsertLocalOrder(ctx context.Context, outTradeNo string, dto TradeCreateDTO, tradeNo string) bool {
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
		return createErr == nil
	}
	if err != nil {
		return false
	}
	existing.TradeNo = tradeNo
	if dto.BuyerOpenID != "" {
		existing.BuyerOpenId = dto.BuyerOpenID
	}
	if dto.Subject != "" && existing.Subject == "" {
		existing.Subject = dto.Subject
	}
	if dto.TotalAmount != "" && existing.TotalAmount == 0 {
		existing.TotalAmount, _ = parseFloat(dto.TotalAmount)
	}
	if bizTypeOf(existing) == payconst.PAY_BIZ_RPG_RECHARGE && dto.BuyerOpenID != "" {
		if existing.ExtendParams == nil {
			existing.ExtendParams = map[string]interface{}{}
		}
		existing.ExtendParams["buyerOpenId"] = dto.BuyerOpenID
	}
	_, saveErr := s.repo.Save(ctx, existing)
	return saveErr == nil
}

// CQueryOrder C 端查询并同步本地。
func (s *PayService) CQueryOrder(ctx context.Context, dto TradeQueryDTO) (map[string]interface{}, error) {
	result := map[string]interface{}{"alipaySuccess": false, "localSuccess": false}
	if s.client == nil {
		return result, nil
	}
	rsp, err := s.client.TradeQuery(ctx, alipay.TradeQuery{OutTradeNo: dto.OutTradeNo})
	if err != nil {
		s.log.Error("c query alipay failed", zap.Error(err))
		return result, nil
	}
	for k, v := range tradeQueryRspToMap(rsp) {
		result[k] = v
	}
	result["alipaySuccess"] = true
	if dto.OutTradeNo != "" {
		if order, err := s.repo.FindByOutTradeNo(ctx, dto.OutTradeNo); err == nil {
			_, _ = s.syncOrderFromAlipayStatus(ctx, order, string(rsp.TradeStatus), rsp.TradeNo)
			result["localSuccess"] = true
		}
	}
	return result, nil
}

// CRefundOrder C 端退款。
func (s *PayService) CRefundOrder(ctx context.Context, dto TradeRefundDTO) (map[string]interface{}, error) {
	result := map[string]interface{}{"alipaySuccess": false, "localSuccess": false}
	if s.client == nil {
		return result, nil
	}
	rsp, err := s.client.TradeRefund(ctx, alipay.TradeRefund{
		OutTradeNo:   dto.OutTradeNo,
		RefundAmount: dto.RefundAmount,
		RefundReason: dto.RefundReason,
		OutRequestNo: dto.OutRequestNo,
	})
	if err != nil {
		return result, nil
	}
	for k, v := range tradeRefundRspToMap(rsp) {
		result[k] = v
	}
	if rsp.FundChange != "Y" {
		return result, nil
	}
	result["alipaySuccess"] = true
	if order, err := s.repo.FindByOutTradeNo(ctx, dto.OutTradeNo); err == nil && order.Status == payconst.OrderStatusPaid {
		refund, _ := parseFloat(dto.RefundAmount)
		order.Status = payconst.OrderStatusRefunded
		order.RefundAmount = refund
		_, _ = s.repo.Save(ctx, order)
		result["localSuccess"] = true
	}
	return result, nil
}

// CCloseOrder C 端关单。
func (s *PayService) CCloseOrder(ctx context.Context, dto TradeCloseDTO) (map[string]interface{}, error) {
	result := map[string]interface{}{"alipaySuccess": false, "localSuccess": false}
	if s.client == nil {
		return result, nil
	}
	_, err := s.client.TradeClose(ctx, alipay.TradeClose{OutTradeNo: dto.OutTradeNo})
	if err != nil {
		return result, nil
	}
	result["alipaySuccess"] = true
	if order, err := s.repo.FindByOutTradeNo(ctx, dto.OutTradeNo); err == nil {
		if order.Status == payconst.OrderStatusPending || order.Status == payconst.OrderStatusFailed {
			order.Status = payconst.OrderStatusClosed
			_, _ = s.repo.Save(ctx, order)
			s.StopPolling(dto.OutTradeNo)
		}
		result["localSuccess"] = true
	}
	return result, nil
}

// GetOpenIDByCode 支付宝小程序 code 换 openid。
func (s *PayService) GetOpenIDByCode(ctx context.Context, dto GetOpenIDDTO) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("alipay client not configured")
	}
	rsp, err := s.client.SystemOauthToken(ctx, alipay.SystemOauthToken{
		GrantType: "authorization_code",
		Code:      dto.Code,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"userId":  rsp.UserId,
		"openId":  rsp.OpenId,
		"accessToken": rsp.AccessToken,
	}, nil
}

// BuildH5OpenMiniURL 生成 H5 拉起小程序链接（支付宝；微信暂返回错误）。
func (s *PayService) BuildH5OpenMiniURL(ctx context.Context, dto H5OpenMiniDTO) (map[string]string, error) {
	if dto.Type == "wechat" {
		return nil, fmt.Errorf("wechat h5 open mini not implemented")
	}
	appID := dto.AppID
	if appID == "" {
		appID = s.cfg.Pay.AlipayAppID
	}
	page := strings.TrimPrefix(dto.Page, "/")
	queryString := ""
	if len(dto.Query) > 0 {
		vals := url.Values{}
		for k, v := range dto.Query {
			vals.Set(k, v)
		}
		queryString = vals.Encode()
	}
	pageParam := page
	if queryString != "" {
		pageParam = page + "?" + queryString
	}
	scheme := fmt.Sprintf("alipays://platformapi/startapp?appId=%s", url.QueryEscape(appID))
	if pageParam != "" {
		scheme += "&page=" + url.QueryEscape(pageParam)
	}
	universal := "https://render.alipay.com/p/s/i?scheme=" + url.QueryEscape(scheme)
	return map[string]string{"scheme": scheme, "universalLink": universal}, nil
}

func urlValuesFromMap(m map[string]string) url.Values {
	v := url.Values{}
	for k, val := range m {
		v.Set(k, val)
	}
	return v
}

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func tradeCreateRspToMap(rsp *alipay.TradeCreateRsp) map[string]interface{} {
	if rsp == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"tradeNo": rsp.TradeNo,
		"outTradeNo": rsp.OutTradeNo,
	}
}

func tradeQueryRspToMap(rsp *alipay.TradeQueryRsp) map[string]interface{} {
	if rsp == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"tradeNo":     rsp.TradeNo,
		"outTradeNo":  rsp.OutTradeNo,
		"tradeStatus": rsp.TradeStatus,
		"totalAmount": rsp.TotalAmount,
	}
}

func tradeRefundRspToMap(rsp *alipay.TradeRefundRsp) map[string]interface{} {
	if rsp == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"fundChange": rsp.FundChange,
		"tradeNo":    rsp.TradeNo,
		"outTradeNo": rsp.OutTradeNo,
	}
}
