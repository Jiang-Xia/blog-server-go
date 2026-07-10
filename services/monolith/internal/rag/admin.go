package rag

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/knowledgechunk"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/ragindexjob"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/ragquerylog"
)

// AdminService RAG 管理端统计与列表。
type AdminService struct {
	client    *ent.Client
	indexer   *Indexer
	hybrid    *HybridSearch
	embedding *EmbeddingService
	cfg       *config.Config
}

// NewAdminService 构造 AdminService。
func NewAdminService(client *ent.Client, indexer *Indexer, hybrid *HybridSearch, embedding *EmbeddingService, cfg *config.Config) *AdminService {
	return &AdminService{client: client, indexer: indexer, hybrid: hybrid, embedding: embedding, cfg: cfg}
}

// GetStats 概览统计。
func (a *AdminService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -7)

	todayQueries, _ := a.client.RagQueryLog.Query().
		Where(ragquerylog.CreateAtGTE(todayStart)).Count(ctx)
	weekQueries, _ := a.client.RagQueryLog.Query().
		Where(ragquerylog.CreateAtGTE(weekStart)).Count(ctx)
	successToday, _ := a.client.RagQueryLog.Query().
		Where(ragquerylog.CreateAtGTE(todayStart), ragquerylog.StatusEQ("success")).Count(ctx)
	quotaToday, _ := a.client.RagQueryLog.Query().
		Where(ragquerylog.CreateAtGTE(todayStart), ragquerylog.StatusEQ("quota_exceeded")).Count(ctx)
	chunkCount, _ := a.hybrid.CountActive(ctx)

	lastJob, _ := a.client.RagIndexJob.Query().
		Order(ent.Desc(ragindexjob.FieldCreateAt)).First(ctx)

	out := map[string]interface{}{
		"todayQueries":              todayQueries,
		"weekQueries":               weekQueries,
		"successToday":              successToday,
		"quotaExceededToday":        quotaToday,
		"chunkCount":                chunkCount,
		"indexing":                  a.indexer.IsIndexing(),
		"embeddingMode":             a.embedding.GetMode(),
		"embeddingRemoteConfigured": a.embedding.IsRemoteConfigured(),
		"embeddingModel":            a.cfg.Rag.Embedding.Model,
		"lastJob":                   lastJob,
		"topUsers":                  []interface{}{},
		"dailyTrend":                []interface{}{},
	}
	return out, nil
}

// ListQueryLogs 查询日志分页。
func (a *AdminService) ListQueryLogs(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := a.client.RagQueryLog.Query()
	if uid > 0 {
		q = q.Where(ragquerylog.UIDEQ(uid))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	list, err := q.Order(ent.Desc(ragquerylog.FieldCreateAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// ListIndexJobs 索引任务分页。
func (a *AdminService) ListIndexJobs(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	total, err := a.client.RagIndexJob.Query().Count(ctx)
	if err != nil {
		return nil, err
	}
	list, err := a.client.RagIndexJob.Query().
		Order(ent.Desc(ragindexjob.FieldCreateAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// ListChunks 知识块列表。
func (a *AdminService) ListChunks(ctx context.Context, articleID int, sourceType string, page, pageSize int) (map[string]interface{}, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := a.client.KnowledgeChunk.Query().Where(knowledgechunk.StatusEQ("active"))
	if articleID > 0 {
		q = q.Where(knowledgechunk.ArticleIDEQ(articleID))
	}
	if sourceType != "" {
		q = q.Where(knowledgechunk.SourceTypeEQ(sourceType))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	list, err := q.
		Select(
			knowledgechunk.FieldID, knowledgechunk.FieldArticleID, knowledgechunk.FieldSourceType,
			knowledgechunk.FieldSourceKey, knowledgechunk.FieldChunkIndex, knowledgechunk.FieldTitle,
			knowledgechunk.FieldURL, knowledgechunk.FieldCategory, knowledgechunk.FieldHeadingPath,
			knowledgechunk.FieldStatus, knowledgechunk.FieldIndexedAt,
		).
		Order(ent.Desc(knowledgechunk.FieldArticleID), ent.Asc(knowledgechunk.FieldChunkIndex)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// TriggerReindex 触发全量或单篇重建。
func (a *AdminService) TriggerReindex(ctx context.Context, articleID int) (map[string]interface{}, error) {
	if a.indexer.IsIndexing() {
		return nil, errcode.WithMessage(errcode.Conflict, "索引任务进行中，请稍后再试")
	}
	var job *ent.RagIndexJob
	var err error
	if articleID > 0 {
		job, err = a.indexer.ReindexArticle(ctx, articleID)
	} else {
		job, err = a.indexer.ReindexAll(ctx)
	}
	if err != nil {
		return nil, err
	}
	msg := "全量索引任务已提交，请在本页查看进度"
	if articleID > 0 {
		msg = "文章 " + itoa(articleID) + " 索引已完成"
	}
	return map[string]interface{}{
		"job":           job,
		"embeddingMode": a.embedding.GetMode(),
		"indexing":      a.indexer.IsIndexing(),
		"message":       msg,
	}, nil
}
