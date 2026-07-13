// Package leaderboard 排行榜（总榜 MySQL + 周期榜 Redis ZSET）。
package leaderboard

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	rpgconst "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/constants"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	userrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/redis/rueidis"
)

// Period 排行榜周期。
type Period string

const (
	PeriodTotal  Period = "total"
	PeriodWeek   Period = "week"
	PeriodMonth  Period = "month"
	PeriodSeason Period = "season"
)

// ScoreType 排行维度。
type ScoreType string

const (
	ScoreExp        ScoreType = "exp"
	ScoreReputation ScoreType = "reputation"
	ScoreCurrency   ScoreType = "currency"
	ScoreLevel      ScoreType = "level"
	ScoreSignDays   ScoreType = "signDays"
)

// Service 排行榜业务。
type Service struct {
	repo  *rpgrepo.RpgRepo
	users *userrepo.UserRepo
	rds   rueidis.Client
}

// NewService 构造排行榜 Service。
func NewService(repo *rpgrepo.RpgRepo, users *userrepo.UserRepo, rds rueidis.Client) *Service {
	return &Service{repo: repo, users: users, rds: rds}
}

// IncrementScore 周期榜累加分数到 Redis ZSET。
func (s *Service) IncrementScore(ctx context.Context, uid int, scoreType ScoreType, delta int, period Period) error {
	if period == PeriodTotal || delta <= 0 || s.rds == nil {
		return nil
	}
	key := s.redisKey(scoreType, period)
	member := fmt.Sprintf("%d", uid)
	resp := s.rds.Do(ctx, s.rds.B().Zincrby().Key(key).Increment(float64(delta)).Member(member).Build())
	if err := resp.Error(); err != nil {
		return err
	}
	ttl := 86400 * 14
	if period == PeriodMonth {
		ttl = 86400 * 45
	} else if period == PeriodSeason {
		ttl = 86400 * 120
	}
	_ = s.rds.Do(ctx, s.rds.B().Expire().Key(key).Seconds(int64(ttl)).Build()).Error()
	return nil
}

// GetLeaderboard 获取 Top N 排行榜。
func (s *Service) GetLeaderboard(ctx context.Context, scoreType ScoreType, period Period, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if period != PeriodTotal && s.rds != nil {
		if rows, err := s.getFromRedis(ctx, scoreType, period, limit); err == nil && len(rows) > 0 {
			return s.enrichEntries(ctx, rows, scoreType)
		}
	}
	return s.getFromMySQL(ctx, scoreType, limit)
}

func (s *Service) getFromRedis(ctx context.Context, scoreType ScoreType, period Period, limit int) ([]entry, error) {
	key := s.redisKey(scoreType, period)
	resp := s.rds.Do(ctx, s.rds.B().Zrevrange().Key(key).Start(0).Stop(int64(limit-1)).Withscores().Build())
	arr, err := resp.AsZScores()
	if err != nil {
		return nil, err
	}
	out := make([]entry, 0, len(arr))
	for _, z := range arr {
		var uid int
		fmt.Sscanf(z.Member, "%d", &uid)
		out = append(out, entry{UID: uid, Score: int(z.Score)})
	}
	return out, nil
}

func (s *Service) getFromMySQL(ctx context.Context, scoreType ScoreType, limit int) ([]map[string]interface{}, error) {
	if scoreType == ScoreCurrency {
		rows, err := s.repo.ListCurrencyLeaderboard(ctx, rpgconst.CurrencyItemCode, limit)
		if err != nil {
			return nil, err
		}
		entries := make([]entry, 0, len(rows))
		for _, r := range rows {
			entries = append(entries, entry{UID: r.UID, Score: r.Currency})
		}
		return s.enrichEntries(ctx, entries, scoreType)
	}
	field := "exp"
	switch scoreType {
	case ScoreLevel:
		field = "level"
	case ScoreReputation:
		field = "reputation"
	case ScoreSignDays:
		field = "signDays"
	}
	rpgs, err := s.repo.ListRpgOrderBy(ctx, field, limit)
	if err != nil {
		return nil, err
	}
	entries := make([]entry, 0, len(rpgs))
	for _, r := range rpgs {
		score := r.Exp
		switch scoreType {
		case ScoreLevel:
			score = r.Level
		case ScoreReputation:
			score = r.Reputation
		case ScoreSignDays:
			score = r.TotalSignDays
		}
		entries = append(entries, entry{UID: r.UID, Score: score})
	}
	return s.enrichEntries(ctx, entries, scoreType)
}

type entry struct {
	UID   int
	Score int
}

func (s *Service) enrichEntries(ctx context.Context, entries []entry, scoreType ScoreType) ([]map[string]interface{}, error) {
	uids := make([]int, 0, len(entries))
	for _, e := range entries {
		uids = append(uids, e.UID)
	}
	rpgs, _ := s.repo.ListRpgByUIDs(ctx, uids)
	rpgMap := map[int]*ent.Rpg{}
	for _, r := range rpgs {
		rpgMap[r.UID] = r
	}
	out := make([]map[string]interface{}, 0, len(entries))
	for i, e := range entries {
		rpg := rpgMap[e.UID]
		level := 1
		exp := 0
		rep := 0
		signDays := 0
		if rpg != nil {
			level = rpg.Level
			exp = rpg.Exp
			rep = rpg.Reputation
			signDays = rpg.TotalSignDays
		}
		nickname := "匿名用户"
		avatar := ""
		if s.users != nil {
			if u, err := s.users.FindByID(ctx, e.UID); err == nil && u != nil {
				if u.Nickname != "" {
					nickname = u.Nickname
				}
				avatar = u.Avatar
			}
		}
		out = append(out, map[string]interface{}{
			"rank":          i + 1,
			"uid":           e.UID,
			"nickname":      nickname,
			"avatar":        avatar,
			"level":         level,
			"exp":           exp,
			"reputation":    rep,
			"totalSignDays": signDays,
			"score":         e.Score,
			"type":          scoreType,
		})
	}
	return out, nil
}

func (s *Service) redisKey(scoreType ScoreType, period Period) string {
	now := time.Now()
	pk := "total"
	switch period {
	case PeriodWeek:
		y, w := now.ISOWeek()
		pk = fmt.Sprintf("%d-W%02d", y, w)
	case PeriodMonth:
		pk = now.Format("2006-01")
	case PeriodSeason:
		pk = fmt.Sprintf("season_%d", now.Year())
	}
	return fmt.Sprintf("lb:%s:%s", scoreType, pk)
}
