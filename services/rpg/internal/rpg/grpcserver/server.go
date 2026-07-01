// Package grpcserver 实现 rpg.v1.RpgService gRPC 服务端。
package grpcserver

import (
	"context"
	"encoding/json"

	rpgv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/rpg/v1"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
	rpgprofile "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/profile"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server 实现 RpgService gRPC。
type Server struct {
	rpgv1.UnimplementedRpgServiceServer
	core    *rpgcore.RpgService
	profile *rpgprofile.Service
}

// New 构造 gRPC RpgService 实现。
func New(core *rpgcore.RpgService, profile *rpgprofile.Service) *Server {
	return &Server{core: core, profile: profile}
}

// GetProfile 返回用户 RPG 等级与经验摘要。
func (s *Server) GetProfile(ctx context.Context, req *rpgv1.GetProfileRequest) (*rpgv1.GetProfileResponse, error) {
	uid := int(req.GetUserId())
	if uid <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	level := int32(1)
	var exp int64
	if s.core != nil {
		row, err := s.core.FindByUid(ctx, uid)
		if err == nil && row != nil {
			level = int32(row.Level)
			exp = int64(row.Exp)
		}
	}
	return &rpgv1.GetProfileResponse{
		UserId: req.GetUserId(),
		Level:  level,
		Exp:    exp,
	}, nil
}

// GetPublicProfile 返回公开主页 JSON（与 HTTP /user/public/:uid 同构）。
func (s *Server) GetPublicProfile(ctx context.Context, req *rpgv1.GetPublicProfileRequest) (*rpgv1.GetPublicProfileResponse, error) {
	uid := int(req.GetUserId())
	if uid <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	if s.profile == nil {
		return nil, status.Error(codes.Unavailable, "profile service unavailable")
	}
	data, err := s.profile.GetPublicProfile(ctx, uid)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "public profile: %v", err)
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal profile: %v", err)
	}
	return &rpgv1.GetPublicProfileResponse{ProfileJson: raw}, nil
}
