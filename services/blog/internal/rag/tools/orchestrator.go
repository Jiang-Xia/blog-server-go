package tools

import (
	"context"
	"regexp"
	"strings"
)

// Orchestrator 规则匹配 Tool 路由（对齐 Nest tryRuleBasedTools；RPG 榜留 Plan 17）。
type Orchestrator struct {
	svc *Service
}

// NewOrchestrator 构造 Orchestrator。
func NewOrchestrator(svc *Service) *Orchestrator {
	return &Orchestrator{svc: svc}
}

// ResolveTools 根据问题决定是否调用 Tool。
func (o *Orchestrator) ResolveTools(ctx context.Context, question string, toolCtx Context) ([]CallRecord, error) {
	q := question
	if q == "" {
		return nil, nil
	}

	rankingPatterns := []struct {
		re     *regexp.Regexp
		metric string
	}{
		{regexp.MustCompile(`浏览|阅读`), "views"},
		{regexp.MustCompile(`点赞`), "likes"},
		{regexp.MustCompile(`收藏`), "collects"},
		{regexp.MustCompile(`评论`), "comments"},
	}
	if regexp.MustCompile(`(?i)(最多|最高|排名|排行|top)`).MatchString(q) {
		for _, item := range rankingPatterns {
			if item.re.MatchString(q) {
				args := map[string]interface{}{"metric": item.metric, "limit": 10}
				res, err := o.svc.Execute(ctx, "get_article_ranking", args, toolCtx)
				if err != nil {
					return nil, err
				}
				return []CallRecord{{Name: "get_article_ranking", Args: args, Result: res}}, nil
			}
		}
	}

	rules := []struct {
		re   *regexp.Regexp
		name string
		args map[string]interface{}
	}{
		{regexp.MustCompile(`有哪些作者|作者列表|全站作者|写手列表`), "list_authors", map[string]interface{}{"page": 1, "pageSize": 20}},
		{regexp.MustCompile(`最新文章|最近发布|最近发了`), "get_recent_articles", map[string]interface{}{"limit": 10}},
		{regexp.MustCompile(`(?i)神作|masterpiece`), "get_masterpiece_articles", map[string]interface{}{"limit": 10}},
		{regexp.MustCompile(`热门标签|标签云|有哪些标签`), "get_tag_cloud", map[string]interface{}{"limit": 20}},
		{regexp.MustCompile(`分类.*(多少|统计|有哪些)|哪个分类`), "get_category_stats", map[string]interface{}{"limit": 20}},
		{regexp.MustCompile(`友链|友情链接`), "list_friend_links", map[string]interface{}{"limit": 30}},
		{regexp.MustCompile(`留言板|最近留言`), "get_msgboard_recent", map[string]interface{}{"limit": 10}},
		{regexp.MustCompile(`工具箱|有哪些工具|站点导航|去哪个页面`), "get_site_nav", map[string]interface{}{}},
	}

	for _, rule := range rules {
		if rule.re.MatchString(q) {
			res, err := o.svc.Execute(ctx, rule.name, rule.args, toolCtx)
			if err != nil {
				return nil, err
			}
			return []CallRecord{{Name: rule.name, Args: rule.args, Result: res}}, nil
		}
	}

	// RPG 榜：返回友好提示（Plan 17 接 gRPC 后再接真实数据）
	if regexp.MustCompile(`RPG.*(排行|榜)|经验榜|等级榜|声望榜|签到榜`).MatchString(q) {
		args := map[string]interface{}{"type": "exp", "period": "total", "limit": 10}
		res, _ := o.svc.Execute(ctx, "get_rpg_leaderboard", args, toolCtx)
		return []CallRecord{{Name: "get_rpg_leaderboard", Args: args, Result: res}}, nil
	}

	if regexp.MustCompile(`(我的|我).*(等级|RPG|经验|钻石|几级)`).MatchString(q) {
		res, _ := o.svc.Execute(ctx, "get_my_rpg_status", map[string]interface{}{}, toolCtx)
		return []CallRecord{{Name: "get_my_rpg_status", Args: map[string]interface{}{}, Result: res}}, nil
	}

	if regexp.MustCompile(`(我的|我).*(发了|文章数|发文)`).MatchString(q) {
		res, err := o.svc.Execute(ctx, "get_my_article_stats", map[string]interface{}{}, toolCtx)
		if err != nil {
			return nil, err
		}
		return []CallRecord{{Name: "get_my_article_stats", Args: map[string]interface{}{}, Result: res}}, nil
	}

	if regexp.MustCompile(`(20\d{2})年.*(多少|几)篇|归档统计`).MatchString(q) {
		args := map[string]interface{}{}
		if m := regexp.MustCompile(`(20\d{2})`).FindStringSubmatch(q); len(m) > 1 {
			args["year"] = mustAtoi(m[1])
		}
		res, err := o.svc.Execute(ctx, "get_article_archive_stats", args, toolCtx)
		if err != nil {
			return nil, err
		}
		return []CallRecord{{Name: "get_article_archive_stats", Args: args, Result: res}}, nil
	}

	if m := regexp.MustCompile(`(.{1,20}?)(有多少|几)篇(文章)?`).FindStringSubmatch(q); len(m) > 1 {
		nickname := regexp.MustCompile(`作者|写手|的`).ReplaceAllString(m[1], "")
		nickname = strings.TrimSpace(nickname)
		if nickname != "" && nickname != "我" {
			args := map[string]interface{}{"nickname": nickname}
			res, err := o.svc.Execute(ctx, "get_author_stats", args, toolCtx)
			if err != nil {
				return nil, err
			}
			return []CallRecord{{Name: "get_author_stats", Args: args, Result: res}}, nil
		}
	}

	return nil, nil
}

func mustAtoi(s string) int {
	n := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}
	return n
}
