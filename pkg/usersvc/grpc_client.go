// Package usersvc 提供 UserService 端口的 gRPC 远程实现。
package usersvc

import (
	"context"
	"fmt"
	"sync"

	userv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type grpcUserService struct {
	client userv1.UserServiceClient
}

var (
	grpcUserOnce sync.Once
	grpcUserInst *grpcUserService
	grpcUserErr  error
)

// NewGRPCUserService 连接 user-service gRPC 并返回 CrossClient。
func NewGRPCUserService(addr string) (CrossClient, error) {
	grpcUserOnce.Do(func() {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			grpcUserErr = fmt.Errorf("dial user grpc %s: %w", addr, err)
			return
		}
		grpcUserInst = &grpcUserService{client: userv1.NewUserServiceClient(conn)}
	})
	if grpcUserErr != nil {
		return nil, grpcUserErr
	}
	return grpcUserInst, nil
}

func (g *grpcUserService) GetUser(ctx context.Context, id uint64) (*UserDTO, error) {
	resp, err := g.client.GetUser(ctx, &userv1.GetUserRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protoToDTO(resp), nil
}

func (g *grpcUserService) GetUserBatch(ctx context.Context, ids []uint64) ([]*UserDTO, error) {
	resp, err := g.client.GetUserBatch(ctx, &userv1.GetUserBatchRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	out := make([]*UserDTO, 0, len(resp.GetUsers()))
	for _, u := range resp.GetUsers() {
		out = append(out, protoToDTO(u))
	}
	return out, nil
}

func (g *grpcUserService) EvaluateContent(ctx context.Context, content string) (*FilterEvaluateResult, error) {
	resp, err := g.client.EvaluateContent(ctx, &userv1.EvaluateContentRequest{Content: content})
	if err != nil {
		return nil, err
	}
	return &FilterEvaluateResult{
		Content:    resp.GetContent(),
		HitWords:   resp.GetHitWords(),
		HpPenalty:  int(resp.GetHpPenalty()),
		NeedReview: resp.GetNeedReview(),
		Rejected:   resp.GetRejected(),
	}, nil
}

func (g *grpcUserService) CreateHitRecord(ctx context.Context, params FilterHitParams) error {
	req := &userv1.CreateHitRecordRequest{
		SourceType: params.SourceType,
		SourceId:   params.SourceID,
		Content:    params.Content,
		HitWords:   params.HitWords,
	}
	if params.UID != nil {
		uid := int32(*params.UID)
		req.Uid = &uid
	}
	if params.IP != nil {
		req.Ip = params.IP
	}
	_, err := g.client.CreateHitRecord(ctx, req)
	return err
}

func (g *grpcUserService) ListActiveUserIDs(ctx context.Context) ([]int, error) {
	resp, err := g.client.ListActiveUserIDs(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(resp.GetUserIds()))
	for _, id := range resp.GetUserIds() {
		out = append(out, int(id))
	}
	return out, nil
}

func (g *grpcUserService) GetDept(ctx context.Context, id int) (*DeptDTO, error) {
	resp, err := g.client.GetDept(ctx, &userv1.GetDeptRequest{Id: int32(id)})
	if err != nil {
		return nil, err
	}
	return &DeptDTO{ID: int(resp.GetId()), DeptName: resp.GetDeptName()}, nil
}

func (g *grpcUserService) ResolveArticleAccessibleDeptIDs(ctx context.Context, uid int) ([]int, error) {
	resp, err := g.client.ResolveAccessibleDeptIDs(ctx, &userv1.ResolveAccessibleDeptIDsRequest{
		Uid:          int32(uid),
		ResourceType: "article",
	})
	if err != nil {
		return nil, err
	}
	if resp.GetUnrestricted() {
		return nil, nil
	}
	out := make([]int, 0, len(resp.GetDeptIds()))
	for _, id := range resp.GetDeptIds() {
		out = append(out, int(id))
	}
	return out, nil
}

func (g *grpcUserService) AssertArticleDeptAccess(ctx context.Context, uid int, deptID *int) error {
	req := &userv1.AssertDeptAccessRequest{
		Uid:          int32(uid),
		ResourceType: "article",
	}
	if deptID != nil {
		d := int32(*deptID)
		req.DeptId = &d
	}
	_, err := g.client.AssertDeptAccess(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.PermissionDenied {
			return fmt.Errorf("%s", st.Message())
		}
	}
	return err
}

func (g *grpcUserService) ListSensitiveWordHits(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error) {
	resp, err := g.client.ListSensitiveWordHits(ctx, &userv1.ListSensitiveWordHitsRequest{
		Uid:      int32(uid),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(resp.GetList()))
	for _, h := range resp.GetList() {
		item := map[string]interface{}{
			"id":         h.GetId(),
			"sourceType": h.GetSourceType(),
			"sourceId":   h.GetSourceId(),
			"content":    h.GetContent(),
			"hitWords":   h.GetHitWords(),
			"status":     h.GetStatus(),
			"createTime": h.GetCreateTime(),
		}
		if h.Uid != nil {
			item["uid"] = h.GetUid()
		}
		if h.Ip != nil {
			item["ip"] = h.GetIp()
		}
		if h.ReviewerId != nil {
			item["reviewerId"] = h.GetReviewerId()
		}
		if h.GetReviewTime() != "" {
			item["reviewTime"] = h.GetReviewTime()
		}
		list = append(list, item)
	}
	return map[string]interface{}{
		"list": list,
		"pagination": map[string]int{
			"total":      int(resp.GetTotal()),
			"page":       int(resp.GetPage()),
			"pageSize":   int(resp.GetPageSize()),
			"totalPages": int(resp.GetTotalPages()),
		},
	}, nil
}

func protoToDTO(u *userv1.GetUserResponse) *UserDTO {
	if u == nil {
		return nil
	}
	dto := &UserDTO{
		ID:       u.GetId(),
		Nickname: u.GetNickname(),
		Username: u.GetUsername(),
		Avatar:   u.GetAvatar(),
		Email:    u.GetEmail(),
		Status:   u.GetStatus(),
	}
	if u.DeptId != nil {
		deptID := int(u.GetDeptId())
		dto.DeptID = &deptID
	}
	return dto
}
