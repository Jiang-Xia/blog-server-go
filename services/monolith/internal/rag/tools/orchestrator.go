package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
)

// Orchestrator 规则匹配 + LLM function calling Tool 路由（对齐 Nest RagOrchestratorService）。
type Orchestrator struct {
	svc    *Service
	cfg    *config.Config
	client *http.Client
}

// NewOrchestrator 构造 Orchestrator。
func NewOrchestrator(svc *Service, cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		svc: svc, cfg: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// ResolveTools 根据问题决定是否调用 Tool。
func (o *Orchestrator) ResolveTools(ctx context.Context, question string, toolCtx Context) ([]CallRecord, error) {
	q := strings.TrimSpace(question)
	if q == "" {
		return nil, nil
	}

	if recs, err := o.tryRuleBasedTools(ctx, q, toolCtx); err != nil {
		return nil, err
	} else if len(recs) > 0 {
		return recs, nil
	}

	return o.tryLLMTools(ctx, q, toolCtx)
}

func (o *Orchestrator) tryRuleBasedTools(ctx context.Context, q string, toolCtx Context) ([]CallRecord, error) {
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
		{regexp.MustCompile(`搜索.*文章|找.*文章|有没有.*文章`), "search_articles", map[string]interface{}{"keyword": extractSearchKeyword(q), "limit": 10}},
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

	if regexp.MustCompile(`RPG.*(排行|榜)|经验榜|等级榜|声望榜|签到榜`).MatchString(q) {
		args := map[string]interface{}{"type": inferRpgScoreType(q), "period": "total", "limit": 10}
		res, err := o.svc.Execute(ctx, "get_rpg_leaderboard", args, toolCtx)
		if err != nil {
			return nil, err
		}
		return []CallRecord{{Name: "get_rpg_leaderboard", Args: args, Result: res}}, nil
	}

	if regexp.MustCompile(`(我的|我).*(等级|RPG|经验|钻石|几级)`).MatchString(q) {
		res, err := o.svc.Execute(ctx, "get_my_rpg_status", map[string]interface{}{}, toolCtx)
		if err != nil {
			return nil, err
		}
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

func (o *Orchestrator) tryLLMTools(ctx context.Context, question string, toolCtx Context) ([]CallRecord, error) {
	if o.cfg == nil || o.svc == nil {
		return nil, nil
	}
	apiKey := strings.TrimSpace(o.cfg.Rag.LLM.APIKey)
	if apiKey == "" {
		return nil, nil
	}
	base := strings.TrimSuffix(o.cfg.Rag.LLM.BaseURL, "/")
	if base == "" {
		base = "https://api.deepseek.com/v1"
	}
	model := o.cfg.Rag.LLM.Model
	if model == "" {
		model = "deepseek-chat"
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "你是博客助手的路由模块。遇到排行、作者、分类标签、最新/神作文章、RPG 榜、站点导航、友链留言、归档统计、个人 RPG/发文等结构化问题时，必须调用对应工具。纯教程/玩法/文章内容类问题不要调用工具。",
			},
			{"role": "user", "content": question},
		},
		"tools":       ragToolDefinitions,
		"tool_choice": "auto",
		"temperature": 0,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode >= 400 {
		return nil, nil
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				ToolCalls []struct {
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, nil
	}
	if len(parsed.Choices) == 0 || len(parsed.Choices[0].Message.ToolCalls) == 0 {
		return nil, nil
	}
	calls := make([]LLMToolCall, 0, len(parsed.Choices[0].Message.ToolCalls))
	for _, tc := range parsed.Choices[0].Message.ToolCalls {
		calls = append(calls, LLMToolCall{Name: tc.Function.Name, Arguments: tc.Function.Arguments})
	}
	return o.svc.RunToolCalls(ctx, calls, toolCtx), nil
}

func extractSearchKeyword(q string) string {
	if m := regexp.MustCompile(`(?:搜索|找|查)(?:一下|下)?[「"']?(.+?)[」"']?(?:相关|的)?文章`).FindStringSubmatch(q); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(q)
}

func inferRpgScoreType(q string) string {
	switch {
	case strings.Contains(q, "等级"):
		return "level"
	case strings.Contains(q, "声望"):
		return "reputation"
	case strings.Contains(q, "钻石"):
		return "currency"
	case strings.Contains(q, "签到"):
		return "signDays"
	default:
		return "exp"
	}
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
