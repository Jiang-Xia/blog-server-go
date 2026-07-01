// Package guild 公会创建、加入与成员管理。
package guild

import (
	"context"
	"time"

	"github.com/Jiang-Xia/blog-server-go/pkg/errcode"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
)

// AchievementTracker 成就/任务追踪。
type AchievementTracker interface {
	TrackProgress(ctx context.Context, uid int, event string) error
}

// Service 公会业务。
type Service struct {
	repo         *rpgrepo.RpgRepo
	achievement  AchievementTracker
	questTracker AchievementTracker
}

// NewService 构造公会 Service。
func NewService(repo *rpgrepo.RpgRepo, achievement, quest AchievementTracker) *Service {
	return &Service{repo: repo, achievement: achievement, questTracker: quest}
}

// List 公会列表。
func (s *Service) List(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	rows, total, err := s.repo.ListGuilds(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}
	list := make([]map[string]interface{}, 0, len(rows))
	for _, g := range rows {
		list = append(list, map[string]interface{}{
			"id":           g.ID,
			"name":         g.Name,
			"leaderUid":    g.LeaderUid,
			"memberCount":  g.MemberCount,
			"announcement": g.Announcement,
		})
	}
	return map[string]interface{}{"list": list, "total": total, "page": page, "pageSize": pageSize}, nil
}

// GetMy 当前用户所属公会。
func (s *Service) GetMy(ctx context.Context, uid int) (map[string]interface{}, error) {
	member, err := s.repo.FindGuildMemberByUID(ctx, uid)
	if ent.IsNotFound(err) {
		return map[string]interface{}{"guild": nil}, nil
	}
	if err != nil {
		return nil, err
	}
	return s.GetDetail(ctx, member.GuildId, uid)
}

// GetDetail 公会详情与成员。
func (s *Service) GetDetail(ctx context.Context, guildID, viewerUID int) (map[string]interface{}, error) {
	guild, err := s.repo.FindGuildByID(ctx, guildID)
	if err != nil {
		return nil, errcode.WithMessage(errcode.NotFound, "公会不存在")
	}
	members, err := s.repo.ListGuildMembers(ctx, guildID)
	if err != nil {
		return nil, err
	}
	memberList := make([]map[string]interface{}, 0, len(members))
	myRole := ""
	for _, m := range members {
		if m.UID == viewerUID {
			myRole = m.Role
		}
		memberList = append(memberList, map[string]interface{}{
			"uid":      m.UID,
			"role":     m.Role,
			"joinTime": m.JoinTime,
		})
	}
	return map[string]interface{}{
		"guild": map[string]interface{}{
			"id":           guild.ID,
			"name":         guild.Name,
			"leaderUid":    guild.LeaderUid,
			"memberCount":  guild.MemberCount,
			"announcement": guild.Announcement,
		},
		"members": memberList,
		"myRole":  myRole,
	}, nil
}

// Create 创建公会。
func (s *Service) Create(ctx context.Context, uid int, name, announcement string) (*ent.RpgGuild, error) {
	if _, err := s.repo.FindGuildMemberByUID(ctx, uid); err == nil {
		return nil, errcode.WithMessage(errcode.Conflict, "已在公会中")
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	if _, err := s.repo.FindGuildByName(ctx, name); err == nil {
		return nil, errcode.WithMessage(errcode.Conflict, "公会名称已存在")
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	var created *ent.RpgGuild
	err := s.repo.WithTx(ctx, func(tx *ent.Tx) error {
		b := tx.RpgGuild.Create().
			SetName(name).
			SetLeaderUid(uid).
			SetMemberCount(1)
		if announcement != "" {
			b.SetAnnouncement(announcement)
		}
		g, err := b.Save(ctx)
		if err != nil {
			return err
		}
		_, err = tx.RpgUserGuildMember.Create().
			SetGuildId(g.ID).
			SetUID(uid).
			SetRole("leader").
			SetJoinTime(time.Now()).
			Save(ctx)
		if err != nil {
			return err
		}
		created = g
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.achievement != nil {
		_ = s.achievement.TrackProgress(ctx, uid, "guild_create")
	}
	s.trackGuildJoin(ctx, uid)
	return created, nil
}

// Join 加入公会。
func (s *Service) Join(ctx context.Context, uid, guildID int) (*ent.RpgGuild, error) {
	if _, err := s.repo.FindGuildMemberByUID(ctx, uid); err == nil {
		return nil, errcode.WithMessage(errcode.Conflict, "已在公会中")
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	guild, err := s.repo.FindGuildByID(ctx, guildID)
	if err != nil {
		return nil, errcode.WithMessage(errcode.NotFound, "公会不存在")
	}
	_, err = s.repo.CreateGuildMember(ctx, &ent.RpgUserGuildMember{
		GuildId:  guildID,
		UID:      uid,
		Role:     "member",
		JoinTime: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	guild.MemberCount++
	guild, err = s.repo.UpdateGuild(ctx, guild)
	if err != nil {
		return nil, err
	}
	s.trackGuildJoin(ctx, uid)
	return guild, nil
}

// Leave 退出公会。
func (s *Service) Leave(ctx context.Context, uid int) error {
	member, err := s.repo.FindGuildMemberByUID(ctx, uid)
	if err != nil {
		return errcode.WithMessage(errcode.NotFound, "未加入公会")
	}
	if member.Role == "leader" {
		return errcode.WithMessage(errcode.InvalidParam, "会长请先转让会长或解散公会")
	}
	guild, err := s.repo.FindGuildByID(ctx, member.GuildId)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteGuildMember(ctx, member.ID); err != nil {
		return err
	}
	guild.MemberCount--
	if guild.MemberCount < 0 {
		guild.MemberCount = 0
	}
	_, err = s.repo.UpdateGuild(ctx, guild)
	return err
}

func (s *Service) trackGuildJoin(ctx context.Context, uid int) {
	if s.achievement != nil {
		_ = s.achievement.TrackProgress(ctx, uid, "guild_join")
	}
	if s.questTracker != nil {
		_ = s.questTracker.TrackProgress(ctx, uid, "guild_join")
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
