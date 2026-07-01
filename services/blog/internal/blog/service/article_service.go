// Package service 文章业务逻辑；跨模块通过 UserService 获取作者，禁止跨表 JOIN user。
package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/pkg/pagination"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/domain"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/blog/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/services/blog/internal/userport"
)

// ArticleService 文章 CRUD 与查询。
type ArticleService struct {
	articles   *blogrepo.ArticleRepo
	categories *CategoryService
	tags       *TagService
	comments   *blogrepo.CommentRepo
	users      usersvc.UserService
	userPort   userport.ArticleUserPort
	admin      userport.ArticleAdminPort
}

// NewArticleService 构造 ArticleService。
func NewArticleService(
	articles *blogrepo.ArticleRepo,
	categories *CategoryService,
	tags *TagService,
	comments *blogrepo.CommentRepo,
	users usersvc.UserService,
	userPort userport.ArticleUserPort,
	adminSvc userport.ArticleAdminPort,
) *ArticleService {
	return &ArticleService{
		articles:   articles,
		categories: categories,
		tags:       tags,
		comments:   comments,
		users:      users,
		userPort:   userPort,
		admin:      adminSvc,
	}
}

// List 分页列表。
func (s *ArticleService) List(ctx context.Context, q domain.ArticleListQuery) (*domain.ArticleListResult, error) {
	filter := blogrepo.ListFilter{
		Page:           q.Page,
		PageSize:       q.PageSize,
		CategoryID:     q.Category,
		TagIDs:         q.Tags,
		Title:          q.Title,
		Description:    q.Description,
		Content:        q.Content,
		SortAsc:        strings.ToUpper(q.Sort) == "ASC",
		Client:         q.Client,
		OnlyNotDeleted: q.Client,
	}
	if q.Client {
		// C 端列表：仅展示未锁定作者的公开文章（user gRPC 拉 active UID，禁止跨库 JOIN user）。
		activeUIDs, err := s.userPort.ListActiveUserIDs(ctx)
		if err != nil {
			return nil, err
		}
		filter.AuthorUIDs = activeUIDs
	} else if q.CallerUID > 0 {
		// 管理端：按 RBAC 数据权限过滤可访问部门下的作者。
		deptIDs, err := s.admin.ResolveArticleAccessibleDeptIDs(ctx, q.CallerUID)
		if err != nil {
			return nil, err
		}
		if deptIDs != nil {
			if len(deptIDs) == 0 {
				return emptyList(q.Page, q.PageSize), nil
			}
			if q.DeptID != nil {
				if !containsInt(deptIDs, *q.DeptID) {
					return emptyList(q.Page, q.PageSize), nil
				}
				filter.DeptID = q.DeptID
			} else {
				filter.DeptIDs = deptIDs
			}
		} else if q.DeptID != nil {
			filter.DeptID = q.DeptID
		}
	} else if q.DeptID != nil {
		filter.DeptID = q.DeptID
	}

	rows, total, err := s.articles.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	articleIDs := make([]int, len(rows))
	uids := make([]uint64, 0, len(rows))
	for i, a := range rows {
		articleIDs[i] = a.ID
		uids = append(uids, uint64(a.UID))
	}
	tagMap, _ := s.articles.ListTagIDsByArticles(ctx, articleIDs)
	userMap := s.batchUsers(ctx, uids)
	deptNames := s.batchDeptNames(ctx, rows)
	commentCounts, _ := s.comments.CountByArticleIDs(ctx, articleIDs)

	list := make([]domain.ArticleListItem, 0, len(rows))
	for i, a := range rows {
		item := s.toListItem(ctx, a, tagMap[a.ID], userMap[uint64(a.UID)], deptNames[i], commentCounts[a.ID])
		list = append(list, item)
	}
	return &domain.ArticleListResult{
		List:       list,
		Pagination: pagination.CalcNestPagination(total, q.PageSize, q.Page),
	}, nil
}

// Info 文章详情。
func (s *ArticleService) Info(ctx context.Context, key string) (*domain.ArticleDetailResult, error) {
	row, err := s.articles.FindByIDOrTitle(ctx, key)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "找不到文章")
		}
		return nil, err
	}
	if row.IsDelete {
		return nil, errcode.WithMessage(errcode.NotFound, "找不到文章")
	}
	author, err := s.users.GetUser(ctx, uint64(row.UID))
	if err != nil || author == nil || strings.EqualFold(author.Status, "locked") {
		return nil, errcode.WithMessage(errcode.NotFound, "找不到文章")
	}
	tagIDs, _ := s.articles.ListTagIDsByArticle(ctx, row.ID)
	tags, _ := s.tags.FindByIDs(ctx, tagIDs)
	var cat *ent.Category
	if row.Articles != nil && *row.Articles != "" {
		cat, _ = s.categories.FindByID(ctx, *row.Articles)
	}
	info := s.toDetailItem(row, cat, tags, author)
	prev, next := s.adjacentArticles(ctx, row)
	return &domain.ArticleDetailResult{Info: info, Prev: prev, Next: next}, nil
}

// Create 创建文章。
func (s *ArticleService) Create(ctx context.Context, uid int, in domain.CreateArticleInput) (interface{}, error) {
	if in.Title == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "请输入文章标题")
	}
	if in.ContentHTML == "" && in.Content != "" {
		html, err := RenderMarkdown(in.Content)
		if err != nil {
			return nil, err
		}
		in.ContentHTML = html
	}
	exists, err := s.articles.ExistsByTitle(ctx, in.Title, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.WithMessage(errcode.InternalError, "文章标题已存在")
	}
	u, err := s.userPort.FindUserForArticle(ctx, uid)
	if err != nil {
		return nil, err
	}
	if u.DeptID == nil {
		return nil, errcode.WithMessage(errcode.InvalidParam, "用户未关联机构，无法创建文章")
	}
	cat, err := s.categories.FindByID(ctx, in.CategoryID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "分类不存在")
		}
		return nil, err
	}
	tagRows, err := s.tags.FindByIDs(ctx, in.TagIDs)
	if err != nil {
		return nil, err
	}
	status := in.Status
	if status == "" {
		status = "publish"
	}
	now := time.Now()
	row := &ent.Article{
		UID:         uid,
		DeptId:      u.DeptID,
		Title:       in.Title,
		Description: in.Description,
		Content:     in.Content,
		ContentHtml: in.ContentHTML,
		Cover:       in.Cover,
		Status:      status,
		UTime:       now.Format(time.RFC3339),
		Articles:    &cat.ID,
	}
	if in.ScheduledPublishAt != nil {
		row.ScheduledPublishAt = in.ScheduledPublishAt
	}
	created, err := s.articles.Create(ctx, row)
	if err != nil {
		return nil, err
	}
	tagIDs := make([]string, len(tagRows))
	for i, t := range tagRows {
		tagIDs[i] = t.ID
	}
	_ = s.articles.ReplaceTags(ctx, created.ID, tagIDs)
	return created, nil
}

// Edit 编辑文章。
func (s *ArticleService) Edit(ctx context.Context, callerUID int, in domain.EditArticleInput) (interface{}, error) {
	row, err := s.articles.GetByID(ctx, in.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "文章不存在")
		}
		return nil, err
	}
	if callerUID > 0 {
		if err := s.admin.AssertArticleDeptAccess(ctx, callerUID, row.DeptId); err != nil {
			return nil, err
		}
	}
	fields := map[string]interface{}{"uTime": time.Now().Format(time.RFC3339)}
	if in.Title != nil {
		exists, err := s.articles.ExistsByTitle(ctx, *in.Title, in.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errcode.WithMessage(errcode.InternalError, "文章标题已存在")
		}
		fields["title"] = *in.Title
	}
	if in.Description != nil {
		fields["description"] = *in.Description
	}
	if in.Content != nil {
		fields["content"] = *in.Content
	}
	if in.ContentHTML != nil {
		fields["contentHtml"] = *in.ContentHTML
	} else if in.Content != nil {
		html, err := RenderMarkdown(*in.Content)
		if err != nil {
			return nil, err
		}
		fields["contentHtml"] = html
	}
	if in.Cover != nil {
		fields["cover"] = *in.Cover
	}
	if in.IsDelete != nil {
		fields["isDelete"] = *in.IsDelete
	}
	if in.Status != nil {
		fields["status"] = *in.Status
		if *in.Status != "scheduled" {
			fields["scheduledPublishAt"] = nil
		}
	}
	if in.ScheduledPublishAt != nil {
		fields["scheduledPublishAt"] = *in.ScheduledPublishAt
	}
	if in.CategoryID != nil {
		cat, err := s.categories.FindByID(ctx, *in.CategoryID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, errcode.WithMessage(errcode.NotFound, "分类不存在")
			}
			return nil, err
		}
		fields["articles"] = cat.ID
	}
	updated, err := s.articles.UpdateFields(ctx, in.ID, fields)
	if err != nil {
		return nil, err
	}
	if len(in.TagIDs) > 0 {
		tagRows, err := s.tags.FindByIDs(ctx, in.TagIDs)
		if err != nil {
			return nil, err
		}
		tagIDs := make([]string, len(tagRows))
		for i, t := range tagRows {
			tagIDs[i] = t.ID
		}
		_ = s.articles.ReplaceTags(ctx, in.ID, tagIDs)
	}
	return map[string]interface{}{"info": updated}, nil
}

// Delete 物理删除。
func (s *ArticleService) Delete(ctx context.Context, callerUID, id int) (interface{}, error) {
	row, err := s.articles.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "文章不存在")
		}
		return nil, err
	}
	if callerUID > 0 {
		if err := s.admin.AssertArticleDeptAccess(ctx, callerUID, row.DeptId); err != nil {
			return nil, err
		}
	}
	if err := s.articles.HardDelete(ctx, id); err != nil {
		return nil, err
	}
	return map[string]interface{}{"info": map[string]string{"message": "删除成功"}}, nil
}

// UpdateViews 阅读量 +1。
func (s *ArticleService) UpdateViews(ctx context.Context, id int) (bool, error) {
	return s.articles.IncrementViews(ctx, id)
}

// UpdateField 禁用/置顶等字段更新。
func (s *ArticleService) UpdateField(ctx context.Context, callerUID int, id int, fields map[string]interface{}) (interface{}, error) {
	row, err := s.articles.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "文章不存在")
		}
		return nil, err
	}
	if callerUID > 0 && row.UID != callerUID {
		if err := s.admin.AssertArticleDeptAccess(ctx, callerUID, row.DeptId); err != nil {
			return nil, err
		}
	}
	fields["uTime"] = time.Now().Format(time.RFC3339)
	updated, err := s.articles.UpdateFields(ctx, id, fields)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// MyList 当前用户文章列表。
func (s *ArticleService) MyList(ctx context.Context, uid, page, pageSize int) (interface{}, error) {
	filter := blogrepo.ListFilter{
		Page:           page,
		PageSize:       pageSize,
		UID:            &uid,
		OnlyNotDeleted: true,
		SortAsc:        false,
	}
	rows, total, err := s.articles.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, a := range rows {
		item := map[string]interface{}{
			"id": a.ID, "title": a.Title, "description": a.Description,
			"status": a.Status, "isDelete": a.IsDelete, "topping": a.Topping,
			"views": a.Views, "likes": a.Likes,
			"createTime": a.CreateTime, "updateTime": a.UpdateTime,
		}
		if a.Articles != nil {
			if cat, err := s.categories.FindByID(ctx, *a.Articles); err == nil {
				item["category"] = map[string]string{"id": cat.ID, "label": cat.Label}
			}
		}
		list = append(list, item)
	}
	return map[string]interface{}{
		"list":       list,
		"pagination": pagination.CalcNestPagination(total, pageSize, page),
	}, nil
}

// Archives 文章归档。
func (s *ArticleService) Archives(ctx context.Context) (interface{}, error) {
	rows, err := s.articles.ListPublishedAll(ctx)
	if err != nil {
		return nil, err
	}
	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	ret := map[int]map[string][]map[string]interface{}{}
	for _, a := range rows {
		year := a.CreateTime.Year()
		month := months[a.CreateTime.Month()-1]
		if ret[year] == nil {
			ret[year] = map[string][]map[string]interface{}{}
		}
		ret[year][month] = append(ret[year][month], map[string]interface{}{
			"title": a.Title, "id": a.ID, "createTime": a.CreateTime, "uTime": a.UTime,
		})
	}
	type yearBlock struct {
		Year string                 `json:"year"`
		Data map[string]interface{} `json:"data"`
	}
	out := make([]yearBlock, 0)
	years := sortedYears(ret)
	for _, y := range years {
		out = append(out, yearBlock{Year: itoa(y), Data: toInterfaceMap(ret[y])})
	}
	return out, nil
}

// Related 相关文章推荐。
func (s *ArticleService) Related(ctx context.Context, id string, limit int) (interface{}, error) {
	articleID, err := parseArticleID(id)
	if err != nil {
		return map[string]interface{}{"list": []interface{}{}}, nil
	}
	current, err := s.articles.GetByID(ctx, articleID)
	if err != nil || current.IsDelete || current.Status != "publish" {
		return map[string]interface{}{"list": []interface{}{}}, nil
	}
	author, _ := s.users.GetUser(ctx, uint64(current.UID))
	if author == nil || strings.EqualFold(author.Status, "locked") {
		return map[string]interface{}{"list": []interface{}{}}, nil
	}
	related, _ := s.articles.ListRelated(ctx, current.UID, articleID, limit)
	if len(related) < limit {
		fill, _ := s.articles.ListLatestPublished(ctx, articleID, limit-len(related))
		related = append(related, fill...)
	}
	list := make([]map[string]interface{}, 0, len(related))
	for _, a := range related {
		list = append(list, map[string]interface{}{
			"id": a.ID, "title": a.Title, "description": a.Description,
			"cover": a.Cover, "views": a.Views, "createTime": a.CreateTime,
		})
	}
	return map[string]interface{}{"list": list}, nil
}

// AuthorStats 作者统计。
func (s *ArticleService) AuthorStats(ctx context.Context, uid int) (interface{}, error) {
	total, _ := s.articles.CountByAuthor(ctx, uid, "", false)
	published, _ := s.articles.CountByAuthor(ctx, uid, "publish", false)
	draft, _ := s.articles.CountByAuthor(ctx, uid, "draft", false)
	scheduled, _ := s.articles.CountByAuthor(ctx, uid, "scheduled", false)
	totalViews, _ := s.articles.SumViewsByAuthor(ctx, uid)
	top, _ := s.articles.TopByAuthor(ctx, uid, 5)
	topList := make([]map[string]interface{}, 0, len(top))
	var totalLikes int
	for _, a := range top {
		totalLikes += a.Likes
		topList = append(topList, map[string]interface{}{
			"id": a.ID, "title": a.Title, "views": a.Views, "likes": a.Likes, "createTime": a.CreateTime,
		})
	}
	allLikes := totalLikes
	if len(top) < int(published) {
		allRows, _, _ := s.articles.List(ctx, blogrepo.ListFilter{UID: &uid, OnlyNotDeleted: true, Page: 1, PageSize: 1000})
		allLikes = 0
		for _, a := range allRows {
			allLikes += a.Likes
		}
	}
	return map[string]interface{}{
		"total": total, "published": published, "draft": draft, "scheduled": scheduled,
		"totalViews": totalViews, "totalLikes": allLikes, "topArticles": topList,
	}, nil
}

// Statistics 数据大屏统计（简化版，评论相关留 Plan 06）。
func (s *ArticleService) Statistics(ctx context.Context) (interface{}, error) {
	rows, err := s.articles.ListPublishedAll(ctx)
	if err != nil {
		return nil, err
	}
	var totalViews, totalLikes int
	top := make([]*ent.Article, 0, 6)
	for i, a := range rows {
		totalViews += a.Views
		totalLikes += a.Likes
		if i < 6 {
			top = append(top, a)
		}
	}
	topArticles := make([]map[string]interface{}, 0, len(top))
	for _, a := range top {
		topArticles = append(topArticles, map[string]interface{}{
			"id": a.ID, "title": a.Title, "views": a.Views, "likes": a.Likes,
		})
	}
	catMap := map[string]map[string]interface{}{}
	tagMap := map[string]map[string]interface{}{}
	for _, a := range rows {
		if a.Articles != nil {
			cid := *a.Articles
			if catMap[cid] == nil {
				if cat, err := s.categories.FindByID(ctx, cid); err == nil {
					catMap[cid] = map[string]interface{}{"id": cid, "label": cat.Label, "count": 0}
				}
			}
			if catMap[cid] != nil {
				catMap[cid]["count"] = catMap[cid]["count"].(int) + 1
			}
		}
		tagIDs, _ := s.articles.ListTagIDsByArticle(ctx, a.ID)
		for _, tid := range tagIDs {
			if tagMap[tid] == nil {
				if t, err := s.tags.FindByIDs(ctx, []string{tid}); err == nil && len(t) > 0 {
					tagMap[tid] = map[string]interface{}{"id": tid, "label": t[0].Label, "count": 0}
				}
			}
			if tagMap[tid] != nil {
				tagMap[tid]["count"] = tagMap[tid]["count"].(int) + 1
			}
		}
	}
	archives, _ := s.Archives(ctx)
	return map[string]interface{}{
		"articles": topArticles, "total": len(rows), "totalViews": totalViews, "totalLikes": totalLikes,
		"totalComments": 0,
		"trends": map[string]int{"article": 0, "views": 0, "likes": 0, "comments": 0},
		"categories": mapValues(catMap), "tags": mapValues(tagMap), "archives": archives,
	}, nil
}

// UpdateLikes 更新点赞计数（Nest 兼容，Plan 06 主路径为 /like）。
func (s *ArticleService) UpdateLikes(ctx context.Context, articleID int, status int) error {
	row, err := s.articles.GetByID(ctx, articleID)
	if err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "文章不存在")
		}
		return err
	}
	likes := row.Likes
	if status == 1 {
		likes++
	} else {
		likes--
		if likes < 0 {
			likes = 0
		}
	}
	_, err = s.articles.UpdateFields(ctx, articleID, map[string]interface{}{"likes": likes})
	return err
}

// --- helpers ---

func (s *ArticleService) toListItem(ctx context.Context, a *ent.Article, tagIDs []string, author *usersvc.UserDTO, deptName string, commentCount int) domain.ArticleListItem {
	var cat *domain.CategoryItem
	if a.Articles != nil {
		if c, err := s.categories.FindByID(ctx, *a.Articles); err == nil {
			cat = s.categories.ToItem(c)
		}
	}
	tagRows, _ := s.tags.FindByIDs(ctx, tagIDs)
	tags := make([]domain.TagItem, 0, len(tagRows))
	for _, t := range tagRows {
		tags = append(tags, domain.TagItem{ID: t.ID, Label: t.Label, Value: t.Value, Color: t.Color})
	}
	var userInfo *domain.UserInfoItem
	authorName := ""
	if author != nil {
		userInfo = &domain.UserInfoItem{ID: author.ID, Nickname: author.Nickname, Username: author.Username, Avatar: author.Avatar}
		authorName = author.Nickname
		if authorName == "" {
			authorName = author.Username
		}
	}
	return domain.ArticleListItem{
		ID: a.ID, Title: a.Title, Description: a.Description, Cover: a.Cover,
		Status: a.Status, Topping: a.Topping, Views: a.Views, Likes: a.Likes,
		CreateTime: a.CreateTime, UpdateTime: a.UpdateTime, UTime: a.UTime,
		Category: cat, Tags: tags, UserInfo: userInfo, AuthorName: authorName, DeptName: deptName,
		CommentCount: commentCount, Content: "", ContentHTML: "",
	}
}

func (s *ArticleService) toDetailItem(a *ent.Article, cat *ent.Category, tagRows []*ent.Tag, author *usersvc.UserDTO) domain.ArticleDetailItem {
	tags := make([]domain.TagItem, 0, len(tagRows))
	for _, t := range tagRows {
		tags = append(tags, domain.TagItem{ID: t.ID, Label: t.Label, Value: t.Value, Color: t.Color})
	}
	var catItem *domain.CategoryItem
	if cat != nil {
		catItem = s.categories.ToItem(cat)
	}
	contentHTML := a.ContentHtml
	if contentHTML == "" && a.Content != "" {
		contentHTML, _ = RenderMarkdown(a.Content)
	}
	var userInfo *domain.UserInfoItem
	if author != nil {
		userInfo = &domain.UserInfoItem{ID: author.ID, Nickname: author.Nickname, Username: author.Username, Avatar: author.Avatar}
	}
	return domain.ArticleDetailItem{
		ID: a.ID, Title: a.Title, Description: a.Description, Cover: a.Cover,
		Status: a.Status, Topping: a.Topping, Views: a.Views, Likes: a.Likes,
		CreateTime: a.CreateTime, UpdateTime: a.UpdateTime, UTime: a.UTime,
		Content: a.Content, ContentHTML: contentHTML,
		Category: catItem, Tags: tags, UserInfo: userInfo,
		ScheduledPublishAt: a.ScheduledPublishAt,
	}
}

func (s *ArticleService) adjacentArticles(ctx context.Context, row *ent.Article) (*domain.NavItem, *domain.NavItem) {
	if row.Status != "publish" || row.IsDelete {
		return nil, nil
	}
	articles, err := s.articles.ListPublishedByAuthor(ctx, row.UID)
	if err != nil {
		return nil, nil
	}
	idx := -1
	for i, a := range articles {
		if a.ID == row.ID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, nil
	}
	var prev, next *domain.NavItem
	if idx > 0 {
		prev = &domain.NavItem{ID: articles[idx-1].ID, Title: articles[idx-1].Title}
	}
	if idx+1 < len(articles) {
		next = &domain.NavItem{ID: articles[idx+1].ID, Title: articles[idx+1].Title}
	}
	return prev, next
}

func (s *ArticleService) batchUsers(ctx context.Context, ids []uint64) map[uint64]*usersvc.UserDTO {
	out := map[uint64]*usersvc.UserDTO{}
	if len(ids) == 0 {
		return out
	}
	rows, err := s.users.GetUserBatch(ctx, ids)
	if err != nil {
		return out
	}
	for _, u := range rows {
		if u != nil {
			out[u.ID] = u
		}
	}
	return out
}

func (s *ArticleService) batchDeptNames(ctx context.Context, rows []*ent.Article) []string {
	names := make([]string, len(rows))
	for i, a := range rows {
		if a.DeptId != nil {
			if d, err := s.userPort.FindDeptByID(ctx, *a.DeptId); err == nil {
				names[i] = d.DeptName
			}
		}
	}
	return names
}

func emptyList(page, pageSize int) *domain.ArticleListResult {
	return &domain.ArticleListResult{
		List:       []domain.ArticleListItem{},
		Pagination: pagination.CalcNestPagination(0, pageSize, page),
	}
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func parseArticleID(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func sortedYears(m map[int]map[string][]map[string]interface{}) []int {
	years := make([]int, 0, len(m))
	for y := range m {
		years = append(years, y)
	}
	for i := 0; i < len(years); i++ {
		for j := i + 1; j < len(years); j++ {
			if years[j] > years[i] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}
	return years
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func toInterfaceMap(m map[string][]map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func mapValues(m map[string]map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(m))
	for _, v := range m {
		v["articleCount"] = v["count"]
		out = append(out, v)
	}
	return out
}
