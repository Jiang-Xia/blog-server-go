// Package user 提供 UserService 端口的本地实现，委托 profile 模块。
package user

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/profile"
)

type localUserService struct {
	profile *profile.Service
}

// NewUserService 构造可被 wire 注入的 UserService。
func NewUserService(profileSvc *profile.Service) usersvc.UserService {
	return &localUserService{profile: profileSvc}
}

func (s *localUserService) GetUser(ctx context.Context, id uint64) (*usersvc.UserDTO, error) {
	dto, err := s.profile.GetUserDTO(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPortDTO(dto), nil
}

func (s *localUserService) GetUserBatch(ctx context.Context, ids []uint64) ([]*usersvc.UserDTO, error) {
	rows, err := s.profile.GetUserDTOBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]*usersvc.UserDTO, 0, len(rows))
	for _, dto := range rows {
		out = append(out, toPortDTO(dto))
	}
	return out, nil
}

func toPortDTO(d *profile.UserDTO) *usersvc.UserDTO {
	if d == nil {
		return nil
	}
	return &usersvc.UserDTO{
		ID:       d.ID,
		Nickname: d.Nickname,
		Username: d.Username,
		Avatar:   d.Avatar,
		Email:    d.Email,
		Status:   d.Status,
	}
}
