// Package usersvc 系统邮件 Kitex 客户端。
package usersvc

import (
	"context"
	"fmt"
	"sync"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/kitexreg"
	userv1 "github.com/Jiang-Xia/blog-server-go/proto/kitex/user/v1"
	"github.com/Jiang-Xia/blog-server-go/proto/kitex/user/v1/userservice"
	"github.com/cloudwego/kitex/client"
)

type kitexSystemEmailSender struct {
	client userservice.Client
}

var (
	emailOnce sync.Once
	emailInst SystemEmailSender
	emailErr  error
)

// NewKitexSystemEmailSender 经 Nacos 发现 user-service 并返回 SystemEmailSender。
func NewKitexSystemEmailSender(reg config.RegistryConfig) (SystemEmailSender, error) {
	if !reg.Enabled() {
		return nil, fmt.Errorf("registry.nacos_addr required for system email")
	}
	emailOnce.Do(func() {
		r, err := kitexreg.NewResolver(reg)
		if err != nil {
			emailErr = err
			return
		}
		cli, err := userservice.NewClient(config.KitexServiceUser, client.WithResolver(r))
		if err != nil {
			emailErr = fmt.Errorf("new user kitex client: %w", err)
			return
		}
		emailInst = &kitexSystemEmailSender{client: cli}
	})
	if emailErr != nil {
		return nil, emailErr
	}
	return emailInst, nil
}

func (g *kitexSystemEmailSender) SendSystemEmail(ctx context.Context, to, subject, htmlBody string) (bool, error) {
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
