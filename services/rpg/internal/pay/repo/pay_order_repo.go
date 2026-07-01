// Package repo 支付订单 Ent 数据访问。
package repo

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent/payorder"
	payconst "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/pay/constants"
)

// PayOrderRepo pay_order 表 CRUD。
type PayOrderRepo struct {
	client *ent.Client
}

// NewPayOrderRepo 构造 PayOrderRepo。
func NewPayOrderRepo(client *ent.Client) *PayOrderRepo {
	return &PayOrderRepo{client: client}
}

// PayOrderListFilter 管理端订单列表筛选。
type PayOrderListFilter struct {
	Page        int
	PageSize    int
	OutTradeNo  string
	Status      string
	Channel     string
	Subject     string
	BizType     string
	RechargeUID string
}

// CreateInput 创建订单入参。
type CreateInput struct {
	OutTradeNo   string
	TradeNo      string
	Subject      string
	TotalAmount  float64
	BuyerOpenID  string
	Status       string
	Channel      string
	ExtendParams map[string]interface{}
}

// FindByOutTradeNo 按商户订单号查询。
func (r *PayOrderRepo) FindByOutTradeNo(ctx context.Context, outTradeNo string) (*ent.PayOrder, error) {
	return r.client.PayOrder.Query().
		Where(payorder.OutTradeNoEQ(outTradeNo)).
		Only(ctx)
}

// FindByID 按主键查询。
func (r *PayOrderRepo) FindByID(ctx context.Context, id int) (*ent.PayOrder, error) {
	return r.client.PayOrder.Query().
		Where(payorder.IDEQ(id)).
		Only(ctx)
}

// Create 新建订单。
func (r *PayOrderRepo) Create(ctx context.Context, in CreateInput) (*ent.PayOrder, error) {
	b := r.client.PayOrder.Create().
		SetOutTradeNo(in.OutTradeNo).
		SetTradeNo(in.TradeNo).
		SetSubject(in.Subject).
		SetTotalAmount(in.TotalAmount).
		SetBuyerOpenId(in.BuyerOpenID).
		SetStatus(defaultStr(in.Status, payconst.OrderStatusPending)).
		SetChannel(defaultStr(in.Channel, payconst.ChannelAlipay))
	if in.ExtendParams != nil {
		b.SetExtendParams(in.ExtendParams)
	}
	return b.Save(ctx)
}

// Save 更新订单实体。
func (r *PayOrderRepo) Save(ctx context.Context, order *ent.PayOrder) (*ent.PayOrder, error) {
	upd := r.client.PayOrder.UpdateOneID(order.ID).
		SetTradeNo(order.TradeNo).
		SetSubject(order.Subject).
		SetTotalAmount(order.TotalAmount).
		SetBuyerOpenId(order.BuyerOpenId).
		SetStatus(order.Status).
		SetRefundAmount(order.RefundAmount).
		SetChannel(order.Channel)
	if order.ExtendParams != nil {
		upd.SetExtendParams(order.ExtendParams)
	}
	return upd.Save(ctx)
}

// List 分页查询订单。
func (r *PayOrderRepo) List(ctx context.Context, f PayOrderListFilter) ([]*ent.PayOrder, int, error) {
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := r.client.PayOrder.Query()
	if f.OutTradeNo != "" {
		q = q.Where(payorder.OutTradeNoEQ(f.OutTradeNo))
	}
	if f.Status != "" {
		q = q.Where(payorder.StatusEQ(f.Status))
	}
	if f.Channel != "" {
		q = q.Where(payorder.ChannelEQ(f.Channel))
	}
	if f.Subject != "" {
		q = q.Where(payorder.SubjectContains(f.Subject))
	}
	if f.BizType != "" {
		q = q.Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"JSON_UNQUOTE(JSON_EXTRACT("+s.C(payorder.FieldExtendParams)+", '$.bizType')) = ?",
				f.BizType,
			))
		})
	}
	if f.RechargeUID != "" {
		q = q.Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"JSON_UNQUOTE(JSON_EXTRACT("+s.C(payorder.FieldExtendParams)+", '$.uid')) = ?",
				f.RechargeUID,
			))
		})
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(payorder.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	return rows, total, err
}

// DeleteByIDs 批量删除本地订单记录。
func (r *PayOrderRepo) DeleteByIDs(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return r.client.PayOrder.Delete().Where(payorder.IDIn(ids...)).Exec(ctx)
}

// ListPendingByBizType 查询指定业务类型 pending 订单。
func (r *PayOrderRepo) ListPendingByBizTypeAndUID(ctx context.Context, bizType string, uid int) ([]*ent.PayOrder, error) {
	uidStr := fmt.Sprintf("%d", uid)
	return r.client.PayOrder.Query().
		Where(
			payorder.StatusEQ(payconst.OrderStatusPending),
			func(s *sql.Selector) {
				s.Where(sql.ExprP(
					"JSON_UNQUOTE(JSON_EXTRACT("+s.C(payorder.FieldExtendParams)+", '$.bizType')) = ?",
					bizType,
				))
				s.Where(sql.ExprP(
					"JSON_UNQUOTE(JSON_EXTRACT("+s.C(payorder.FieldExtendParams)+", '$.uid')) = ?",
					uidStr,
				))
			},
		).
		All(ctx)
}

func defaultStr(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
