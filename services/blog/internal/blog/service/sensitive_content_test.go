package service

import (
	"context"
	"testing"

	contentfilter "github.com/Jiang-Xia/blog-server-go/services/blog/internal/contentfilter"
)

type mockContentFilter struct {
	result *contentfilter.EvaluateResult
	err    error
}

func (m *mockContentFilter) EvaluateContent(_ context.Context, _ string) (*contentfilter.EvaluateResult, error) {
	return m.result, m.err
}

func (m *mockContentFilter) CreateHitRecord(_ context.Context, _ contentfilter.CreateHitParams) error {
	return nil
}

func TestApplySensitiveFilterNeedReviewPending(t *testing.T) {
	filter := &mockContentFilter{result: &contentfilter.EvaluateResult{
		Content:    "masked",
		HitWords:   []string{"bad"},
		NeedReview: true,
	}}
	got, err := applySensitiveFilter(context.Background(), filter, "comment", "0", "bad word", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "pending" {
		t.Fatalf("status=%q want pending", got.Status)
	}
}

func TestApplySensitiveFilterRejected(t *testing.T) {
	filter := &mockContentFilter{result: &contentfilter.EvaluateResult{
		HitWords: []string{"blocked"},
		Rejected: true,
	}}
	_, err := applySensitiveFilter(context.Background(), filter, "comment", "0", "blocked", nil, nil)
	if err == nil {
		t.Fatal("expected error for rejected content")
	}
}
