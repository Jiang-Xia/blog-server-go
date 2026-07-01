// category_repo 分类表 Ent 读写。
package repo

import (
	"context"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/category"
)

// CategoryRepo 分类表读写。
type CategoryRepo struct {
	client *ent.Client
}

// NewCategoryRepo 构造 CategoryRepo。
func NewCategoryRepo(client *ent.Client) *CategoryRepo {
	return &CategoryRepo{client: client}
}

// List 查询全部分类。
func (r *CategoryRepo) List(ctx context.Context, title string) ([]*ent.Category, error) {
	q := r.client.Category.Query().Order(ent.Asc(category.FieldCreateAt))
	if title != "" {
		q = q.Where(category.LabelContains(title))
	}
	return q.All(ctx)
}

// FindByID 按 id/label/value 查询。
func (r *CategoryRepo) FindByID(ctx context.Context, key string) (*ent.Category, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, &ent.NotFoundError{}
	}
	return r.client.Category.Query().
		Where(category.Or(category.IDEQ(key), category.LabelEQ(key), category.ValueEQ(key))).
		Only(ctx)
}

// FindByIDs 批量按 ID 查询。
func (r *CategoryRepo) FindByIDs(ctx context.Context, ids []string) ([]*ent.Category, error) {
	if len(ids) == 0 {
		return []*ent.Category{}, nil
	}
	return r.client.Category.Query().Where(category.IDIn(ids...)).All(ctx)
}

// ExistsByLabel 检查 label 是否已存在。
func (r *CategoryRepo) ExistsByLabel(ctx context.Context, label string) (bool, error) {
	return r.client.Category.Query().Where(category.LabelEQ(label)).Exist(ctx)
}

// Create 创建分类。
func (r *CategoryRepo) Create(ctx context.Context, row *ent.Category) (*ent.Category, error) {
	now := time.Now()
	b := r.client.Category.Create().
		SetID(row.ID).
		SetUID(row.UID).
		SetLabel(row.Label).
		SetValue(row.Value).
		SetColor(row.Color).
		SetCreateAt(now).
		SetUpdateAt(now)
	return b.Save(ctx)
}

// Update 更新分类。
func (r *CategoryRepo) Update(ctx context.Context, id string, label, value, color *string) (*ent.Category, error) {
	up := r.client.Category.UpdateOneID(id).SetUpdateAt(time.Now())
	if label != nil {
		up = up.SetLabel(*label)
	}
	if value != nil {
		up = up.SetValue(*value)
	}
	if color != nil {
		up = up.SetColor(*color)
	}
	return up.Save(ctx)
}

// Delete 删除分类。
func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	return r.client.Category.DeleteOneID(id).Exec(ctx)
}
