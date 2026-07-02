package pagination_test

import (
	"testing"

	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
)

func TestCalcNestPaginationDefaults(t *testing.T) {
	got := pagination.CalcNestPagination(100, 0, 0)
	if got.PageSize != 20 {
		t.Fatalf("pageSize default want 20, got %d", got.PageSize)
	}
	if got.Page != 1 {
		t.Fatalf("page default want 1, got %d", got.Page)
	}
	if got.Pages != 5 {
		t.Fatalf("pages want 5, got %d", got.Pages)
	}
}

func TestCalcNestPaginationExactPages(t *testing.T) {
	got := pagination.CalcNestPagination(40, 10, 2)
	if got.Total != 40 || got.Page != 2 || got.PageSize != 10 || got.Pages != 4 {
		t.Fatalf("unexpected pagination: %+v", got)
	}
}

func TestCalcNestPaginationZeroTotal(t *testing.T) {
	got := pagination.CalcNestPagination(0, 10, 1)
	if got.Pages != 0 {
		t.Fatalf("zero total pages want 0, got %d", got.Pages)
	}
}

func TestCalcNestPaginationPartialLastPage(t *testing.T) {
	got := pagination.CalcNestPagination(21, 10, 3)
	if got.Pages != 3 {
		t.Fatalf("pages want 3, got %d", got.Pages)
	}
}
