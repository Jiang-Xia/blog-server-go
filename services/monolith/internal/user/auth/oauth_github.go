package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHubOAuth GitHub OAuth2 客户端配置与回调处理。
type GitHubOAuth struct {
	cfg    *config.Config
	oauth  *oauth2.Config
	auth   *AuthService
	users  *repo.UserRepo
}

// NewGitHubOAuth 从 config.OAuth 构造 GitHub OAuth 助手。
func NewGitHubOAuth(cfg *config.Config, authSvc *AuthService, users *repo.UserRepo) *GitHubOAuth {
	o := cfg.OAuth
	oauthCfg := &oauth2.Config{
		ClientID:     o.GithubClientID,
		ClientSecret: o.GithubClientSecret,
		RedirectURL:  o.GithubCallbackURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
	return &GitHubOAuth{cfg: cfg, oauth: oauthCfg, auth: authSvc, users: users}
}

// AuthCodeURL 生成 GitHub 授权跳转 URL。
func (g *GitHubOAuth) AuthCodeURL(state string) string {
	return g.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// GitHubProfile GitHub 用户资料（HandleCallback 中间类型）。
type GitHubProfile struct {
	ID        string
	Username  string
	Email     string
	AvatarURL string
	ProfileURL string
}

// CallbackResult OAuth 回调结果：token 与用户。
type CallbackResult struct {
	Tokens *TokenTriple
	User   map[string]interface{}
	Ticket string
}

// HandleCallback 用 authorization code 换取 token、查找/创建用户并签发 JWT。
func (g *GitHubOAuth) HandleCallback(ctx context.Context, code string) (*CallbackResult, error) {
	if code == "" {
		return nil, errcode.InvalidParam
	}
	tok, err := g.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, errcode.WithMessage(errcode.Unauthorized, "GitHub 授权失败")
	}
	profile, err := g.fetchGitHubProfile(ctx, tok.AccessToken)
	if err != nil {
		return nil, err
	}
	u, err := g.findOrCreateUser(ctx, profile)
	if err != nil {
		return nil, err
	}
	login, err := g.auth.LoginAfterOAuth(ctx, u)
	if err != nil {
		return nil, err
	}
	tokens := &TokenTriple{
		Token:        login.Token,
		AccessToken:  login.AccessToken,
		RefreshToken: login.RefreshToken,
	}
	ticket, err := g.auth.CreateOAuthTicket(ctx, tokens)
	if err != nil {
		return nil, err
	}
	return &CallbackResult{
		Tokens: tokens,
		User:   login.User,
		Ticket: ticket,
	}, nil
}

func (g *GitHubOAuth) findOrCreateUser(ctx context.Context, p *GitHubProfile) (*ent.User, error) {
	if p.ID == "" {
		return nil, errcode.WithMessage(errcode.InvalidParam, "GitHub 用户 ID 为空")
	}
	existing, err := g.users.FindByGithubID(ctx, p.ID)
	if err == nil {
		return g.updateUserFromGitHub(ctx, existing, p)
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	username := p.Username
	if username == "" && p.Email != "" {
		username = p.Email
	}
	nickname := p.Username
	if nickname == "" {
		nickname = p.Email
	}
	in := repo.CreateUserInput{
		Username: username,
		Nickname: nickname,
		Password: RandomPassword(),
		Email:    p.Email,
		Avatar:   p.AvatarURL,
		GithubID: p.ID,
		Homepage: p.ProfileURL,
	}
	return g.auth.CreateUserWithDefaultRole(ctx, in)
}

func (g *GitHubOAuth) updateUserFromGitHub(ctx context.Context, u *ent.User, p *GitHubProfile) (*ent.User, error) {
	fields := map[string]interface{}{}
	if p.Username != "" {
		fields["nickname"] = p.Username
	}
	if p.Email != "" {
		fields["email"] = p.Email
	}
	if p.AvatarURL != "" {
		fields["avatar"] = p.AvatarURL
	}
	if p.ProfileURL != "" {
		fields["homepage"] = p.ProfileURL
	}
	if len(fields) == 0 {
		return u, nil
	}
	return g.users.UpdateFields(ctx, u.ID, fields)
}

func (g *GitHubOAuth) fetchGitHubProfile(ctx context.Context, accessToken string) (*GitHubProfile, error) {
	userBody, err := githubAPIGet(ctx, "https://api.github.com/user", accessToken)
	if err != nil {
		return nil, err
	}
	var userResp struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
		Email     string `json:"email"`
	}
	if err := json.Unmarshal(userBody, &userResp); err != nil {
		return nil, err
	}
	email := userResp.Email
	if email == "" {
		emailsBody, err := githubAPIGet(ctx, "https://api.github.com/user/emails", accessToken)
		if err == nil {
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if json.Unmarshal(emailsBody, &emails) == nil {
				for _, e := range emails {
					if e.Primary && e.Verified {
						email = e.Email
						break
					}
				}
				if email == "" && len(emails) > 0 {
					email = emails[0].Email
				}
			}
		}
	}
	return &GitHubProfile{
		ID:         strconv.FormatInt(userResp.ID, 10),
		Username:   userResp.Login,
		Email:      email,
		AvatarURL:  userResp.AvatarURL,
		ProfileURL: userResp.HTMLURL,
	}, nil
}

func githubAPIGet(ctx context.Context, url, accessToken string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("github api %s: %s", url, string(body))
	}
	return body, nil
}
