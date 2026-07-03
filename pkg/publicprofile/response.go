package publicprofile

import (
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
)

// ListResult 公开收藏/点赞列表 HTTP 响应体 data 段。
type ListResult struct {
	List       []map[string]interface{} `json:"list"`
	Pagination pagination.NestPagination `json:"pagination"`
}

// BuildListResult 将 repo 行映射为 Nest mapPublicArticle + getPagination 结构。
func BuildListResult(rows []ArticleRow, total, page, pageSize int) ListResult {
	list := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapRow(row))
	}
	return ListResult{
		List:       list,
		Pagination: pagination.CalcNestPagination(total, pageSize, page),
	}
}

func mapRow(row ArticleRow) map[string]interface{} {
	return map[string]interface{}{
		"id":             row.ID,
		"title":          row.Title,
		"description":    row.Description,
		"cover":          row.Cover,
		"views":          row.Views,
		"likes":          row.Likes,
		"articleLevel":   row.ArticleLevel,
		"isMasterpiece":  row.IsMasterpiece,
		"tipTotal":       row.TipTotal,
		"createTime":     row.CreateTime.UTC().Format(time.RFC3339Nano),
	}
}
