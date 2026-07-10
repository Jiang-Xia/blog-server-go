package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"go.uber.org/zap"
)

// EmbeddingMode remote | local。
type EmbeddingMode string

const (
	EmbeddingRemote EmbeddingMode = "remote"
	EmbeddingLocal  EmbeddingMode = "local"
)

// EmbeddingService OpenAI 兼容 Embedding 或本地 hash 回退。
type EmbeddingService struct {
	cfg        *config.Config
	log        *zap.Logger
	mode       EmbeddingMode
	modeLogged bool
	localDim   int
	client     *http.Client
}

// NewEmbeddingService 构造 EmbeddingService。
func NewEmbeddingService(cfg *config.Config, log *zap.Logger) *EmbeddingService {
	s := &EmbeddingService{
		cfg:      cfg,
		log:      log,
		localDim: 384,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
	s.resolveMode()
	s.logModeOnce()
	return s
}

// IsAvailable 是否已配置可用。
func (s *EmbeddingService) IsAvailable() bool {
	return s.mode == EmbeddingRemote || s.cfg.Rag.AllowLocalFallback
}

// GetMode 当前 embedding 模式。
func (s *EmbeddingService) GetMode() EmbeddingMode {
	return s.mode
}

// IsRemoteConfigured 环境是否满足远程 Embedding。
func (s *EmbeddingService) IsRemoteConfigured() bool {
	return s.canUseRemoteEmbedding()
}

// Embed 批量生成 embedding。
func (s *EmbeddingService) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	if s.mode == EmbeddingRemote {
		vecs, err := s.embedViaAPI(ctx, texts)
		if err == nil {
			return vecs, nil
		}
		if s.log != nil {
			s.log.Warn("embedding API failed, fallback to local", zap.Error(err))
		}
		if s.cfg.Rag.AllowLocalFallback {
			s.mode = EmbeddingLocal
			s.logModeOnce()
		} else {
			return nil, err
		}
	}
	if s.cfg.Rag.AllowLocalFallback || s.mode == EmbeddingLocal {
		out := make([][]float64, len(texts))
		for i, t := range texts {
			out[i] = s.localEmbed(t)
		}
		return out, nil
	}
	return nil, fmt.Errorf("RAG embedding 未配置")
}

// EmbedOne 单条 embedding。
func (s *EmbeddingService) EmbedOne(ctx context.Context, text string) ([]float64, error) {
	vecs, err := s.Embed(ctx, []string{text})
	if err != nil || len(vecs) == 0 {
		return nil, err
	}
	return vecs[0], nil
}

func (s *EmbeddingService) resolveMode() {
	if s.canUseRemoteEmbedding() {
		s.mode = EmbeddingRemote
	} else {
		s.mode = EmbeddingLocal
	}
}

func (s *EmbeddingService) canUseRemoteEmbedding() bool {
	base := s.resolveEmbeddingBaseURL()
	key := s.resolveEmbeddingAPIKey()
	if key == "" {
		return false
	}
	if s.cfg.Rag.Embedding.RemoteURL == "" && isDeepSeekHost(base) {
		return false
	}
	return true
}

func (s *EmbeddingService) resolveEmbeddingAPIKey() string {
	if k := strings.TrimSpace(s.cfg.Rag.Embedding.APIKey); k != "" {
		return k
	}
	if s.cfg.Rag.Embedding.RemoteURL == "" {
		return strings.TrimSpace(s.cfg.Rag.LLM.APIKey)
	}
	return ""
}

func (s *EmbeddingService) resolveEmbeddingBaseURL() string {
	if u := strings.TrimSpace(s.cfg.Rag.Embedding.RemoteURL); u != "" {
		return u
	}
	if u := strings.TrimSpace(s.cfg.Rag.LLM.BaseURL); u != "" {
		return u
	}
	return "https://api.deepseek.com/v1"
}

func (s *EmbeddingService) embedViaAPI(ctx context.Context, texts []string) ([][]float64, error) {
	model := s.cfg.Rag.Embedding.Model
	if model == "" {
		model = "BAAI/bge-large-zh-v1.5"
	}
	base := strings.TrimSuffix(s.resolveEmbeddingBaseURL(), "/")
	url := base + "/embeddings"
	key := s.resolveEmbeddingAPIKey()

	var results [][]float64
	for i := 0; i < len(texts); i += EmbedBatchSize {
		end := i + EmbedBatchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		body, _ := json.Marshal(map[string]interface{}{
			"model": model,
			"input": batch,
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+key)
		resp, err := s.client.Do(req)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("embedding API %d: %s", resp.StatusCode, string(data))
		}
		var parsed struct {
			Data []struct {
				Index     int       `json:"index"`
				Embedding []float64 `json:"embedding"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return nil, err
		}
		sorted := make([][]float64, len(batch))
		for _, item := range parsed.Data {
			if item.Index >= 0 && item.Index < len(sorted) {
				sorted[item.Index] = item.Embedding
			}
		}
		results = append(results, sorted...)
		if end < len(texts) {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return results, nil
}

func (s *EmbeddingService) localEmbed(text string) []float64 {
	vec := make([]float64, s.localDim)
	lower := strings.ToLower(text)
	tokens := tokenizeEmbed(lower)
	if len(tokens) == 0 {
		return vec
	}
	for _, token := range tokens {
		for i := 0; i+1 < len(token); i++ {
			gram := token[i : i+2]
			idx := hashToIndex(gram, s.localDim)
			vec[idx] += 1
		}
		idx := hashToIndex(token, s.localDim)
		vec[idx] += 2
	}
	return normalizeVec(vec)
}

func tokenizeEmbed(text string) []string {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() >= 1 {
			tokens = append(tokens, cur.String())
		}
		cur.Reset()
	}
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r >= 0x4e00 && r <= 0x9fff) || r == '_' {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return tokens
}

func hashToIndex(s string, dim int) int {
	var h uint32
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return int(h % uint32(dim))
}

func normalizeVec(vec []float64) []float64 {
	var sum float64
	for _, v := range vec {
		sum += v * v
	}
	norm := math.Sqrt(sum)
	if norm == 0 {
		return vec
	}
	out := make([]float64, len(vec))
	for i, v := range vec {
		out[i] = v / norm
	}
	return out
}

func isDeepSeekHost(url string) bool {
	return strings.Contains(strings.ToLower(url), "deepseek.com")
}

func (s *EmbeddingService) logModeOnce() {
	if s.modeLogged || s.log == nil {
		return
	}
	s.modeLogged = true
	if s.mode == EmbeddingRemote {
		s.log.Info("RAG embedding: remote", zap.String("base", s.resolveEmbeddingBaseURL()), zap.String("model", s.cfg.Rag.Embedding.Model))
		return
	}
	if s.cfg.Rag.AllowLocalFallback {
		s.log.Info("RAG embedding: local hash 向量（开发用）")
	}
}
