package repo

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent/msgboard"
)

// MsgboardRepo 留言板表读写。
type MsgboardRepo struct {
	client *ent.Client
}

// NewMsgboardRepo 构造 MsgboardRepo。
func NewMsgboardRepo(client *ent.Client) *MsgboardRepo {
	return &MsgboardRepo{client: client}
}

// MsgboardFilter 留言列表筛选。
type MsgboardFilter struct {
	Status   string
	Name     string
	Comment  string
	Page     int
	PageSize int
}

// Create 写入留言。
func (r *MsgboardRepo) Create(ctx context.Context, row *ent.Msgboard) (*ent.Msgboard, error) {
	b := r.client.Msgboard.Create().
		SetName(row.Name).
		SetEamil(row.Eamil).
		SetAddress(row.Address).
		SetComment(row.Comment).
		SetAvatar(row.Avatar).
		SetLocation(row.Location).
		SetSystem(row.System).
		SetBrowser(row.Browser).
		SetPId(row.PId).
		SetStatus(row.Status)
	if row.Respondent != nil {
		b.SetRespondent(*row.Respondent)
	}
	if row.ImgUrl != nil {
		b.SetImgUrl(*row.ImgUrl)
	}
	if row.IP != nil {
		b.SetIP(*row.IP)
	}
	if row.ReplyId != nil {
		b.SetReplyId(*row.ReplyId)
	}
	return b.Save(ctx)
}

// DeleteByIDs 批量删除留言。
func (r *MsgboardRepo) DeleteByIDs(ctx context.Context, ids []int) (int, error) {
	return r.client.Msgboard.Delete().Where(msgboard.IDIn(ids...)).Exec(ctx)
}

// List 分页查询留言。
func (r *MsgboardRepo) List(ctx context.Context, f MsgboardFilter) ([]*ent.Msgboard, int, error) {
	page, pageSize := f.Page, f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := r.client.Msgboard.Query()
	if f.Status != "" && f.Status != "all" {
		q = q.Where(msgboard.StatusEQ(f.Status))
	} else if f.Status == "" {
		q = q.Where(msgboard.StatusEQ("approved"))
	}
	if f.Name != "" {
		q = q.Where(msgboard.NameContains(f.Name))
	}
	if f.Comment != "" {
		q = q.Where(msgboard.CommentContains(f.Comment))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := q.Order(ent.Desc(msgboard.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	return rows, total, err
}

// UpdateStatus 更新审核状态（敏感词审核同步）。
func (r *MsgboardRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	_, err := r.client.Msgboard.UpdateOneID(id).SetStatus(status).Save(ctx)
	return err
}
