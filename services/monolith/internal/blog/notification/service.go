// Package notification 站内通知 CRUD 与分页查询；WS 推送在 Plan 08。
package notification

import (
	"context"
	"encoding/json"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/sitenotification"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

// Service 站内通知业务逻辑。
type Service struct {
	client *ent.Client
}

// NewService 构造通知服务。
func NewService(client *ent.Client) *Service {
	return &Service{client: client}
}

// NotificationItem 下发给前端的通知项（payload 已反序列化）。
type NotificationItem struct {
	ID         int                    `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	Read       int                    `json:"read"`
	CreateTime interface{}            `json:"createTime"`
}

// Create 写入一条站内通知（内部调用；Plan 08 完善 WS 推送）。
func (s *Service) Create(ctx context.Context, uid int, typ string, payload map[string]interface{}) (*ent.SiteNotification, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte("{}")
	}
	return s.client.SiteNotification.Create().
		SetUID(uid).
		SetType(typ).
		SetPayload(string(raw)).
		SetRead(0).
		Save(ctx)
}

// ListByUID 分页查询用户通知。
func (s *Service) ListByUID(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	query := s.client.SiteNotification.Query().
		Where(sitenotification.UIDEQ(uid))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := query.
		Order(ent.Desc(sitenotification.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]NotificationItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toNotificationItem(row))
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": repo.CalcNestPagination(total, pageSize, page),
	}, nil
}

// CountUnread 未读通知数量。
func (s *Service) CountUnread(ctx context.Context, uid int) (int, error) {
	return s.client.SiteNotification.Query().
		Where(
			sitenotification.UIDEQ(uid),
			sitenotification.ReadEQ(0),
		).
		Count(ctx)
}

// MarkRead 标记已读：传 id 则单条，否则全部未读。
func (s *Service) MarkRead(ctx context.Context, uid int, id *int) error {
	if id != nil {
		_, err := s.client.SiteNotification.Update().
			Where(
				sitenotification.IDEQ(*id),
				sitenotification.UIDEQ(uid),
			).
			SetRead(1).
			Save(ctx)
		return err
	}
	_, err := s.client.SiteNotification.Update().
		Where(
			sitenotification.UIDEQ(uid),
			sitenotification.ReadEQ(0),
		).
		SetRead(1).
		Save(ctx)
	return err
}

// Since 返回 id > seq 的通知列表（HTTP 骨架，Plan 08 完善 WS 补漏）。
func (s *Service) Since(ctx context.Context, uid, seq int) ([]NotificationItem, error) {
	q := s.client.SiteNotification.Query().
		Where(sitenotification.UIDEQ(uid))
	if seq > 0 {
		q = q.Where(sitenotification.IDGT(seq))
	}
	rows, err := q.Order(ent.Asc(sitenotification.FieldID)).Limit(100).All(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]NotificationItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toNotificationItem(row))
	}
	return list, nil
}

func toNotificationItem(row *ent.SiteNotification) NotificationItem {
	payload := map[string]interface{}{}
	if row.Payload != "" {
		_ = json.Unmarshal([]byte(row.Payload), &payload)
	}
	return NotificationItem{
		ID:         row.ID,
		Type:       row.Type,
		Payload:    payload,
		Read:       row.Read,
		CreateTime: row.CreateTime,
	}
}
