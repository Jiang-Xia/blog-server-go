package publicprofile

import (
	"testing"
	"time"
)

func TestBuildListResultFiltersEmptyList(t *testing.T) {
	res := BuildListResult(nil, 0, 1, 10)
	if len(res.List) != 0 {
		t.Fatalf("expected empty list, got %d", len(res.List))
	}
	if res.Pagination.Total != 0 || res.Pagination.Page != 1 || res.Pagination.PageSize != 10 {
		t.Fatalf("unexpected pagination: %+v", res.Pagination)
	}
}

func TestMapRowMatchesNestFields(t *testing.T) {
	ct := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	m := mapRow(ArticleRow{
		ID: 7, Title: "t", Description: "d", Cover: "c",
		Views: 10, Likes: 3, ArticleLevel: 2, IsMasterpiece: 1, TipTotal: 5,
		CreateTime: ct,
	})
	for _, key := range []string{"id", "title", "description", "cover", "views", "likes", "articleLevel", "isMasterpiece", "tipTotal", "createTime"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing field %s", key)
		}
	}
	if m["id"] != 7 || m["likes"] != 3 {
		t.Fatalf("unexpected map: %+v", m)
	}
}

func TestNormalizePageDefaults(t *testing.T) {
	p, ps := normalizePage(0, 0)
	if p != 1 || ps != 10 {
		t.Fatalf("got page=%d pageSize=%d", p, ps)
	}
}

func TestAuthorStatusActiveConstant(t *testing.T) {
	if authorStatusActive != "active" {
		t.Fatalf("author status must be active for Nest parity, got %q", authorStatusActive)
	}
}
