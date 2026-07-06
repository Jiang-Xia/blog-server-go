// Package domain 定义 blog 域 DTO 与查询参数，对齐 Nest article/category/tag 接口。
package domain

import (
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
)

// CategoryItem 分类摘要（列表/详情嵌套）。
type CategoryItem struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Value        string `json:"value,omitempty"`
	Color        string `json:"color,omitempty"`
	ArticleCount int    `json:"articleCount,omitempty"`
}

// TagItem 标签摘要。
type TagItem struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Value        string `json:"value,omitempty"`
	Color        string `json:"color,omitempty"`
	ArticleCount int    `json:"articleCount,omitempty"`
}

// UserInfoItem 作者摘要（由 UserService 组装，禁止跨表 JOIN user）。
type UserInfoItem struct {
	ID       uint64 `json:"id"`
	Nickname string `json:"nickname"`
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar"`
}

// ArticleListQuery 文章列表筛选，对齐 Nest ListDTO。
type ArticleListQuery struct {
	Page        int
	PageSize    int
	Category    string
	Tags        []string
	Title       string
	Description string
	Content     string
	Sort        string
	Client      bool
	Admin       bool
	DeptID      *int
	CallerUID   int
}

// ArticleListResult 分页列表响应。
type ArticleListResult struct {
	List       []ArticleListItem      `json:"list"`
	Pagination pagination.NestPagination    `json:"pagination"`
}

// ArticleListItem 列表项（正文置空减少传输）。
type ArticleListItem struct {
	ID           int           `json:"id"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Cover        string        `json:"cover,omitempty"`
	Status       string        `json:"status"`
	Topping      int           `json:"topping"`
	Views        int           `json:"views"`
	Likes        int           `json:"likes"`
	CreateTime   time.Time     `json:"createTime"`
	UpdateTime   time.Time     `json:"updateTime"`
	UTime        string        `json:"uTime,omitempty"`
	Category     *CategoryItem `json:"category,omitempty"`
	Tags         []TagItem     `json:"tags,omitempty"`
	UserInfo     *UserInfoItem `json:"userInfo,omitempty"`
	AuthorName   string        `json:"authorName,omitempty"`
	DeptName     string        `json:"deptName,omitempty"`
	CommentCount int           `json:"commentCount"`
	Content      string        `json:"content"`
	ContentHTML  string        `json:"contentHtml"`
}

// ArticleDetailResult 详情响应（含 prev/next）。
type ArticleDetailResult struct {
	Info ArticleDetailItem `json:"info"`
	Prev *NavItem          `json:"prev"`
	Next *NavItem          `json:"next"`
}

// ArticleDetailItem 详情 info 字段。
type ArticleDetailItem struct {
	ID                 int           `json:"id"`
	Title              string        `json:"title"`
	Description        string        `json:"description"`
	Cover              string        `json:"cover,omitempty"`
	Status             string        `json:"status"`
	Topping            int           `json:"topping"`
	Views              int           `json:"views"`
	Likes              int           `json:"likes"`
	CreateTime         time.Time     `json:"createTime"`
	UpdateTime         time.Time     `json:"updateTime"`
	UTime              string        `json:"uTime,omitempty"`
	Content            string        `json:"content"`
	ContentHTML        string        `json:"contentHtml"`
	Category           *CategoryItem `json:"category,omitempty"`
	Tags               []TagItem     `json:"tags,omitempty"`
	UserInfo           *UserInfoItem `json:"userInfo,omitempty"`
	ScheduledPublishAt *time.Time    `json:"scheduledPublishAt,omitempty"`
	ArticleExp         int           `json:"articleExp,omitempty"`
	ArticleLevel       int           `json:"articleLevel,omitempty"`
	ReputationGained   int           `json:"reputationGained,omitempty"`
	IsMasterpiece      int           `json:"isMasterpiece,omitempty"`
	TipTotal           int           `json:"tipTotal,omitempty"`
}

// NavItem 同作者相邻文章导航项。
type NavItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// CreateArticleInput 创建文章参数。
type CreateArticleInput struct {
	Title              string
	Description        string
	Content            string
	ContentHTML        string
	Cover              string
	Status             string
	ScheduledPublishAt *time.Time
	CategoryID         string
	TagIDs             []string
}

// EditArticleInput 编辑文章参数。
type EditArticleInput struct {
	ID                 int
	Title              *string
	Description        *string
	Content            *string
	ContentHTML        *string
	Cover              *string
	Status             *string
	IsDelete           *bool
	ScheduledPublishAt *time.Time
	CategoryID         *string
	TagIDs             []string
}
