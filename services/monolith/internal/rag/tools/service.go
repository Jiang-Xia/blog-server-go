package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/link"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/msgboard"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/crossdb"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/leaderboard"
)

// Service RAG 只读 Tool 实现（blog 域 + 跨库 user + 单体 RPG）。
type Service struct {
	client      *ent.Client
	cross       *crossdb.CrossDB
	leaderboard *leaderboard.Service
	rpg         *rpgcore.RpgService
}

// NewService 构造 Tool Service。
func NewService(client *ent.Client, cross *crossdb.CrossDB, lb *leaderboard.Service, rpg *rpgcore.RpgService) *Service {
	return &Service{client: client, cross: cross, leaderboard: lb, rpg: rpg}
}

// Execute 按名称分发 Tool。
func (s *Service) Execute(ctx context.Context, name string, args map[string]interface{}, toolCtx Context) (interface{}, error) {
	switch name {
	case "get_article_ranking":
		return s.getArticleRanking(ctx, strArg(args, "metric", "views"), intArg(args, "limit", 10))
	case "list_authors":
		return s.listAuthors(ctx, intArg(args, "page", 1), intArg(args, "pageSize", 20))
	case "get_author_stats":
		return s.cross.RagAuthorStats(ctx, intArg(args, "uid", 0), strArg(args, "nickname", ""))
	case "search_articles":
		items, err := s.cross.RagSearchArticles(ctx, strArg(args, "keyword", ""), intArg(args, "limit", 10))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"keyword": strArg(args, "keyword", ""), "items": mapArticleItems(items)}, nil
	case "list_articles_by_category":
		items, err := s.cross.RagListByCategory(ctx, strArg(args, "categoryName", ""), intArg(args, "limit", 10))
		if err != nil {
			return map[string]interface{}{"error": err.Error()}, nil
		}
		return map[string]interface{}{"category": strArg(args, "categoryName", ""), "items": mapArticleItems(items)}, nil
	case "list_articles_by_tag":
		items, err := s.cross.RagListByTag(ctx, strArg(args, "tagName", ""), intArg(args, "limit", 10))
		if err != nil {
			return map[string]interface{}{"error": err.Error()}, nil
		}
		return map[string]interface{}{"tag": strArg(args, "tagName", ""), "items": mapArticleItems(items)}, nil
	case "get_recent_articles":
		return s.getRecentArticles(ctx, intArg(args, "limit", 10))
	case "get_masterpiece_articles":
		return s.getMasterpieceArticles(ctx, intArg(args, "limit", 10))
	case "get_article_stats":
		return s.cross.RagArticleStats(ctx, intArg(args, "articleId", 0), strArg(args, "title", ""))
	case "get_category_stats":
		return s.getCategoryStats(ctx, intArg(args, "limit", 20))
	case "get_tag_cloud":
		return s.getTagCloud(ctx, intArg(args, "limit", 20))
	case "get_site_nav":
		return s.getSiteNav(), nil
	case "search_site_pages":
		return s.searchSitePages(strArg(args, "keyword", ""), intArg(args, "limit", 10)), nil
	case "get_my_article_stats":
		if toolCtx.RequestUID <= 0 {
			return map[string]interface{}{"error": "请先登录后查询个人发文统计"}, nil
		}
		return s.cross.RagAuthorStats(ctx, toolCtx.RequestUID, "")
	case "get_msgboard_recent":
		return s.getMsgboardRecent(ctx, intArg(args, "limit", 10))
	case "list_friend_links":
		return s.listFriendLinks(ctx, intArg(args, "limit", 30))
	case "get_article_archive_stats":
		y := intArg(args, "year", 0)
		if y > 0 {
			return s.getArchiveStats(ctx, &y)
		}
		return s.getArchiveStats(ctx, nil)
	case "get_rpg_leaderboard":
		return s.getRpgLeaderboard(ctx, strArg(args, "type", "exp"), strArg(args, "period", "total"), intArg(args, "limit", 10))
	case "get_my_rpg_status":
		return s.getMyRpgStatus(ctx, toolCtx)
	default:
		return map[string]interface{}{"error": "未知工具: " + name}, nil
	}
}

// RunToolCalls 批量执行 LLM 返回的 tool_calls。
func (s *Service) RunToolCalls(ctx context.Context, calls []LLMToolCall, toolCtx Context) []CallRecord {
	records := make([]CallRecord, 0, len(calls))
	for _, call := range calls {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
			args = map[string]interface{}{}
		}
		res, err := s.Execute(ctx, call.Name, args, toolCtx)
		if err != nil {
			res = map[string]interface{}{"error": err.Error()}
		}
		records = append(records, CallRecord{Name: call.Name, Args: args, Result: res})
	}
	return records
}

func (s *Service) getRpgLeaderboard(ctx context.Context, scoreType, period string, limit int) (map[string]interface{}, error) {
	if s.leaderboard == nil {
		return map[string]interface{}{"error": "RPG 排行榜服务不可用"}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	items, err := s.leaderboard.GetLeaderboard(ctx, leaderboard.ScoreType(scoreType), leaderboard.Period(period), limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"type": scoreType, "period": period, "items": items}, nil
}

func (s *Service) getMyRpgStatus(ctx context.Context, toolCtx Context) (map[string]interface{}, error) {
	if toolCtx.RequestUID <= 0 {
		return map[string]interface{}{"error": "请先登录后查询个人 RPG 状态"}, nil
	}
	if s.rpg == nil {
		return map[string]interface{}{"error": "RPG 服务不可用"}, nil
	}
	status, err := s.rpg.GetFullStatus(ctx, toolCtx.RequestUID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"uid":                 toolCtx.RequestUID,
		"level":               status.Level,
		"exp":                 status.Exp,
		"lifeValue":           status.LifeValue,
		"reputation":          status.Reputation,
		"currency":            status.Currency,
		"lotteryTickets":      status.LotteryTickets,
		"totalSignDays":       status.TotalSignDays,
		"consecutiveSignDays": status.ConsecutiveSignDays,
		"equippedTitle":       status.EquippedTitle,
		"equippedAvatarFrame": status.EquippedAvatarFrame,
		"rpgUrl":              "/rpg",
	}, nil
}

func (s *Service) getArticleRanking(ctx context.Context, metric string, limit int) (map[string]interface{}, error) {
	items, err := s.cross.RagArticleRanking(ctx, metric, limit)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, len(items))
	for i, it := range items {
		row := map[string]interface{}{
			"articleId": it.ArticleID, "title": it.Title, "url": it.URL,
			"views": it.Views, "likes": it.Likes,
		}
		if it.Comments != nil {
			row["comments"] = *it.Comments
		}
		if it.Collects != nil {
			row["collects"] = *it.Collects
		}
		list[i] = row
	}
	return map[string]interface{}{"metric": metric, "items": list}, nil
}

func (s *Service) listAuthors(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	rows, total, err := s.cross.RagListAuthors(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		items[i] = map[string]interface{}{
			"uid": r.UID, "nickname": r.Nickname, "articleCount": r.ArticleCount,
			"profileUrl": "/user/" + itoa(r.UID),
		}
	}
	return map[string]interface{}{"page": page, "pageSize": pageSize, "total": total, "items": items}, nil
}

func (s *Service) getRecentArticles(ctx context.Context, limit int) (map[string]interface{}, error) {
	items, err := s.cross.RagRecentArticles(ctx, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"items": mapArticleItems(items)}, nil
}

func (s *Service) getMasterpieceArticles(ctx context.Context, limit int) (map[string]interface{}, error) {
	items, err := s.cross.RagMasterpieceArticles(ctx, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"items": mapArticleItems(items)}, nil
}

func mapArticleItems(items []crossdb.RagArticleItem) []map[string]interface{} {
	out := make([]map[string]interface{}, len(items))
	for i, a := range items {
		out[i] = map[string]interface{}{
			"articleId": a.ArticleID, "title": a.Title, "url": a.URL,
			"views": a.Views, "likes": a.Likes,
			"articleLevel": a.ArticleLevel, "isMasterpiece": a.IsMasterpiece,
			"createTime": a.CreateTime,
		}
		if a.CategoryLabel != nil {
			out[i]["category"] = *a.CategoryLabel
		} else {
			out[i]["category"] = nil
		}
		if a.TagLabels != nil && *a.TagLabels != "" {
			out[i]["tags"] = strings.Split(*a.TagLabels, ",")
		} else {
			out[i]["tags"] = []string{}
		}
	}
	return out
}

func (s *Service) getCategoryStats(ctx context.Context, limit int) (map[string]interface{}, error) {
	rows, err := s.cross.RagCategoryStats(ctx, limit)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		items[i] = map[string]interface{}{"category": r.Category, "articleCount": r.ArticleCount}
	}
	return map[string]interface{}{"items": items}, nil
}

func (s *Service) getTagCloud(ctx context.Context, limit int) (map[string]interface{}, error) {
	rows, err := s.cross.RagTagCloud(ctx, limit)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		items[i] = map[string]interface{}{"tag": r.Tag, "articleCount": r.ArticleCount}
	}
	return map[string]interface{}{"items": items}, nil
}

func (s *Service) getSiteNav() map[string]interface{} {
	nav := make([]map[string]string, len(siteNavLinks))
	for i, l := range siteNavLinks {
		nav[i] = map[string]string{"path": l.Path, "title": l.Title}
	}
	tools := make([]map[string]string, len(toolLinks))
	for i, l := range toolLinks {
		tools[i] = map[string]string{"path": l.Path, "title": l.Title}
	}
	pages := make([]map[string]interface{}, len(featurePages))
	for i, p := range featurePages {
		pages[i] = map[string]interface{}{"title": p.Title, "url": p.URL, "description": p.Description}
	}
	return map[string]interface{}{"navLinks": nav, "toolLinks": tools, "featurePages": pages}
}

func (s *Service) searchSitePages(keyword string, limit int) map[string]interface{} {
	q := strings.ToLower(strings.TrimSpace(keyword))
	if q == "" {
		return map[string]interface{}{"items": []interface{}{}}
	}
	type poolItem struct {
		title, url, itemType string
		description          string
	}
	var pool []poolItem
	for _, l := range siteNavLinks {
		pool = append(pool, poolItem{l.Title, l.Path, "nav", ""})
	}
	for _, l := range toolLinks {
		pool = append(pool, poolItem{l.Title, l.Path, "tool", ""})
	}
	for _, p := range featurePages {
		pool = append(pool, poolItem{p.Title, p.URL, "page", p.Description})
	}
	var items []map[string]interface{}
	for _, item := range pool {
		if strings.Contains(strings.ToLower(item.title), q) ||
			strings.Contains(strings.ToLower(item.url), q) ||
			strings.Contains(strings.ToLower(item.description), q) {
			row := map[string]interface{}{"title": item.title, "url": item.url, "type": item.itemType}
			if item.description != "" {
				row["description"] = item.description
			}
			items = append(items, row)
			if len(items) >= limit {
				break
			}
		}
	}
	return map[string]interface{}{"keyword": q, "items": items}
}

func (s *Service) getMsgboardRecent(ctx context.Context, limit int) (map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	list, err := s.client.Msgboard.Query().
		Where(msgboard.StatusEQ("approved")).
		Order(ent.Desc(msgboard.FieldCreateTime)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(list))
	for i, m := range list {
		excerpt := m.Comment
		if len([]rune(excerpt)) > 120 {
			excerpt = string([]rune(excerpt)[:120])
		}
		items[i] = map[string]interface{}{
			"id": m.ID, "name": m.Name, "excerpt": excerpt,
			"createTime": m.CreateTime, "isReply": m.PId > 0,
		}
	}
	return map[string]interface{}{"url": "/msgboard", "items": items}, nil
}

func (s *Service) listFriendLinks(ctx context.Context, limit int) (map[string]interface{}, error) {
	if limit <= 0 {
		limit = 30
	}
	list, err := s.client.Link.Query().
		Where(link.AgreedEQ(1)).
		Order(ent.Desc(link.FieldCreateTime)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(list))
	for i, l := range list {
		desp := l.Desp
		if len([]rune(desp)) > 80 {
			desp = string([]rune(desp)[:80])
		}
		items[i] = map[string]interface{}{
			"title": l.Title, "url": l.URL, "description": desp, "status": l.LastCheckStatus,
		}
	}
	return map[string]interface{}{"url": "/links", "items": items}, nil
}

func (s *Service) getArchiveStats(ctx context.Context, year *int) (map[string]interface{}, error) {
	rows, err := s.cross.RagArchiveStats(ctx, year)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		items[i] = map[string]interface{}{"year": r.Year, "articleCount": r.ArticleCount}
	}
	out := map[string]interface{}{"items": items, "archiveUrl": "/archives"}
	if year != nil {
		out["year"] = *year
	} else {
		out["year"] = nil
	}
	return out, nil
}

func strArg(args map[string]interface{}, key, def string) string {
	if v, ok := args[key]; ok {
		switch t := v.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				return t
			}
		}
	}
	return def
}

func intArg(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key]; ok {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case json.Number:
			n, _ := t.Int64()
			return int(n)
		}
	}
	return def
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
