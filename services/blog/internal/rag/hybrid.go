package rag

import (
	"context"
	"math"
	"strings"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/knowledgechunk"
)

// SearchHit 混合检索命中项。
type SearchHit struct {
	ID           int
	ArticleID    int
	SourceType   string
	SourceKey    string
	Title        string
	Content      string
	URL          string
	HeadingPath  string
	Score        float64
	VectorScore  float64
	KeywordScore float64
}

// Citation 引用来源，随 SSE 下发。
type Citation struct {
	ArticleID  int    `json:"articleId"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Snippet    string `json:"snippet"`
	SourceType string `json:"sourceType,omitempty"`
}

// HybridSearch 向量 + 关键词混合检索。
type HybridSearch struct {
	client *ent.Client
}

// NewHybridSearch 构造 HybridSearch。
func NewHybridSearch(client *ent.Client) *HybridSearch {
	return &HybridSearch{client: client}
}

type scoredHit struct {
	hit SearchHit
}

type kwHit struct {
	id    int
	score float64
}

// Search 混合检索主入口。
func (h *HybridSearch) Search(ctx context.Context, cfg *config.Config, query string, queryVector []float64, topK int) ([]SearchHit, error) {
	if topK <= 0 {
		topK = cfg.Rag.RagTopKOrDefault()
	}
	poolSize := topK * 3
	if poolSize < HybridCandidatePool {
		poolSize = HybridCandidatePool
	}

	keywordHits, _ := h.searchByKeyword(ctx, query, poolSize)
	kwMap := make(map[int]float64, len(keywordHits))
	for _, kh := range keywordHits {
		kwMap[kh.id] = kh.score
	}

	chunks, err := h.client.KnowledgeChunk.Query().
		Where(knowledgechunk.StatusEQ("active")).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(chunks) == 0 {
		return nil, nil
	}

	var scoredList []scoredHit
	for _, chunk := range chunks {
		vecScore := cosineSimilarity(queryVector, chunk.EmbeddingJSON)
		searchText := derefStr(chunk.SearchText)
		if searchText == "" {
			searchText = chunk.Title + "\n" + chunk.Content
		}
		kwScore, ok := kwMap[chunk.ID]
		if !ok {
			kwScore = keywordScore(query, searchText)
		}
		total := vecScore*VectorScoreWeight + kwScore*KeywordScoreWeight
		if !isFinite(total) || total <= 0 {
			continue
		}
		scoredList = append(scoredList, scoredHit{hit: SearchHit{
			ID: chunk.ID, ArticleID: chunk.ArticleID, SourceType: chunk.SourceType,
			SourceKey: chunk.SourceKey, Title: chunk.Title, Content: chunk.Content,
			URL: chunk.URL, HeadingPath: derefStr(chunk.HeadingPath),
			Score: total, VectorScore: vecScore, KeywordScore: kwScore,
		}})
	}

	sortScoredHits(scoredList)
	if len(scoredList) > poolSize {
		scoredList = scoredList[:poolSize]
	}
	if len(scoredList) > topK {
		scoredList = scoredList[:topK]
	}
	out := make([]SearchHit, len(scoredList))
	for i, s := range scoredList {
		out[i] = s.hit
	}
	return out, nil
}

// CountActive 当前 active 知识块数量。
func (h *HybridSearch) CountActive(ctx context.Context) (int, error) {
	return h.client.KnowledgeChunk.Query().Where(knowledgechunk.StatusEQ("active")).Count(ctx)
}

func (h *HybridSearch) searchByKeyword(ctx context.Context, query string, limit int) ([]kwHit, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}

	chunks, err := h.client.KnowledgeChunk.Query().
		Where(knowledgechunk.StatusEQ("active")).
		Select(knowledgechunk.FieldID, knowledgechunk.FieldSearchText, knowledgechunk.FieldTitle, knowledgechunk.FieldContent).
		All(ctx)
	if err != nil {
		return nil, err
	}
	var hits []kwHit
	for _, chunk := range chunks {
		text := derefStr(chunk.SearchText)
		if text == "" {
			text = chunk.Title + "\n" + chunk.Content
		}
		score := keywordScore(q, text)
		if score > 0 {
			hits = append(hits, kwHit{id: chunk.ID, score: score})
		}
	}
	sortKwHits(hits)
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

func keywordScore(query, text string) float64 {
	q := strings.ToLower(strings.TrimSpace(query))
	t := strings.ToLower(text)
	if q == "" || t == "" {
		return 0
	}
	if strings.Contains(t, q) {
		return 1
	}
	tokens := tokenizeQuery(q)
	if len(tokens) == 0 {
		return 0
	}
	hits := 0
	for _, token := range tokens {
		if strings.Contains(t, token) {
			hits++
		}
	}
	return float64(hits) / float64(len(tokens))
}

func tokenizeQuery(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		switch r {
		case ' ', ',', '，', '。', '!', '?', '！', '？', '、', ';', '；', ':', '：':
			return true
		}
		return false
	})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len([]rune(p)) >= 2 {
			out = append(out, p)
		}
	}
	return out
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	return dot / denom
}

func sortScoredHits(list []scoredHit) {
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].hit.Score > list[i].hit.Score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

func sortKwHits(hits []kwHit) {
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[j].score > hits[i].score {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}
}

func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
