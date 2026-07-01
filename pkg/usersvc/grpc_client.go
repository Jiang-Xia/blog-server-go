// Package usersvc 提供 UserService 端口的 gRPC 远程实现。
package usersvc

import (
	"context"
	"fmt"
	"sync"

	userv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcUserService struct {
	client userv1.UserServiceClient
}

var (
	grpcUserOnce sync.Once
	grpcUserInst UserService
	grpcUserErr  error
)

// NewGRPCUserService 连接 user-service gRPC 并返回 UserService 实现。
func NewGRPCUserService(addr string) (UserService, error) {
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

func protoToDTO(u *userv1.GetUserResponse) *UserDTO {
	if u == nil {
		return nil
	}
	return &UserDTO{
		ID:       u.GetId(),
		Nickname: u.GetNickname(),
		Username: u.GetUsername(),
		Avatar:   u.GetAvatar(),
		Email:    u.GetEmail(),
		Status:   u.GetStatus(),
	}
}
