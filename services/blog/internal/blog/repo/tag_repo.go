package repo

import (
	"context"
	"strings"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/blog/ent"
	"github.com/Jiang-Xia/blog-server-go/services/blog/ent/tag"
)

// TagRepo 标签表读写。
type TagRepo struct {
	client *ent.Client
}

// NewTagRepo 构造 TagRepo。
func NewTagRepo(client *ent.Client) *TagRepo {
	return &TagRepo{client: client}
}

// List 查询全部标签。
func (r *TagRepo) List(ctx context.Context, title string) ([]*ent.Tag, error) {
	q := r.client.Tag.Query().Order(ent.Asc(tag.FieldCreateAt))
	if title != "" {
		q = q.Where(tag.LabelContains(title))
	}
	return q.All(ctx)
}

// FindByID 按 id/label/value 查询。
func (r *TagRepo) FindByID(ctx context.Context, key string) (*ent.Tag, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, &ent.NotFoundError{}
	}
	return r.client.Tag.Query().
		Where(tag.Or(tag.IDEQ(key), tag.LabelEQ(key), tag.ValueEQ(key))).
		Only(ctx)
}

// FindByIDs 批量按 ID 查询。
func (r *TagRepo) FindByIDs(ctx context.Context, ids []string) ([]*ent.Tag, error) {
	if len(ids) == 0 {
		return []*ent.Tag{}, nil
	}
	return r.client.Tag.Query().Where(tag.IDIn(ids...)).All(ctx)
}

// ExistsByLabel 检查 label 是否已存在。
func (r *TagRepo) ExistsByLabel(ctx context.Context, label string) (bool, error) {
	return r.client.Tag.Query().Where(tag.LabelEQ(label)).Exist(ctx)
}

// Create 创建标签。
func (r *TagRepo) Create(ctx context.Context, row *ent.Tag) (*ent.Tag, error) {
	now := time.Now()
	return r.client.Tag.Create().
		SetID(row.ID).
		SetUID(row.UID).
		SetLabel(row.Label).
		SetValue(row.Value).
		SetColor(row.Color).
		SetCreateAt(now).
		SetUpdateAt(now).
		Save(ctx)
}

// Update 更新标签。
func (r *TagRepo) Update(ctx context.Context, id string, label, value, color *string) (*ent.Tag, error) {
	up := r.client.Tag.UpdateOneID(id).SetUpdateAt(time.Now())
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

// Delete 删除标签。
func (r *TagRepo) Delete(ctx context.Context, id string) error {
	return r.client.Tag.DeleteOneID(id).Exec(ctx)
}
