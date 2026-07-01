// msgboard_service 留言板业务逻辑（限流与敏感词）。
package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/util"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/sensitive"
	userrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
)

const (
	msgboardRateWindowSec = 24 * 60 * 60
	msgboardRateMax       = 10
)

// MsgboardService 留言板业务逻辑。
type MsgboardService struct {
	msgboards *blogrepo.MsgboardRepo
	filter    sensitive.FilterService
	redis     *redisutil.Store
}

// NewMsgboardService 构造 MsgboardService。
func NewMsgboardService(
	msgboards *blogrepo.MsgboardRepo,
	filter sensitive.FilterService,
	redis *redisutil.Store,
) *MsgboardService {
	return &MsgboardService{msgboards: msgboards, filter: filter, redis: redis}
}

// Create 创建留言。
func (s *MsgboardService) Create(ctx context.Context, in map[string]interface{}, ip, userAgent string) (*ent.Msgboard, error) {
	if err := s.assertRateLimit(ctx, ip); err != nil {
		return nil, err
	}
	ip = normalizeIP(ip)
	commentText := strField(in, "comment")
	eval, err := s.filter.EvaluateContent(ctx, commentText)
	if err != nil {
		return nil, err
	}
	if len(eval.HitWords) > 0 && eval.Rejected {
		ipPtr := &ip
		recordSensitiveHit(ctx, s.filter, "msgboard", "0", commentText, eval.HitWords, nil, ipPtr)
		return nil, errcode.WithMessage(errcode.InvalidParam, "内容包含违规词汇，无法发布")
	}
	status := "approved"
	if eval.NeedReview {
		status = "pending"
	}
	name := util.EscapeHTML(strField(in, "name"))
	respondent := util.EscapeHTML(strField(in, "respondent"))
	body := commentText
	if len(eval.HitWords) > 0 {
		body = eval.Content
	}
	email := strField(in, "eamil")
	if email == "" {
		email = strField(in, "email")
	}
	avatar := strField(in, "avatar")
	if avatar == "" && email != "" {
		sum := md5.Sum([]byte(email))
		avatar = fmt.Sprintf("https://cravatar.cn/avatar/%s?s=100", hex.EncodeToString(sum[:]))
	}
	osName, browser := parseUA(userAgent)
	row, err := s.msgboards.Create(ctx, &ent.Msgboard{
		Name:       name,
		Eamil:      email,
		Address:    strField(in, "address"),
		Comment:    util.EscapeHTML(body),
		Avatar:     avatar,
		Location:   "未知",
		System:     osName,
		Browser:    browser,
		Status:     status,
		PId:        intField(in, "pId", 0),
		Respondent: strPtr(respondent),
		ImgUrl:     strPtr(strField(in, "imgUrl")),
		IP:         &ip,
	})
	if err != nil {
		return nil, err
	}
	ipPtr := &ip
	recordSensitiveHit(ctx, s.filter, "msgboard", fmt.Sprintf("%d", row.ID), commentText, eval.HitWords, nil, ipPtr)
	return row, nil
}

// List 留言列表。
func (s *MsgboardService) List(ctx context.Context, q map[string]interface{}) (map[string]interface{}, error) {
	page := intField(q, "page", 1)
	pageSize := intField(q, "pageSize", 10)
	status := strField(q, "status")
	rows, total, err := s.msgboards.List(ctx, blogrepo.MsgboardFilter{
		Status:   status,
		Name:     strField(q, "name"),
		Comment:  strField(q, "comment"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, m := range rows {
		list = append(list, msgboardToMap(m))
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": userrepo.CalcNestPagination(total, pageSize, page),
	}, nil
}

// Delete 批量删除留言。
func (s *MsgboardService) Delete(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return errcode.InvalidParam
	}
	n, err := s.msgboards.DeleteByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if n == 0 {
		return errcode.WithMessage(errcode.InternalError, "删除失败")
	}
	return nil
}

func (s *MsgboardService) assertRateLimit(ctx context.Context, ip string) error {
	key := "msgboard:rate:" + sanitizeRateKey(normalizeIP(ip))
	count, err := s.redis.Incr(ctx, key)
	if err != nil {
		return err
	}
	if count == 1 {
		// 按 IP 限流：msgboardRateWindowSec=24h，msgboardRateMax=10 条/天。
		_ = s.redis.Expire(ctx, key, msgboardRateWindowSec)
	}
	if count > msgboardRateMax {
		return errcode.WithMessage(errcode.TooManyRequests, "一天只能留言10条哦！")
	}
	return nil
}

func msgboardToMap(m *ent.Msgboard) map[string]interface{} {
	item := map[string]interface{}{
		"id": m.ID, "name": m.Name, "eamil": m.Eamil, "address": m.Address,
		"comment": m.Comment, "avatar": m.Avatar, "location": m.Location,
		"system": m.System, "browser": m.Browser, "pId": m.PId,
		"status": m.Status, "createTime": m.CreateTime, "updateTime": m.UpdateTime,
	}
	if m.Respondent != nil {
		item["respondent"] = *m.Respondent
	}
	if m.ImgUrl != nil {
		item["imgUrl"] = *m.ImgUrl
	}
	if m.IP != nil {
		item["ip"] = *m.IP
	}
	if m.ReplyId != nil {
		item["replyId"] = *m.ReplyId
	}
	return item
}

func normalizeIP(ip string) string {
	if strings.HasPrefix(ip, "::ffff:") {
		return ip[7:]
	}
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}

func parseUA(ua string) (osName, browser string) {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "windows"):
		osName = "Windows"
	case strings.Contains(ua, "mac"):
		osName = "MacOS"
	case strings.Contains(ua, "android"):
		osName = "Android"
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"):
		osName = "iOS"
	default:
		osName = "Unknown"
	}
	switch {
	case strings.Contains(ua, "chrome"):
		browser = "Chrome"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "safari"):
		browser = "Safari"
	case strings.Contains(ua, "edge"):
		browser = "Edge"
	default:
		browser = "Unknown"
	}
	return osName, browser
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
