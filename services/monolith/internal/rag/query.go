package rag

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/ragquerylog"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rag/tools"
	"go.uber.org/zap"
)

// ChatMessage OpenAI chat message。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// QueryLogParams 写入 rag_query_log 参数。
type QueryLogParams struct {
	UID           int
	Question      string
	AnswerPreview string
	Citations     []Citation
	LatencyMs     int
	Status        string
}

// QueryService RAG 问答编排。
type QueryService struct {
	cfg          *config.Config
	client       *ent.Client
	embedding    *EmbeddingService
	hybrid       *HybridSearch
	orchestrator *tools.Orchestrator
	log          *zap.Logger
	clientHTTP   *http.Client
}

// NewQueryService 构造 QueryService。
func NewQueryService(cfg *config.Config, client *ent.Client, embedding *EmbeddingService, hybrid *HybridSearch, orchestrator *tools.Orchestrator, log *zap.Logger) *QueryService {
	return &QueryService{
		cfg: cfg, client: client, embedding: embedding, hybrid: hybrid,
		orchestrator: orchestrator, log: log,
		clientHTTP: &http.Client{Timeout: 120 * time.Second},
	}
}

// AssertEnabled 校验 RAG 开关与 Embedding。
func (s *QueryService) AssertEnabled() error {
	if !s.cfg.Rag.Enabled {
		return errcode.WithMessage(errcode.ServiceUnavailable, "AI 助手暂未开启")
	}
	if !s.embedding.IsAvailable() {
		return errcode.WithMessage(errcode.ServiceUnavailable, "AI 助手未配置 API Key")
	}
	return nil
}

// Retrieve 混合检索并生成 citations。
func (s *QueryService) Retrieve(ctx context.Context, question string) ([]SearchHit, []Citation, error) {
	q := normalizeQuestion(question)
	vec, err := s.embedding.EmbedOne(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	hits, err := s.hybrid.Search(ctx, s.cfg, q, vec, s.cfg.Rag.RagTopKOrDefault())
	if err != nil {
		return nil, nil, err
	}
	citations := make([]Citation, len(hits))
	for i, h := range hits {
		snippet := h.Content
		if len([]rune(snippet)) > 200 {
			snippet = string([]rune(snippet)[:200])
		}
		citations[i] = Citation{
			ArticleID: h.ArticleID, Title: h.Title, URL: h.URL,
			Snippet: snippet, SourceType: h.SourceType,
		}
	}
	return hits, citations, nil
}

// PrepareQuery 检索、Tool 路由并拼装 messages。
func (s *QueryService) PrepareQuery(ctx context.Context, question string, requestUID int, history []ChatTurn) ([]ChatMessage, []Citation, error) {
	q := normalizeQuestion(question)
	vec, err := s.embedding.EmbedOne(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	hits, err := s.hybrid.Search(ctx, s.cfg, q, vec, s.cfg.Rag.RagTopKOrDefault())
	if err != nil {
		return nil, nil, err
	}
	citations := make([]Citation, len(hits))
	for i, h := range hits {
		snippet := h.Content
		if len([]rune(snippet)) > 200 {
			snippet = string([]rune(snippet)[:200])
		}
		citations[i] = Citation{
			ArticleID: h.ArticleID, Title: h.Title, URL: h.URL,
			Snippet: snippet, SourceType: h.SourceType,
		}
	}

	var toolRecords []tools.CallRecord
	if s.orchestrator != nil {
		toolRecords, err = s.orchestrator.ResolveTools(ctx, q, tools.Context{RequestUID: requestUID})
		if err != nil {
			return nil, nil, err
		}
	}

	messages := buildMessages(q, hits, toolRecords, history)
	return messages, citations, nil
}

func buildMessages(question string, hits []SearchHit, toolRecords []tools.CallRecord, history []ChatTurn) []ChatMessage {
	var contextParts []string
	for i, h := range hits {
		part := fmt.Sprintf("[%d] 标题：%s\n链接：%s\n章节：%s\n内容：\n%s",
			i+1, h.Title, h.URL, orDash(h.HeadingPath), h.Content)
		contextParts = append(contextParts, part)
	}
	contextText := strings.Join(contextParts, "\n\n---\n\n")

	var toolParts []string
	for i, t := range toolRecords {
		b, _ := json.Marshal(t.Result)
		toolParts = append(toolParts, fmt.Sprintf("[Tool %d] %s\n参数：%s\n结果：%s",
			i+1, t.Name, mustJSON(t.Args), string(b)))
	}
	toolContext := strings.Join(toolParts, "\n\n")

	system := `你是博客网站助手，结合「检索内容」与「工具查询结果」回答用户问题。
规则：
1. 优先使用检索内容与工具结果，不要编造事实。
2. 工具结果用于排行、作者统计、分类标签、友链留言、站点导航等实时数据；检索内容用于教程、玩法、文章知识。
3. 不知道或信息不足时，明确说明「根据现有资料未找到相关信息」。
4. 回答末尾列出引用来源：文章用 [1]《标题》格式；工具数据注明「数据来源：站内统计」。
5. 使用简洁中文。`
	if len(history) > 0 {
		system += "\n6. 用户可能在追问上一轮内容，结合对话历史理解指代。"
	}

	userParts := []string{}
	if contextText != "" {
		userParts = append(userParts, "检索内容：\n"+contextText)
	} else {
		userParts = append(userParts, "检索内容：（无检索结果）")
	}
	if toolContext != "" {
		userParts = append(userParts, "工具查询结果：\n"+toolContext)
	}
	userParts = append(userParts, "用户问题："+question)

	var messages []ChatMessage
	messages = append(messages, ChatMessage{Role: "system", Content: system})
	for _, t := range history {
		messages = append(messages, ChatMessage{Role: t.Role, Content: t.Content})
	}
	messages = append(messages, ChatMessage{Role: "user", Content: strings.Join(userParts, "\n\n")})
	return messages
}

func mustJSON(v map[string]interface{}) string {
	if len(v) == 0 {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

// StreamChat 调用 LLM 流式补全，通过 channel 推送 delta。
func (s *QueryService) StreamChat(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error) {
	out := make(chan string, 32)
	errCh := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errCh)
		apiKey := strings.TrimSpace(s.cfg.Rag.LLM.APIKey)
		if apiKey == "" {
			errCh <- fmt.Errorf("LLM API Key 未配置")
			return
		}
		base := strings.TrimSuffix(s.cfg.Rag.LLM.BaseURL, "/")
		if base == "" {
			base = "https://api.deepseek.com/v1"
		}
		model := s.cfg.Rag.LLM.Model
		if model == "" {
			model = "deepseek-chat"
		}
		body, _ := json.Marshal(map[string]interface{}{
			"model":       model,
			"messages":    messages,
			"stream":      true,
			"temperature": 0.3,
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		resp, err := s.clientHTTP.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			data, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("LLM API %d: %s", resp.StatusCode, string(data))
			return
		}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				break
			}
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}
			if delta := chunk.Choices[0].Delta.Content; delta != "" {
				out <- delta
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
		}
	}()
	return out, errCh
}

// SaveQueryLog 异步写入 rag_query_log。
func (s *QueryService) SaveQueryLog(ctx context.Context, p QueryLogParams) {
	citationsJSON := make([]map[string]interface{}, len(p.Citations))
	for i, c := range p.Citations {
		citationsJSON[i] = map[string]interface{}{
			"articleId": c.ArticleID, "title": c.Title, "url": c.URL,
			"snippet": c.Snippet, "sourceType": c.SourceType,
		}
	}
	question := p.Question
	if len([]rune(question)) > MaxQuestionChars {
		question = string([]rune(question)[:MaxQuestionChars])
	}
	preview := p.AnswerPreview
	if len([]rune(preview)) > 1000 {
		preview = string([]rune(preview)[:1000])
	}
	_, err := s.client.RagQueryLog.Create().
		SetUID(p.UID).
		SetQuestion(question).
		SetNillableAnswerPreview(strPtr(preview)).
		SetCitationsJSON(citationsJSON).
		SetLatencyMs(p.LatencyMs).
		SetStatus(p.Status).
		SetCreateAt(time.Now()).
		Save(ctx)
	if err != nil && s.log != nil {
		s.log.Warn("save rag query log failed", zap.Error(err))
	}
}

func normalizeQuestion(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		panic("empty question")
	}
	if len([]rune(q)) > MaxQuestionChars {
		q = string([]rune(q)[:MaxQuestionChars])
	}
	return q
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ListQueryLogs 管理端查询日志（供 Admin 使用）。
func (s *QueryService) ListQueryLogs(ctx context.Context, uid, page, pageSize int) ([]*ent.RagQueryLog, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := s.client.RagQueryLog.Query()
	if uid > 0 {
		q = q.Where(ragquerylog.UIDEQ(uid))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	list, err := q.Order(ent.Desc(ragquerylog.FieldCreateAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	return list, total, err
}
