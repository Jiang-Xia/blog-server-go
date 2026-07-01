// Package grpcserver 实现 user.v1.UserService gRPC 服务端。
package grpcserver

import (
	"context"

	userv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/user/v1"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/auth"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/profile"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server 实现 UserService gRPC。
type Server struct {
	userv1.UnimplementedUserServiceServer
	profile *profile.Service
	jwt     *auth.JWTService
}

// New 构造 gRPC UserService 实现。
func New(profileSvc *profile.Service, jwtSvc *auth.JWTService) *Server {
	return &Server{profile: profileSvc, jwt: jwtSvc}
}

// GetUser 按 ID 返回用户摘要。
func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	dto, err := s.profile.GetUserDTO(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	return toProto(dto), nil
}

// GetUserBatch 批量返回用户摘要。
func (s *Server) GetUserBatch(ctx context.Context, req *userv1.GetUserBatchRequest) (*userv1.GetUserBatchResponse, error) {
	rows, err := s.profile.GetUserDTOBatch(ctx, req.GetIds())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "batch get users: %v", err)
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
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
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

func toProto(d *profile.UserDTO) *userv1.GetUserResponse {
	if d == nil {
		return nil
	}
	return &userv1.GetUserResponse{
		Id:       d.ID,
		Nickname: d.Nickname,
		Username: d.Username,
		Avatar:   d.Avatar,
		Email:    d.Email,
		Status:   d.Status,
	}
}
