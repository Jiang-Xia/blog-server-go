// pagination Nest 风格分页元信息计算（page/pages）。
package repo

import "math"

// NestPagination 与 Nest getPagination 对齐（page/pages，非 user list 的 currentPage）。
type NestPagination struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Pages    int `json:"pages"`
}

// CalcNestPagination 计算分页元信息。
func CalcNestPagination(total, pageSize, page int) NestPagination {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	if pages == 0 && total > 0 {
		pages = 1
	}
	return NestPagination{Total: total, Page: page, PageSize: pageSize, Pages: pages}
}
