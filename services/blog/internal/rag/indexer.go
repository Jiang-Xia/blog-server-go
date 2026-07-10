package rag

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/article"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/category"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/knowledgechunk"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/crossdb"
	"go.uber.org/zap"
)

// Indexer RAG 索引器：文章分块 → embedding → knowledge_chunk。
type Indexer struct {
	cfg        *config.Config
	client     *ent.Client
	articles   *blogrepo.ArticleRepo
	cross      *crossdb.CrossDB
	embedding  *EmbeddingService
	chunk      *ChunkService
	log        *zap.Logger
	indexing   bool
	indexingMu sync.Mutex
}

// NewIndexer 构造 Indexer。
func NewIndexer(
	cfg *config.Config,
	client *ent.Client,
	articles *blogrepo.ArticleRepo,
	cross *crossdb.CrossDB,
	embedding *EmbeddingService,
	log *zap.Logger,
) *Indexer {
	return &Indexer{
		cfg: cfg, client: client, articles: articles, cross: cross,
		embedding: embedding, chunk: NewChunkService(cfg), log: log,
	}
}

// IsIndexing 是否有索引任务在跑。
func (x *Indexer) IsIndexing() bool {
	x.indexingMu.Lock()
	defer x.indexingMu.Unlock()
	return x.indexing
}

// ReindexAll 全量重建（异步）。
func (x *Indexer) ReindexAll(ctx context.Context) (*ent.RagIndexJob, error) {
	now := time.Now()
	job, err := x.client.RagIndexJob.Create().
		SetArticleID(0).SetStatus("pending").SetCreateAt(now).SetUpdateAt(now).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		_ = x.runJob(context.Background(), job.ID, func(ctx context.Context) (int, error) {
			return x.indexAllSources(ctx)
		})
	}()
	return job, nil
}

// ReindexArticle 单篇重建（同步）。
func (x *Indexer) ReindexArticle(ctx context.Context, articleID int) (*ent.RagIndexJob, error) {
	now := time.Now()
	job, err := x.client.RagIndexJob.Create().
		SetArticleID(articleID).SetStatus("pending").SetCreateAt(now).SetUpdateAt(now).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := x.runJob(ctx, job.ID, func(ctx context.Context) (int, error) {
		return x.indexOneArticle(ctx, articleID)
	}); err != nil {
		return job, err
	}
	return x.client.RagIndexJob.Get(ctx, job.ID)
}

// IndexArticleByID 事件驱动增量索引。
func (x *Indexer) IndexArticleByID(ctx context.Context, articleID int) {
	if !x.cfg.Rag.Enabled {
		return
	}
	if _, err := x.indexOneArticle(ctx, articleID); err != nil && x.log != nil {
		x.log.Warn("index article failed", zap.Int("articleId", articleID), zap.Error(err))
	}
}

// EnsureStaticPagesIndexed 启动时检查静态页 chunk，缺失则自动索引。
func (x *Indexer) EnsureStaticPagesIndexed(ctx context.Context) {
	if !x.cfg.Rag.Enabled || !x.embedding.IsAvailable() {
		return
	}
	for _, def := range RAGStaticPages {
		sourceKey := RagPageSourceKey(def.Slug)
		n, err := x.client.KnowledgeChunk.Query().
			Where(
				knowledgechunk.SourceTypeEQ(SourcePage),
				knowledgechunk.SourceKeyEQ(sourceKey),
				knowledgechunk.StatusEQ("active"),
			).
			Count(ctx)
		if err != nil || n > 0 {
			continue
		}
		defCopy := def
		go func() {
			if _, err := x.indexStaticPage(context.Background(), defCopy); err != nil && x.log != nil {
				x.log.Warn("static page auto index failed", zap.String("slug", defCopy.Slug), zap.Error(err))
			}
		}()
	}
}

// RemoveArticleChunks 下架/删除时软删 chunk。
func (x *Indexer) RemoveArticleChunks(ctx context.Context, articleID int) error {
	sourceKey := RagArticleSourceKey(articleID)
	_, err := x.client.KnowledgeChunk.Update().
		Where(
			knowledgechunk.SourceTypeEQ(SourceArticle),
			knowledgechunk.SourceKeyEQ(sourceKey),
		).
		SetStatus("deleted").
		SetUpdateAt(time.Now()).
		Save(ctx)
	return err
}

func (x *Indexer) runJob(ctx context.Context, jobID int, fn func(context.Context) (int, error)) error {
	x.indexingMu.Lock()
	if x.indexing {
		x.indexingMu.Unlock()
		msg := "已有索引任务正在运行"
		_, _ = x.client.RagIndexJob.UpdateOneID(jobID).
			SetStatus("failed").SetNillableErrorMsg(&msg).SetUpdateAt(time.Now()).Save(ctx)
		return fmt.Errorf("已有索引任务正在运行")
	}
	x.indexing = true
	x.indexingMu.Unlock()

	defer func() {
		x.indexingMu.Lock()
		x.indexing = false
		x.indexingMu.Unlock()
	}()

	_, _ = x.client.RagIndexJob.UpdateOneID(jobID).
		SetStatus("running").SetUpdateAt(time.Now()).Save(ctx)

	count, err := fn(ctx)
	if err != nil {
		msg := err.Error()
		if len([]rune(msg)) > 2000 {
			msg = string([]rune(msg)[:2000])
		}
		_, _ = x.client.RagIndexJob.UpdateOneID(jobID).
			SetStatus("failed").SetNillableErrorMsg(&msg).SetUpdateAt(time.Now()).Save(ctx)
		return err
	}
	_, _ = x.client.RagIndexJob.UpdateOneID(jobID).
		SetStatus("success").SetChunkCount(count).SetUpdateAt(time.Now()).Save(ctx)
	return nil
}

func (x *Indexer) indexAllSources(ctx context.Context) (int, error) {
	total := 0
	n, err := x.indexAllArticles(ctx)
	if err != nil {
		return total, err
	}
	total += n
	for _, def := range RAGStaticPages {
		c, err := x.indexStaticPage(ctx, def)
		if err != nil {
			return total, err
		}
		total += c
	}
	return total, nil
}

func (x *Indexer) indexAllArticles(ctx context.Context) (int, error) {
	ids, err := x.loadPublishedArticleIDs(ctx)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, id := range ids {
		n, err := x.indexOneArticle(ctx, id)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func (x *Indexer) indexOneArticle(ctx context.Context, articleID int) (int, error) {
	art, categoryLabel, tagLabels, err := x.loadPublishedArticle(ctx, articleID)
	if err != nil {
		return 0, err
	}
	sourceKey := RagArticleSourceKey(articleID)
	if art == nil {
		_ = x.RemoveArticleChunks(ctx, articleID)
		return 0, nil
	}

	_, err = x.client.KnowledgeChunk.Delete().
		Where(knowledgechunk.SourceTypeEQ(SourceArticle), knowledgechunk.SourceKeyEQ(sourceKey)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	pieces := x.chunk.SplitMarkdown(art.Content, art.Title, art.Description)
	if len(pieces) == 0 {
		return 0, nil
	}

	embedTexts := BuildEmbedTextsFromChunks(art.Title, pieces, art.Description, categoryLabel, "博客文章", tagLabels)
	vectors, err := x.embedding.Embed(ctx, embedTexts)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	url := fmt.Sprintf("/detail/%d", art.ID)
	bulk := make([]*ent.KnowledgeChunkCreate, len(pieces))
	for i, piece := range pieces {
		searchText := BuildRagSearchText(art.Title, piece.Content, categoryLabel, piece.HeadingPath, tagLabels)
		vec := []float64{}
		if i < len(vectors) {
			vec = vectors[i]
		}
		ct := piece.ContentType
		if ct == "" {
			ct = "prose"
		}
		create := x.client.KnowledgeChunk.Create().
			SetArticleID(art.ID).
			SetSourceType(SourceArticle).
			SetSourceKey(sourceKey).
			SetChunkIndex(piece.ChunkIndex).
			SetTitle(art.Title).
			SetContent(piece.Content).
			SetURL(url).
			SetTags(tagLabels).
			SetContentType(ct).
			SetEmbeddingJSON(vec).
			SetStatus("active").
			SetIndexedAt(now).
			SetCreateAt(now).
			SetUpdateAt(now).
			SetNillableSearchText(&searchText)
		if categoryLabel != "" {
			create.SetNillableCategory(&categoryLabel)
		}
		if piece.HeadingPath != "" {
			hp := piece.HeadingPath
			create.SetNillableHeadingPath(&hp)
		}
		bulk[i] = create
	}
	if err := x.client.KnowledgeChunk.CreateBulk(bulk...).Exec(ctx); err != nil {
		return 0, err
	}
	return len(pieces), nil
}

func (x *Indexer) indexStaticPage(ctx context.Context, def StaticPageDef) (int, error) {
	sourceKey := RagPageSourceKey(def.Slug)
	_, err := x.client.KnowledgeChunk.Delete().
		Where(knowledgechunk.SourceTypeEQ(SourcePage), knowledgechunk.SourceKeyEQ(sourceKey)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	markdown, err := LoadStaticPageMarkdown(def)
	if err != nil {
		return 0, err
	}
	pieces := x.chunk.SplitMarkdown(markdown, def.Title, def.Description)
	if len(pieces) == 0 {
		return 0, nil
	}

	embedTexts := BuildEmbedTextsFromChunks(def.Title, pieces, def.Description, def.Category, def.SourceLabel, def.Tags)
	vectors, err := x.embedding.Embed(ctx, embedTexts)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	bulk := make([]*ent.KnowledgeChunkCreate, len(pieces))
	for i, piece := range pieces {
		searchText := BuildRagSearchText(def.Title, piece.Content, def.Category, piece.HeadingPath, def.Tags)
		vec := []float64{}
		if i < len(vectors) {
			vec = vectors[i]
		}
		ct := piece.ContentType
		if ct == "" {
			ct = "prose"
		}
		cat := def.Category
		create := x.client.KnowledgeChunk.Create().
			SetArticleID(0).
			SetSourceType(SourcePage).
			SetSourceKey(sourceKey).
			SetChunkIndex(piece.ChunkIndex).
			SetTitle(def.Title).
			SetContent(piece.Content).
			SetURL(def.URL).
			SetTags(def.Tags).
			SetContentType(ct).
			SetEmbeddingJSON(vec).
			SetStatus("active").
			SetIndexedAt(now).
			SetCreateAt(now).
			SetUpdateAt(now).
			SetNillableSearchText(&searchText)
		if cat != "" {
			create.SetNillableCategory(&cat)
		}
		if piece.HeadingPath != "" {
			hp := piece.HeadingPath
			create.SetNillableHeadingPath(&hp)
		}
		bulk[i] = create
	}
	if err := x.client.KnowledgeChunk.CreateBulk(bulk...).Exec(ctx); err != nil {
		return 0, err
	}
	return len(pieces), nil
}

type articleRow struct {
	ID          int
	Title       string
	Content     string
	Description string
}

func (x *Indexer) loadPublishedArticle(ctx context.Context, articleID int) (*articleRow, string, []string, error) {
	art, err := x.client.Article.Query().
		Where(
			article.IDEQ(articleID),
			article.IsDeleteEQ(false),
			article.StatusEQ("publish"),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", nil, nil
		}
		return nil, "", nil, err
	}
	ok, err := x.cross.IsUserActive(ctx, art.UID)
	if err != nil {
		return nil, "", nil, err
	}
	if !ok {
		return nil, "", nil, nil
	}

	categoryLabel := ""
	if art.Articles != nil && *art.Articles != "" {
		cat, err := x.client.Category.Query().Where(category.IDEQ(*art.Articles)).Only(ctx)
		if err == nil {
			categoryLabel = cat.Label
		}
	}
	tagLabels, err := x.cross.ArticleTagLabels(ctx, articleID)
	if err != nil {
		return nil, "", nil, err
	}
	return &articleRow{
		ID: art.ID, Title: art.Title, Content: art.Content, Description: art.Description,
	}, categoryLabel, tagLabels, nil
}

func (x *Indexer) loadPublishedArticleIDs(ctx context.Context) ([]int, error) {
	arts, err := x.client.Article.Query().
		Where(article.IsDeleteEQ(false), article.StatusEQ("publish")).
		Select(article.FieldID, article.FieldUID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	var ids []int
	for _, a := range arts {
		ok, err := x.cross.IsUserActive(ctx, a.UID)
		if err != nil || !ok {
			continue
		}
		ids = append(ids, a.ID)
	}
	return ids, nil
}

// PurgeAuthorArticles 用户禁用时软删其全部文章 chunk。
func (x *Indexer) PurgeAuthorArticles(ctx context.Context, authorUID int) (int, error) {
	arts, err := x.client.Article.Query().
		Where(article.UIDEQ(authorUID), article.IsDeleteEQ(false)).
		Select(article.FieldID).
		All(ctx)
	if err != nil {
		return 0, err
	}
	if len(arts) == 0 {
		return 0, nil
	}
	ids := make([]int, len(arts))
	for i, a := range arts {
		ids[i] = a.ID
	}
	n, err := x.client.KnowledgeChunk.Update().
		Where(knowledgechunk.SourceTypeEQ(SourceArticle), knowledgechunk.ArticleIDIn(ids...)).
		SetStatus("deleted").
		SetUpdateAt(time.Now()).
		Save(ctx)
	return n, err
}
