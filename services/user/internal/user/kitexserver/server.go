// Package kitexserver 实现 user.v1.UserService Kitex 服务端。
// 数据来源：profile / auth / email / sensitive / admin / user repo。
package kitexserver

import (
	"context"
	"fmt"
	"log"
	"os"

	userv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/user/v1"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/admin"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/email"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/profile"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/sensitive"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server 实现 UserService Kitex 接口。
type Server struct {
	profile   *profile.Service
	jwt       *auth.JWTService
	email     *email.Service
	sensitive *sensitive.Service
	admin     *admin.Service
	users     *repo.UserRepo
}

// New 构造 Kitex UserService 实现。
func New(
	profileSvc *profile.Service,
	jwtSvc *auth.JWTService,
	emailSvc *email.Service,
	sensitiveSvc *sensitive.Service,
	adminSvc *admin.Service,
	userRepo *repo.UserRepo,
) *Server {
	return &Server{
		profile:   profileSvc,
		jwt:       jwtSvc,
		email:     emailSvc,
		sensitive: sensitiveSvc,
		admin:     adminSvc,
		users:     userRepo,
	}
}

// GetUser 按 ID 返回用户摘要。
func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	dto, err := s.profile.GetUserDTO(ctx, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("not found: user not found: %w", err)
	}
	return toProto(dto), nil
}

// GetUserBatch 批量返回用户摘要。
func (s *Server) GetUserBatch(ctx context.Context, req *userv1.GetUserBatchRequest) (*userv1.GetUserBatchResponse, error) {
	rows, err := s.profile.GetUserDTOBatch(ctx, req.GetIds())
	if err != nil {
		return nil, fmt.Errorf("batch get users: %w", err)
	}
	out := make([]*userv1.GetUserResponse, 0, len(rows))
	for _, dto := range rows {
		out = append(out, toProto(dto))
	}
	return &userv1.GetUserBatchResponse{Users: out}, nil
}

// VerifyToken 校验 JWT 并返回 userID 与角色名列表。
func (s *Server) VerifyToken(ctx context.Context, req *userv1.VerifyTokenRequest) (*userv1.VerifyTokenResponse, error) {
	claims, err := s.jwt.Verify(req.GetToken())
	if err != nil {
		return nil, fmt.Errorf("unauthenticated: invalid token")
	}
	roles := make([]string, 0, len(claims.Role))
	for _, r := range claims.Role {
		roles = append(roles, r.RoleName)
	}
	return &userv1.VerifyTokenResponse{
		UserId: uint64(claims.ID),
		Roles:  roles,
	}, nil
}

// CountUsers 返回未删除用户总数（gateway pub/stats BFF）。
func (s *Server) CountUsers(ctx context.Context, _ *emptypb.Empty) (*userv1.CountUsersResponse, error) {
	total, err := s.profile.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	// 多实例学习：打出容器 hostname，便于对照 docker logs 观察 Kitex 负载均衡。
	host, _ := os.Hostname()
	log.Printf("[kitex] CountUsers instance=%s total=%d", host, total)
	return &userv1.CountUsersResponse{Total: int32(total)}, nil
}

// SendSystemEmail 发送系统 HTML 邮件（定时任务告警等）。
func (s *Server) SendSystemEmail(ctx context.Context, req *userv1.SendSystemEmailRequest) (*userv1.SendSystemEmailResponse, error) {
	if s.email == nil {
		return nil, fmt.Errorf("unavailable: email service not configured")
	}
	sent, err := s.email.SendSystemEmail(ctx, req.GetTo(), req.GetSubject(), req.GetHtmlBody())
	if err != nil {
		return nil, fmt.Errorf("send email: %w", err)
	}
	return &userv1.SendSystemEmailResponse{Sent: sent}, nil
}

func toProto(d *profile.UserDTO) *userv1.GetUserResponse {
	if d == nil {
		return nil
	}
	resp := &userv1.GetUserResponse{
		Id:       d.ID,
		Nickname: d.Nickname,
		Username: d.Username,
		Avatar:   d.Avatar,
		Email:    d.Email,
		Status:   d.Status,
	}
	if d.DeptID != nil {
		deptID := int32(*d.DeptID)
		resp.DeptId = &deptID
	}
	return resp
}
