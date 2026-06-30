// Package user 提供 UserService 端口的本地实现，委托 profile 模块。
package user

import (
	"context"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/profile"
)

// localUserService UserService 的 monolith 本地实现。
type localUserService struct {
	profile *profile.Service
}

// NewUserService 构造可被 wire 注入的 UserService。
func NewUserService(profileSvc *profile.Service) UserService {
	return &localUserService{profile: profileSvc}
}

// GetUser 按 ID 查询用户摘要。
func (s *localUserService) GetUser(ctx context.Context, id uint64) (*UserDTO, error) {
	dto, err := s.profile.GetUserDTO(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPortDTO(dto), nil
}

// GetUserBatch 批量查询用户摘要，顺序与 ids 一致，缺失 ID 跳过。
func (s *localUserService) GetUserBatch(ctx context.Context, ids []uint64) ([]*UserDTO, error) {
	rows, err := s.profile.GetUserDTOBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]*UserDTO, 0, len(rows))
	for _, dto := range rows {
		out = append(out, toPortDTO(dto))
	}
	return out, nil
}

func toPortDTO(d *profile.UserDTO) *UserDTO {
	if d == nil {
		return nil
	}
	return &UserDTO{
		ID:       d.ID,
		Nickname: d.Nickname,
		Username: d.Username,
		Avatar:   d.Avatar,
		Email:    d.Email,
		Status:   d.Status,
	}
}
