package service

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
)

// LinkService 友链业务逻辑。
type LinkService struct {
	links *blogrepo.LinkRepo
}

// NewLinkService 构造 LinkService。
func NewLinkService(links *blogrepo.LinkRepo) *LinkService {
	return &LinkService{links: links}
}

// Create 创建友链。
func (s *LinkService) Create(ctx context.Context, icon, url, title, desp string) (*ent.Link, error) {
	if _, err := s.links.FindByURL(ctx, url); err == nil {
		return nil, errcode.WithMessage(errcode.InternalError, "外链已存在")
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	return s.links.Create(ctx, &ent.Link{
		Icon:            icon,
		URL:             url,
		Title:           title,
		Desp:            desp,
		Agreed:          0,
		LastCheckStatus: "unchecked",
	})
}

// Get 单条友链。
func (s *LinkService) Get(ctx context.Context, id int) (*ent.Link, error) {
	row, err := s.links.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "外链不存在")
		}
		return nil, err
	}
	return row, nil
}

// List 友链列表。
func (s *LinkService) List(ctx context.Context, client bool, title, url, agreedRaw string) ([]*ent.Link, error) {
	var agreed *bool
	if agreedRaw != "" {
		v := agreedRaw == "true" || agreedRaw == "1"
		agreed = &v
	}
	return s.links.List(ctx, client, title, url, agreed)
}

// Update 更新友链。
func (s *LinkService) Update(ctx context.Context, id int, fields map[string]interface{}) (*ent.Link, error) {
	if _, err := s.links.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return nil, errcode.WithMessage(errcode.NotFound, "外链不存在")
		}
		return nil, err
	}
	return s.links.Update(ctx, id, fields)
}

// Delete 删除友链。
func (s *LinkService) Delete(ctx context.Context, id int) error {
	if _, err := s.links.GetByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return errcode.WithMessage(errcode.NotFound, "外链不存在")
		}
		return err
	}
	if err := s.links.Delete(ctx, id); err != nil {
		return errcode.WithMessage(errcode.InternalError, "删除失败")
	}
	return nil
}
