//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/Jiang-Xia/blog-server-go/test/testutil"
)

func TestIntegrationPublicProfileCollectsLikesPagination(t *testing.T) {
	c := newClient(t)
	ctx := context.Background()

	for _, path := range []string{
		"/user/public/1/collects?page=1&pageSize=5",
		"/user/public/1/likes?page=1&pageSize=5",
	} {
		resp, status, err := c.GET(ctx, path, "")
		testutil.AssertOK(t, path, resp, status, err)

		var data struct {
			List []map[string]interface{} `json:"list"`
			Pagination struct {
				Total    int `json:"total"`
				Page     int `json:"page"`
				PageSize int `json:"pageSize"`
				Pages    int `json:"pages"`
			} `json:"pagination"`
		}
		if err := testutil.UnmarshalData(resp, &data); err != nil {
			t.Fatalf("%s unmarshal: %v", path, err)
		}
		if data.Pagination.Page != 1 || data.Pagination.PageSize != 5 {
			t.Fatalf("%s pagination mismatch: %+v", path, data.Pagination)
		}
		if data.List == nil {
			t.Fatalf("%s list is nil", path)
		}
		for i, item := range data.List {
			for _, key := range []string{"id", "title", "views", "likes", "createTime"} {
				if _, ok := item[key]; !ok {
					t.Fatalf("%s item[%d] missing %s", path, i, key)
				}
			}
		}
	}
}
