// Package grpcserver 实现 rpg.v1.RpgService gRPC 服务端。
package grpcserver

import (
	"context"
	"encoding/json"

	rpgv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/rpg/v1"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
	rpgprofile "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/profile"
	rpgpunish "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/punishment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server 实现 RpgService gRPC。
type Server struct {
	rpgv1.UnimplementedRpgServiceServer
	core       *rpgcore.RpgService
	profile    *rpgprofile.Service
	punishment *rpgpunish.PunishmentService
}

// New 构造 gRPC RpgService 实现。
func New(core *rpgcore.RpgService, profile *rpgprofile.Service, punishment *rpgpunish.PunishmentService) *Server {
	return &Server{core: core, profile: profile, punishment: punishment}
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

// AssertNotBanned 检查用户禁言状态，供 blog BanGuard 调用。
func (s *Server) AssertNotBanned(ctx context.Context, req *rpgv1.AssertNotBannedRequest) (*rpgv1.AssertNotBannedResponse, error) {
	uid := int(req.GetUserId())
	if uid <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	if s.punishment == nil {
		return &rpgv1.AssertNotBannedResponse{}, nil
	}
	if err := s.punishment.AssertNotBanned(ctx, uid); err != nil {
		if ec, ok := err.(errcode.ErrCode); ok {
			return &rpgv1.AssertNotBannedResponse{Banned: true, Message: ec.Message()}, nil
		}
		return nil, status.Errorf(codes.Internal, "ban check: %v", err)
	}
	return &rpgv1.AssertNotBannedResponse{}, nil
}
