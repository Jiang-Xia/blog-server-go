package level

import (
	"context"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/blogsvc"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/constants"
)

type mockArticleRPGStore struct {
	snap      *blogsvc.ArticleRPGFields
	updated   *blogsvc.ArticleRPGFields
	tipAdds   []int
	updateErr error
}

func (m *mockArticleRPGStore) GetArticleRPGFields(context.Context, int) (*blogsvc.ArticleRPGFields, error) {
	if m.snap == nil {
		return nil, nil
	}
	copy := *m.snap
	return &copy, nil
}

func (m *mockArticleRPGStore) UpdateArticleRPGFields(_ context.Context, articleID int, exp, level, repGained, isMasterpiece int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated = &blogsvc.ArticleRPGFields{
		ArticleID:        articleID,
		ArticleExp:       exp,
		ArticleLevel:     level,
		ReputationGained: repGained,
		IsMasterpiece:    isMasterpiece,
	}
	if m.snap != nil {
		m.snap.ArticleExp = exp
		m.snap.ArticleLevel = level
		m.snap.ReputationGained = repGained
		m.snap.IsMasterpiece = isMasterpiece
	}
	return nil
}

func (m *mockArticleRPGStore) AddArticleTipTotal(_ context.Context, _ int, amount int) error {
	m.tipAdds = append(m.tipAdds, amount)
	return nil
}

func TestArticleLevelThreshold(t *testing.T) {
	cases := []struct {
		level int
		want  int
	}{
		{1, 0},
		{2, 40},
		{3, 120},
		{10, 1800},
	}
	for _, c := range cases {
		if got := articleLevelThreshold(c.level); got != c.want {
			t.Fatalf("level %d: got %d want %d", c.level, got, c.want)
		}
	}
}

func TestCheckMasterpiece(t *testing.T) {
	if !checkMasterpiece(rpgconst.Economy.MasterpieceLevel, 0) {
		t.Fatal("level threshold should be masterpiece")
	}
	if !checkMasterpiece(1, rpgconst.Economy.MasterpieceExp) {
		t.Fatal("exp threshold should be masterpiece")
	}
	if checkMasterpiece(9, rpgconst.Economy.MasterpieceExp-1) {
		t.Fatal("below thresholds should not be masterpiece")
	}
}

func TestAddArticleExpLevelUpAndMasterpiece(t *testing.T) {
	store := &mockArticleRPGStore{snap: &blogsvc.ArticleRPGFields{
		ArticleID: 1, AuthorUID: 9, ArticleLevel: 1, ArticleExp: 0,
	}}
	svc := NewArticleLevelService(store, nil, nil)

	res, err := svc.AddArticleExp(context.Background(), 1, 1000, 9, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.LeveledUp {
		t.Fatal("expected level up")
	}
	if !res.IsMasterpiece {
		t.Fatal("expected masterpiece")
	}
	if store.updated == nil || store.updated.ArticleExp != 1000 {
		t.Fatalf("unexpected update: %+v", store.updated)
	}
	if store.updated.ArticleLevel < 2 {
		t.Fatalf("expected higher level, got %d", store.updated.ArticleLevel)
	}
	if store.updated.IsMasterpiece != 1 {
		t.Fatal("isMasterpiece should be 1")
	}
}

func TestAddArticleExpPublishSkipsReputation(t *testing.T) {
	store := &mockArticleRPGStore{snap: &blogsvc.ArticleRPGFields{
		ArticleID: 2, AuthorUID: 3, ArticleLevel: 1,
	}}
	svc := NewArticleLevelService(store, nil, nil)
	_, err := svc.AddArticleExp(context.Background(), 2, 10, 3, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if store.updated.ReputationGained != 10 {
		t.Fatalf("reputationGained=%d want 10", store.updated.ReputationGained)
	}
}

func TestAddTipTotal(t *testing.T) {
	store := &mockArticleRPGStore{}
	svc := NewArticleLevelService(store, nil, nil)
	if err := svc.AddTipTotal(context.Background(), 5, 20); err != nil {
		t.Fatal(err)
	}
	if len(store.tipAdds) != 1 || store.tipAdds[0] != 20 {
		t.Fatalf("tip adds: %v", store.tipAdds)
	}
}
