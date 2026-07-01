// link_repo 友链表 Ent 读写。
package repo

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/link"
)

// LinkRepo 友链表读写。
type LinkRepo struct {
	client *ent.Client
}

// NewLinkRepo 构造 LinkRepo。
func NewLinkRepo(client *ent.Client) *LinkRepo {
	return &LinkRepo{client: client}
}

// Create 创建友链。
func (r *LinkRepo) Create(ctx context.Context, row *ent.Link) (*ent.Link, error) {
	return r.client.Link.Create().
		SetIcon(row.Icon).
		SetURL(row.URL).
		SetTitle(row.Title).
		SetDesp(row.Desp).
		SetAgreed(row.Agreed).
		SetLastCheckStatus(row.LastCheckStatus).
		Save(ctx)
}

// FindByURL 按 URL 查重。
func (r *LinkRepo) FindByURL(ctx context.Context, url string) (*ent.Link, error) {
	return r.client.Link.Query().Where(link.URLEQ(url)).First(ctx)
}

// GetByID 按 id 查询。
func (r *LinkRepo) GetByID(ctx context.Context, id int) (*ent.Link, error) {
	return r.client.Link.Query().Where(link.IDEQ(id)).Only(ctx)
}

// List 友链列表。
func (r *LinkRepo) List(ctx context.Context, client bool, title, url string, agreed *bool) ([]*ent.Link, error) {
	q := r.client.Link.Query()
	if client {
		q = q.Where(link.AgreedEQ(1))
	}
	if title != "" {
		q = q.Where(link.TitleContains(title))
	}
	if url != "" {
		q = q.Where(link.URLContains(url))
	}
	if agreed != nil {
		v := 0
		if *agreed {
			v = 1
		}
		q = q.Where(link.AgreedEQ(v))
	}
	return q.Order(ent.Asc(link.FieldCreateTime)).All(ctx)
}

// Update 部分更新友链。
func (r *LinkRepo) Update(ctx context.Context, id int, fields map[string]interface{}) (*ent.Link, error) {
	up := r.client.Link.UpdateOneID(id)
	if v, ok := fields["icon"].(string); ok {
		up.SetIcon(v)
	}
	if v, ok := fields["url"].(string); ok {
		up.SetURL(v)
	}
	if v, ok := fields["title"].(string); ok {
		up.SetTitle(v)
	}
	if v, ok := fields["desp"].(string); ok {
		up.SetDesp(v)
	}
	if v, ok := fields["agreed"]; ok {
		switch a := v.(type) {
		case bool:
			if a {
				up.SetAgreed(1)
			} else {
				up.SetAgreed(0)
			}
		case float64:
			up.SetAgreed(int(a))
		}
	}
	return up.Save(ctx)
}

// Delete 删除友链。
func (r *LinkRepo) Delete(ctx context.Context, id int) error {
	return r.client.Link.DeleteOneID(id).Exec(ctx)
}
