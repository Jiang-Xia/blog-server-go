// rag_handler RAG 知识库 HTTP 端点（C 端 + admin）。
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/ctxutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/response"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rag"
	"github.com/cloudwego/hertz/pkg/app"
)

// RagHandler RAG HTTP 端点。
type RagHandler struct {
	mod *rag.Module
	jwt *auth.JWTService
}

// NewRagHandler 构造 RagHandler。
func NewRagHandler(mod *rag.Module, jwt *auth.JWTService) *RagHandler {
	return &RagHandler{mod: mod, jwt: jwt}
}

// Quota GET /rag/quota
func (h *RagHandler) Quota(ctx context.Context, c *app.RequestContext) {
	uid := ragUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}
	usage, err := h.mod.Quota.GetUsage(ctx, uid)
	if err != nil {
		response.FromError(ctx, c, err)
		return
	}
	response.Success(ctx, c, usage)
}

// Status GET /rag/status（公开）
func (h *RagHandler) Status(ctx context.Context, c *app.RequestContext) {
	chunkCount, _ := h.mod.Hybrid.CountActive(ctx)
	response.Success(ctx, c, map[string]interface{}{
		"enabled":                   h.mod.Cfg.Rag.Enabled,
		"configured":                h.mod.Embedding.IsAvailable(),
		"embeddingMode":             h.mod.Embedding.GetMode(),
		"embeddingRemoteConfigured": h.mod.Embedding.IsRemoteConfigured(),
		"chunkCount":                chunkCount,
		"ready":                     chunkCount > 0 && h.mod.Cfg.Rag.Enabled,
	})
}

// QueryStream POST /rag/query-stream（SSE）
func (h *RagHandler) QueryStream(ctx context.Context, c *app.RequestContext) {
	uid := ragUID(ctx, c, h.jwt)
	if uid == 0 {
		response.Error(ctx, c, errcode.Unauthorized)
		return
	}

	body, err := rag.ParseQueryBody(c.Request.Body())
	if err != nil {
		response.Error(ctx, c, errcode.InvalidParam)
		return
	}
	question, err := rag.ResolveQuestion(body)
	if err != nil {
		response.Error(ctx, c, errcode.WithMessage(errcode.InvalidParam, err.Error()))
		return
	}
	history := rag.ExtractChatHistory(body)
	historyTurns := make([]rag.ChatTurn, len(history))
	copy(historyTurns, history)

	started := time.Now()
	ids := rag.CreateUiMessageStreamIDs()

	if err := h.mod.Query.AssertEnabled(); err != nil {
		writeRagHTTPError(ctx, c, err)
		return
	}

	if err := h.mod.Quota.AssertQuota(ctx, uid); err != nil {
		h.mod.Query.SaveQueryLog(ctx, rag.QueryLogParams{
			UID: uid, Question: question, Status: "quota_exceeded",
			LatencyMs: int(time.Since(started).Milliseconds()),
		})
		writeRagHTTPError(ctx, c, err)
		return
	}

	messages, citations, err := h.mod.Query.PrepareQuery(ctx, question, uid, historyTurns)
	if err != nil {
		h.mod.Query.SaveQueryLog(ctx, rag.QueryLogParams{
			UID: uid, Question: question, Status: "failed",
			AnswerPreview: err.Error(), LatencyMs: int(time.Since(started).Milliseconds()),
		})
		writeRagHTTPError(ctx, c, err)
		return
	}

	c.Response.Header.Set("Content-Type", "text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")
	c.Response.Header.Set("x-vercel-ai-ui-message-stream", "v1")
	c.SetStatusCode(http.StatusOK)

	writeSSE(c, map[string]interface{}{"type": "start", "messageId": ids.MessageID})
	writeSSE(c, map[string]interface{}{"type": "data-citations", "data": map[string]interface{}{"citations": citations}})
	writeSSE(c, map[string]interface{}{"type": "text-start", "id": ids.TextID})

	answer := ""
	streamStarted := false
	deltas, errCh := h.mod.Query.StreamChat(ctx, messages)

streamLoop:
	for {
		select {
		case delta, ok := <-deltas:
			if !ok {
				break streamLoop
			}
			if !streamStarted {
				streamStarted = true
				_ = h.mod.Quota.Consume(ctx, uid)
			}
			answer += delta
			writeSSE(c, map[string]interface{}{"type": "text-delta", "id": ids.TextID, "delta": delta})
		case err := <-errCh:
			if err != nil && !streamStarted {
				writeRagHTTPError(ctx, c, errcode.WithMessage(errcode.InternalError, "AI 回答失败"))
				h.mod.Query.SaveQueryLog(ctx, rag.QueryLogParams{
					UID: uid, Question: question, Citations: citations, Status: "failed",
					AnswerPreview: err.Error(), LatencyMs: int(time.Since(started).Milliseconds()),
				})
				return
			}
			if err != nil && streamStarted {
				writeSSE(c, map[string]interface{}{"type": "error", "errorText": err.Error()})
			}
			break streamLoop
		}
	}

	writeSSE(c, map[string]interface{}{"type": "text-end", "id": ids.TextID})
	writeSSE(c, map[string]interface{}{"type": "finish"})
	c.WriteString(rag.FormatUiMessageSSEDone())

	status := "failed"
	if streamStarted {
		status = "success"
	}
	h.mod.Query.SaveQueryLog(ctx, rag.QueryLogParams{
		UID: uid, Question: question, AnswerPreview: answer, Citations: citations,
		LatencyMs: int(time.Since(started).Milliseconds()), Status: status,
	})
}

// AdminStats GET /admin/rag/stats
func (h *RagHandler) AdminStats(ctx context.Context, c *app.RequestContext) {
	data, err := h.mod.Admin.GetStats(ctx)
	handleAdminResult(ctx, c, data, err)
}

// AdminQueryLogs GET /admin/rag/query-logs
func (h *RagHandler) AdminQueryLogs(ctx context.Context, c *app.RequestContext) {
	uid, _ := strconv.Atoi(string(c.Query("uid")))
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.mod.Admin.ListQueryLogs(ctx, uid, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

// AdminIndexJobs GET /admin/rag/index-jobs
func (h *RagHandler) AdminIndexJobs(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.mod.Admin.ListIndexJobs(ctx, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

// AdminChunks GET /admin/rag/chunks
func (h *RagHandler) AdminChunks(ctx context.Context, c *app.RequestContext) {
	articleID, _ := strconv.Atoi(string(c.Query("articleId")))
	sourceType := string(c.Query("sourceType"))
	page, _ := strconv.Atoi(string(c.Query("page")))
	pageSize, _ := strconv.Atoi(string(c.Query("pageSize")))
	data, err := h.mod.Admin.ListChunks(ctx, articleID, sourceType, page, pageSize)
	handleAdminResult(ctx, c, data, err)
}

// AdminReindex POST /admin/rag/reindex
func (h *RagHandler) AdminReindex(ctx context.Context, c *app.RequestContext) {
	var body struct {
		ArticleID int `json:"articleId"`
	}
	_ = c.Bind(&body)
	data, err := h.mod.Admin.TriggerReindex(ctx, body.ArticleID)
	handleAdminResult(ctx, c, data, err)
}

func ragUID(ctx context.Context, c *app.RequestContext, jwt *auth.JWTService) int {
	if uid := ctxutil.UserID(ctx); uid != 0 {
		return uid
	}
	if jwt == nil {
		return 0
	}
	authz := strings.TrimSpace(string(c.GetHeader("Authorization")))
	if authz == "" {
		return 0
	}
	token := strings.TrimPrefix(authz, "Bearer ")
	claims, err := jwt.Verify(strings.TrimSpace(token))
	if err != nil || claims == nil {
		return 0
	}
	return claims.ID
}

func writeSSE(c *app.RequestContext, payload interface{}) {
	c.WriteString(rag.FormatUiMessageSSE(payload))
}

func writeRagHTTPError(ctx context.Context, c *app.RequestContext, err error) {
	if ec, ok := err.(errcode.ErrCode); ok {
		if ec.Code() == 429 || ec.Code() == 503 {
			c.SetStatusCode(ec.Code())
		}
		response.Error(ctx, c, ec)
		return
	}
	response.FromError(ctx, c, err)
}
