package repo

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/reply"
)

// ReplyRepo 回复表读写。
type ReplyRepo struct {
	client *ent.Client
}

// NewReplyRepo 构造 ReplyRepo。
func NewReplyRepo(client *ent.Client) *ReplyRepo {
	return &ReplyRepo{client: client}
}

// ReplyFilter 回复列表筛选。
type ReplyFilter struct {
	ParentID string
	UID      *int
	Status   string
	Page     int
	PageSize int
	SortAsc  bool
}

// Create 写入回复。
func (r *ReplyRepo) Create(ctx context.Context, row *ent.Reply) (*ent.Reply, error) {
	return r.client.Reply.Create().
		SetID(row.ID).
		SetParentId(row.ParentId).
		SetReplyUid(row.ReplyUid).
		SetContent(row.Content).
		SetUID(row.UID).
		SetStatus(row.Status).
		Save(ctx)
}

// Delete 物理删除回复。
func (r *ReplyRepo) Delete(ctx context.Context, id string) error {
	_, err := r.client.Reply.Delete().Where(reply.IDEQ(id)).Exec(ctx)
	return err
}

// DeleteByParentID 删除评论下全部回复。
func (r *ReplyRepo) DeleteByParentID(ctx context.Context, parentID string) error {
	_, err := r.client.Reply.Delete().Where(reply.ParentIdEQ(parentID)).Exec(ctx)
	return err
}

// UpdateStatus 更新审核状态。
func (r *ReplyRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.client.Reply.UpdateOneID(id).SetStatus(status).Save(ctx)
	return err
}

// List 分页查询回复。
func (r *ReplyRepo) List(ctx context.Context, f ReplyFilter) ([]*ent.Reply, int, error) {
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	q := r.client.Reply.Query()
	if f.ParentID != "" {
		q = q.Where(reply.ParentIdEQ(f.ParentID))
	}
	if f.UID != nil {
		q = q.Where(reply.UIDEQ(*f.UID))
	}
	if f.Status != "" {
		q = q.Where(reply.StatusEQ(f.Status))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if f.SortAsc {
		q = q.Order(ent.Asc(reply.FieldCreateTime))
	} else {
		q = q.Order(ent.Desc(reply.FieldCreateTime))
	}
	rows, err := q.Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	return rows, total, err
}
