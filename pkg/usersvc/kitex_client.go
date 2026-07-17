// Package usersvc 提供 UserService 端口的 Kitex + etcd 远程实现。
package usersvc

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexreg"
	userv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/user/v1"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/user/v1/userservice"
	"github.com/cloudwego/kitex/client"
	"google.golang.org/protobuf/types/known/emptypb"
)

type kitexUserService struct {
	client userservice.Client
}

var (
	kitexUserOnce sync.Once
	kitexUserInst *kitexUserService
	kitexUserErr  error
)

// NewKitexUserService 经 etcd 发现 user-service 并返回 CrossClient。
// endpoints 为空时返回错误（blog/rpg 微服务必须配置 registry.etcd_endpoints）。
func NewKitexUserService(endpoints []string) (CrossClient, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("registry.etcd_endpoints required for user Kitex client")
	}
	kitexUserOnce.Do(func() {
		r, err := kitexreg.NewResolver(endpoints)
		if err != nil {
			kitexUserErr = err
			return
		}
		cli, err := userservice.NewClient(config.KitexServiceUser, client.WithResolver(r))
		if err != nil {
			kitexUserErr = fmt.Errorf("new user kitex client: %w", err)
			return
		}
		kitexUserInst = &kitexUserService{client: cli}
	})
	if kitexUserErr != nil {
		return nil, kitexUserErr
	}
	return kitexUserInst, nil
}

func (g *kitexUserService) GetUser(ctx context.Context, id uint64) (*UserDTO, error) {
	resp, err := g.client.GetUser(ctx, &userv1.GetUserRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protoToDTO(resp), nil
}

func (g *kitexUserService) GetUserBatch(ctx context.Context, ids []uint64) ([]*UserDTO, error) {
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

func (g *kitexUserService) EvaluateContent(ctx context.Context, content string) (*FilterEvaluateResult, error) {
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

func (g *kitexUserService) CreateHitRecord(ctx context.Context, params FilterHitParams) error {
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

func (g *kitexUserService) ListActiveUserIDs(ctx context.Context) ([]int, error) {
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

func (g *kitexUserService) GetDept(ctx context.Context, id int) (*DeptDTO, error) {
	resp, err := g.client.GetDept(ctx, &userv1.GetDeptRequest{Id: int32(id)})
	if err != nil {
		return nil, err
	}
	return &DeptDTO{ID: int(resp.GetId()), DeptName: resp.GetDeptName()}, nil
}

func (g *kitexUserService) ResolveArticleAccessibleDeptIDs(ctx context.Context, uid int) ([]int, error) {
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

func (g *kitexUserService) AssertArticleDeptAccess(ctx context.Context, uid int, deptID *int) error {
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
		if strings.Contains(strings.ToLower(err.Error()), "permission denied") {
			return fmt.Errorf("%s", err.Error())
		}
	}
	return err
}

func (g *kitexUserService) ListSensitiveWordHits(ctx context.Context, uid, page, pageSize int) (map[string]interface{}, error) {
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
