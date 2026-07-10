// Package rag RAG 知识库模块：分块、Embedding、混合检索、流式问答与索引（对齐 Nest RagModule）。
package rag

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/crossdb"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/rag/tools"
	"go.uber.org/zap"
)

// Module 聚合 RAG 子服务，供 HTTP handler 与 Stream 消费者注入。
type Module struct {
	Cfg       *config.Config
	Quota     *QuotaService
	Query     *QueryService
	Hybrid    *HybridSearch
	Embedding *EmbeddingService
	Indexer   *Indexer
	Admin     *AdminService
}

// NewModule 装配 RAG 模块依赖。
func NewModule(
	cfg *config.Config,
	client *ent.Client,
	redis *redisutil.Store,
	articles *blogrepo.ArticleRepo,
	cross *crossdb.CrossDB,
	log *zap.Logger,
) *Module {
	emb := NewEmbeddingService(cfg, log)
	hybrid := NewHybridSearch(client)
	quota := NewQuotaService(cfg, redis)
	indexer := NewIndexer(cfg, client, articles, cross, emb, log)
	toolsSvc := tools.NewService(client, cross)
	orch := tools.NewOrchestrator(toolsSvc, cfg)
	query := NewQueryService(cfg, client, emb, hybrid, orch, log)
	admin := NewAdminService(client, indexer, hybrid, emb, cfg)
	return &Module{
		Cfg:       cfg,
		Quota:     quota,
		Query:     query,
		Hybrid:    hybrid,
		Embedding: emb,
		Indexer:   indexer,
		Admin:     admin,
	}
}
