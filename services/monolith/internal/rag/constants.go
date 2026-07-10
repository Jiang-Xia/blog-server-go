// RAG 常量，对齐 Nest rag.constants.ts。
package rag

const (
	DailyQueryLimitDefault = 20
	DefaultTopK            = 6
	ChunkMinChars          = 300
	ChunkMaxChars          = 600
	ChunkOverlapChars      = 120
	VectorScoreWeight      = 0.7
	KeywordScoreWeight     = 0.3
	HybridCandidatePool    = 24
	EmbedBatchSize         = 16
	MaxQuestionChars       = 500
	MaxHistoryTurns        = 4
	MaxHistoryChars        = 3000
	QueryRedisPrefix       = "rag:query"
	SourceArticle          = "article"
	SourcePage             = "page"
)

// RagArticleSourceKey 文章 source_key。
func RagArticleSourceKey(articleID int) string {
	return "article:" + itoa(articleID)
}

// RagPageSourceKey 静态页 source_key。
func RagPageSourceKey(slug string) string {
	return "page:" + slug
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
