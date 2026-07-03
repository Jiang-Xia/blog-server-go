// cross.go Plan 17 跨服务协作 gRPC 方法（敏感词、数据权限、部门、命中记录）。
package grpcserver

import (
	"context"
	"errors"
	"time"

	userv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/user/v1"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/user/ent"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/repo"
	"github.com/Jiang-Xia/blog-server-go/services/user/internal/user/sensitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const defaultArticleResource = "article"

// EvaluateContent 敏感词分级检测。
func (s *Server) EvaluateContent(ctx context.Context, req *userv1.EvaluateContentRequest) (*userv1.EvaluateContentResponse, error) {
	if s.sensitive == nil {
		return nil, status.Errorf(codes.Unavailable, "sensitive service not configured")
	}
	result, err := s.sensitive.EvaluateContent(ctx, req.GetContent())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "evaluate content: %v", err)
	}
	hits := make([]*userv1.HitDetailMessage, 0, len(result.Hits))
	for _, h := range result.Hits {
		hits = append(hits, &userv1.HitDetailMessage{
			Word:       h.Word,
			Level:      int32(h.Level),
			HpPenalty:  int32(h.HpPenalty),
			NeedReview: int32(h.NeedReview),
			Action:     int32(h.Action),
		})
	}
	return &userv1.EvaluateContentResponse{
		Content:    result.Content,
		Hits:       hits,
		HitWords:   result.HitWords,
		HpPenalty:  int32(result.HpPenalty),
		NeedReview: result.NeedReview,
		Rejected:   result.Rejected,
	}, nil
}

// CreateHitRecord 写入敏感词命中记录。
func (s *Server) CreateHitRecord(ctx context.Context, req *userv1.CreateHitRecordRequest) (*userv1.CreateHitRecordResponse, error) {
	if s.sensitive == nil {
		return nil, status.Errorf(codes.Unavailable, "sensitive service not configured")
	}
	params := sensitive.CreateHitParams{
		SourceType: req.GetSourceType(),
		SourceID:   req.GetSourceId(),
		Content:    req.GetContent(),
		HitWords:   req.GetHitWords(),
	}
	if req.Uid != nil {
		uid := int(req.GetUid())
		params.UID = &uid
	}
	if req.Ip != nil {
		ip := req.GetIp()
		params.IP = &ip
	}
	if err := s.sensitive.CreateHitRecord(ctx, params); err != nil {
		return nil, status.Errorf(codes.Internal, "create hit record: %v", err)
	}
	return &userv1.CreateHitRecordResponse{Ok: true}, nil
}

// ListActiveUserIDs 返回 active 且未删除用户 ID，供 C 端文章列表过滤。
func (s *Server) ListActiveUserIDs(ctx context.Context, _ *emptypb.Empty) (*userv1.ListActiveUserIDsResponse, error) {
	if s.users == nil {
		return nil, status.Errorf(codes.Unavailable, "user repo not configured")
	}
	ids, err := s.users.ListActiveUserIDs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list active users: %v", err)
	}
	out := make([]int32, 0, len(ids))
	for _, id := range ids {
		out = append(out, int32(id))
	}
	return &userv1.ListActiveUserIDsResponse{UserIds: out}, nil
}

// GetDept 按 ID 返回部门名称。
func (s *Server) GetDept(ctx context.Context, req *userv1.GetDeptRequest) (*userv1.GetDeptResponse, error) {
	if s.users == nil {
		return nil, status.Errorf(codes.Unavailable, "user repo not configured")
	}
	d, err := s.users.FindDeptByID(ctx, int(req.GetId()))
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "dept not found")
		}
		return nil, status.Errorf(codes.Internal, "get dept: %v", err)
	}
	return &userv1.GetDeptResponse{Id: int32(d.ID), DeptName: d.DeptName}, nil
}

// ResolveAccessibleDeptIDs 解析用户数据权限可访问机构（当前仅 article 资源）。
func (s *Server) ResolveAccessibleDeptIDs(ctx context.Context, req *userv1.ResolveAccessibleDeptIDsRequest) (*userv1.ResolveAccessibleDeptIDsResponse, error) {
	if s.admin == nil {
		return nil, status.Errorf(codes.Unavailable, "admin service not configured")
	}
	resourceType := req.GetResourceType()
	if resourceType == "" {
		resourceType = defaultArticleResource
	}
	if resourceType != defaultArticleResource {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource_type: %s", resourceType)
	}
	deptIDs, err := s.admin.ResolveArticleAccessibleDeptIDs(ctx, int(req.GetUid()))
	if err != nil {
		return nil, grpcErrFromApp(err)
	}
	if deptIDs == nil {
		return &userv1.ResolveAccessibleDeptIDsResponse{Unrestricted: true}, nil
	}
	out := make([]int32, 0, len(deptIDs))
	for _, id := range deptIDs {
		out = append(out, int32(id))
	}
	return &userv1.ResolveAccessibleDeptIDsResponse{DeptIds: out}, nil
}

// AssertDeptAccess 校验用户是否有权访问指定机构。
func (s *Server) AssertDeptAccess(ctx context.Context, req *userv1.AssertDeptAccessRequest) (*userv1.AssertDeptAccessResponse, error) {
	if s.admin == nil {
		return nil, status.Errorf(codes.Unavailable, "admin service not configured")
	}
	resourceType := req.GetResourceType()
	if resourceType == "" {
		resourceType = defaultArticleResource
	}
	if resourceType != defaultArticleResource {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource_type: %s", resourceType)
	}
	var deptID *int
	if req.DeptId != nil {
		d := int(req.GetDeptId())
		deptID = &d
	}
	if err := s.admin.AssertArticleDeptAccess(ctx, int(req.GetUid()), deptID); err != nil {
		return nil, grpcErrFromApp(err)
	}
	return &userv1.AssertDeptAccessResponse{Allowed: true}, nil
}

// ListSensitiveWordHits 分页查询用户敏感词命中记录（RPG C 端）。
func (s *Server) ListSensitiveWordHits(ctx context.Context, req *userv1.ListSensitiveWordHitsRequest) (*userv1.ListSensitiveWordHitsResponse, error) {
	if s.sensitive == nil {
		return nil, status.Errorf(codes.Unavailable, "sensitive service not configured")
	}
	data, err := s.sensitive.ListHitsByUID(ctx, int(req.GetUid()), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list hits: %v", err)
	}
	return hitsMapToProto(data), nil
}

func hitsMapToProto(data map[string]interface{}) *userv1.ListSensitiveWordHitsResponse {
	resp := &userv1.ListSensitiveWordHitsResponse{}
	if data == nil {
		return resp
	}
	if pag, ok := data["pagination"].(repo.NestPagination); ok {
		resp.Total = int32(pag.Total)
		resp.Page = int32(pag.Page)
		resp.PageSize = int32(pag.PageSize)
		resp.TotalPages = int32(pag.Pages)
	}
	list, _ := data["list"].([]*ent.SensitiveWordHit)
	items := make([]*userv1.SensitiveWordHitItem, 0, len(list))
	for _, h := range list {
		if h == nil {
			continue
		}
		item := &userv1.SensitiveWordHitItem{
			Id:         int32(h.ID),
			SourceType: h.SourceType,
			SourceId:   h.SourceId,
			Content:    h.Content,
			HitWords:   h.HitWords,
			Status:     h.Status,
			CreateTime: formatHitTime(h.CreateTime),
		}
		if h.UID != nil {
			uid := int32(*h.UID)
			item.Uid = &uid
		}
		if h.IP != nil {
			item.Ip = h.IP
		}
		if h.ReviewerId != nil {
			rid := int32(*h.ReviewerId)
			item.ReviewerId = &rid
		}
		if h.ReviewTime != nil {
			item.ReviewTime = formatHitTime(*h.ReviewTime)
		}
		items = append(items, item)
	}
	resp.List = items
	return resp
}

func formatHitTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func grpcErrFromApp(err error) error {
	if err == nil {
		return nil
	}
	var ec errcode.ErrCode
	if errors.As(err, &ec) {
		switch ec.Code() {
		case errcode.NotFound.Code():
			return status.Errorf(codes.NotFound, "%s", ec.Message())
		case errcode.Forbidden.Code():
			return status.Errorf(codes.PermissionDenied, "%s", ec.Message())
		case errcode.InvalidParam.Code():
			return status.Errorf(codes.InvalidArgument, "%s", ec.Message())
		}
	}
	return status.Errorf(codes.Internal, "%v", err)
}
