// Package usersvc 系统邮件 gRPC 客户端。
package usersvc

import (
	"context"
	"fmt"
	"sync"

	userv1 "github.com/Jiang-Xia/blog-server-go/proto/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcSystemEmailSender struct {
	client userv1.UserServiceClient
}

var (
	emailOnce sync.Once
	emailInst SystemEmailSender
	emailErr  error
)

// NewGRPCSystemEmailSender 连接 user-service gRPC 并返回 SystemEmailSender。
func NewGRPCSystemEmailSender(addr string) (SystemEmailSender, error) {
	emailOnce.Do(func() {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			emailErr = fmt.Errorf("dial user grpc %s: %w", addr, err)
			return
		}
		emailInst = &grpcSystemEmailSender{client: userv1.NewUserServiceClient(conn)}
	})
	if emailErr != nil {
		return nil, emailErr
	}
	return emailInst, nil
}

func (g *grpcSystemEmailSender) SendSystemEmail(ctx context.Context, to, subject, htmlBody string) (bool, error) {
	resp, err := g.client.SendSystemEmail(ctx, &userv1.SendSystemEmailRequest{
		To:       to,
		Subject:  subject,
		HtmlBody: htmlBody,
	})
	if err != nil {
		return false, err
	}
	return resp.GetSent(), nil
}
