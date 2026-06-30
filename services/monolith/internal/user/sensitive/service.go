// Package sensitive 敏感词 CRUD 与内容过滤实现。
package sensitive

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/sensitiveword"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/sensitivewordhit"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"go.uber.org/zap"
)

const cacheTTL = 5 * time.Minute

// Service 敏感词业务逻辑；实现 FilterService。
type Service struct {
	client *ent.Client
	log    *zap.Logger

	cacheMu        sync.RWMutex
	cachedWords    []string
	cacheUpdatedAt time.Time
}

// NewService 构造敏感词服务。
func NewService(client *ent.Client, log *zap.Logger) *Service {
	return &Service{client: client, log: log}
}

// --- FilterService ---

// EvaluateContent 分级检测：替换/拒绝/审核策略 + HP 惩罚。
func (s *Service) EvaluateContent(ctx context.Context, content string) (*EvaluateResult, error) {
	hits, err := s.checkContentDetailed(ctx, content)
	if err != nil {
		return nil, err
	}
	if len(hits) == 0 {
		return &EvaluateResult{Content: content}, nil
	}

	hitWords := make([]string, len(hits))
	maxPenalty := 0
	needReview := false
	rejected := false
	for i, h := range hits {
		hitWords[i] = h.Word
		if h.HpPenalty > maxPenalty {
			maxPenalty = h.HpPenalty
		}
		if h.NeedReview == 1 {
			needReview = true
		}
		if h.Action == 2 {
			rejected = true
		}
	}

	processed := content
	positiveIndex := 0
	for _, hit := range hits {
		if hit.Action != 1 {
			continue
		}
		escaped := regexp.QuoteMeta(hit.Word)
		var replacement string
		if hit.Level == 2 {
			replacement = positiveWordReplacements[positiveIndex%len(positiveWordReplacements)]
			positiveIndex++
		} else {
			maskLen := len([]rune(hit.Word))
			if maskLen > 6 {
				maskLen = 6
			}
			replacement = strings.Repeat("*", maskLen)
		}
		re := regexp.MustCompile("(?i)" + escaped)
		processed = re.ReplaceAllString(processed, replacement)
	}

	return &EvaluateResult{
		Content:    processed,
		Hits:       hits,
		HitWords:   hitWords,
		HpPenalty:  maxPenalty,
		NeedReview: needReview,
		Rejected:   rejected,
	}, nil
}

// CreateHitRecord 写入敏感词命中记录。
func (s *Service) CreateHitRecord(ctx context.Context, params CreateHitParams) error {
	b := s.client.SensitiveWordHit.Create().
		SetSourceType(params.SourceType).
		SetSourceId(params.SourceID).
		SetContent(params.Content).
		SetHitWords(strings.Join(params.HitWords, ",")).
		SetStatus("pending")
	if params.UID != nil {
		b.SetUID(*params.UID)
	}
	if params.IP != nil {
		b.SetIP(*params.IP)
	}
	_, err := b.Save(ctx)
	return err
}

func (s *Service) checkContentDetailed(ctx context.Context, content string) ([]HitDetail, error) {
	normalized := strings.ToLower(strings.TrimSpace(content))
	words, err := s.client.SensitiveWord.Query().
		Where(sensitiveword.StatusEQ(1)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	var hits []HitDetail
	for _, w := range words {
		if strings.Contains(normalized, strings.ToLower(w.Word)) {
			hits = append(hits, HitDetail{
				Word:       w.Word,
				Level:      w.Level,
				HpPenalty:  w.HpPenalty,
				NeedReview: w.NeedReview,
				Action:     w.Action,
			})
		}
	}
	return hits, nil
}

func (s *Service) refreshCache(ctx context.Context) error {
	words, err := s.client.SensitiveWord.Query().
		Where(sensitiveword.StatusEQ(1)).
		Select(sensitiveword.FieldWord).
		All(ctx)
	if err != nil {
		return err
	}
	lowered := make([]string, len(words))
	for i, w := range words {
		lowered[i] = strings.ToLower(w.Word)
	}
	s.cacheMu.Lock()
	s.cachedWords = lowered
	s.cacheUpdatedAt = time.Now()
	s.cacheMu.Unlock()
	return nil
}

// --- CRUD ---

// ListQuery 敏感词列表查询参数。
type ListQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Category string
	Status   *int
}

// List 分页查询敏感词。
func (s *Service) List(ctx context.Context, q ListQuery) (map[string]interface{}, error) {
	page, pageSize := q.Page, q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	query := s.client.SensitiveWord.Query()
	if q.Keyword != "" {
		query = query.Where(sensitiveword.WordContains(q.Keyword))
	}
	if q.Category != "" {
		query = query.Where(sensitiveword.CategoryEQ(q.Category))
	}
	if q.Status != nil {
		query = query.Where(sensitiveword.StatusEQ(*q.Status))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	list, err := query.
		Order(ent.Desc(sensitiveword.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": repo.CalcNestPagination(total, pageSize, page),
	}, nil
}

// Create 新增敏感词。
func (s *Service) Create(ctx context.Context, data map[string]interface{}) (*ent.SensitiveWord, error) {
	word := strVal(data, "word")
	if word == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "敏感词不能为空")
	}
	exists, err := s.client.SensitiveWord.Query().
		Where(sensitiveword.WordEQ(word)).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.WithMessage(errcode.InvalidParam, "该敏感词已存在")
	}
	b := s.client.SensitiveWord.Create().SetWord(word)
	if v := strVal(data, "category"); v != "" {
		b.SetCategory(v)
	}
	if v, ok := intVal(data, "status"); ok {
		b.SetStatus(v)
	}
	if v, ok := intVal(data, "level"); ok {
		b.SetLevel(v)
	}
	if v, ok := intVal(data, "hpPenalty"); ok {
		b.SetHpPenalty(v)
	}
	if v, ok := intVal(data, "needReview"); ok {
		b.SetNeedReview(v)
	}
	if v, ok := intVal(data, "action"); ok {
		b.SetAction(v)
	}
	saved, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.refreshCache(ctx)
	return saved, nil
}

// BatchCreate 批量导入敏感词。
func (s *Service) BatchCreate(ctx context.Context, items []map[string]string) ([]*ent.SensitiveWord, error) {
	if len(items) == 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "请提供要批量添加的敏感词")
	}
	builders := make([]*ent.SensitiveWordCreate, 0, len(items))
	for _, item := range items {
		word := strings.TrimSpace(item["word"])
		if word == "" {
			continue
		}
		cat := item["category"]
		if cat == "" {
			cat = "自定义"
		}
		builders = append(builders, s.client.SensitiveWord.Create().
			SetWord(word).SetCategory(cat).SetStatus(1))
	}
	if len(builders) == 0 {
		return nil, errcode.WithMessage(errcode.InvalidParam, "请提供要批量添加的敏感词")
	}
	saved, err := s.client.SensitiveWord.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.refreshCache(ctx)
	return saved, nil
}

// Update 更新敏感词。
func (s *Service) Update(ctx context.Context, id int, data map[string]interface{}) (*ent.SensitiveWord, error) {
	_, err := s.client.SensitiveWord.Query().
		Where(sensitiveword.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "敏感词不存在")
		}
		return nil, err
	}
	b := s.client.SensitiveWord.UpdateOneID(id)
	if v := strVal(data, "word"); v != "" {
		b.SetWord(v)
	}
	if v := strVal(data, "category"); v != "" {
		b.SetCategory(v)
	}
	if v, ok := intVal(data, "status"); ok {
		b.SetStatus(v)
	}
	if v, ok := intVal(data, "level"); ok {
		b.SetLevel(v)
	}
	if v, ok := intVal(data, "hpPenalty"); ok {
		b.SetHpPenalty(v)
	}
	if v, ok := intVal(data, "needReview"); ok {
		b.SetNeedReview(v)
	}
	if v, ok := intVal(data, "action"); ok {
		b.SetAction(v)
	}
	saved, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	_ = s.refreshCache(ctx)
	return saved, nil
}

// Delete 删除敏感词。
func (s *Service) Delete(ctx context.Context, id int) error {
	_, err := s.client.SensitiveWord.Query().
		Where(sensitiveword.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "敏感词不存在")
		}
		return err
	}
	if err := s.client.SensitiveWord.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}
	_ = s.refreshCache(ctx)
	return nil
}

// HitListQuery 命中记录查询参数。
type HitListQuery struct {
	Page       int
	PageSize   int
	SourceType string
	Status     string
}

// ListHits 分页查询命中记录。
func (s *Service) ListHits(ctx context.Context, q HitListQuery) (map[string]interface{}, error) {
	page, pageSize := q.Page, q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	query := s.client.SensitiveWordHit.Query()
	if q.SourceType != "" {
		query = query.Where(sensitivewordhit.SourceTypeEQ(q.SourceType))
	}
	if q.Status != "" {
		query = query.Where(sensitivewordhit.StatusEQ(q.Status))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	list, err := query.
		Order(ent.Desc(sensitivewordhit.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": repo.CalcNestPagination(total, pageSize, page),
	}, nil
}

// Approve 审核通过命中记录（来源实体状态同步留待 Plan 06/07）。
func (s *Service) Approve(ctx context.Context, hitID, reviewerID int) (*ent.SensitiveWordHit, error) {
	return s.reviewHit(ctx, hitID, reviewerID, "approved")
}

// Reject 审核拒绝命中记录。
func (s *Service) Reject(ctx context.Context, hitID, reviewerID int) (*ent.SensitiveWordHit, error) {
	return s.reviewHit(ctx, hitID, reviewerID, "rejected")
}

func (s *Service) reviewHit(ctx context.Context, hitID, reviewerID int, status string) (*ent.SensitiveWordHit, error) {
	hit, err := s.client.SensitiveWordHit.Query().
		Where(sensitivewordhit.IDEQ(hitID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "命中记录不存在")
		}
		return nil, err
	}
	if hit.Status != "pending" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "该记录已审核")
	}
	now := time.Now()
	saved, err := s.client.SensitiveWordHit.UpdateOneID(hitID).
		SetStatus(status).
		SetReviewerId(reviewerID).
		SetReviewTime(now).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	s.log.Info("敏感词命中审核完成",
		zap.Int("hitId", hitID),
		zap.String("status", status),
		zap.String("sourceType", hit.SourceType),
		zap.String("sourceId", hit.SourceId),
	)
	return saved, nil
}

func strVal(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func intVal(m map[string]interface{}, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return int(t), true
	case int:
		return t, true
	case int64:
		return int(t), true
	default:
		return 0, false
	}
}
