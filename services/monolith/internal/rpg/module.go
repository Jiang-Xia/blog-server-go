// Package rpg RPG 模块化单体域：repo + 全部 gameplay 服务 DI 聚合。
package rpg

import (
	blogrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/repo"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/ent"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/blog/ws"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/event"
	rpgachievement "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/achievement"
	rpgactivity "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/activity"
	rpgadmin "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/admin"
	rpgbuff "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/buff"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/core"
	rpgguild "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/guild"
	"github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/inventory"
	rpgleaderboard "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/leaderboard"
	rpglevel "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/level"
	rpglottery "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/lottery"
	rpgnotify "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/notify"
	rpgpet "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/pet"
	rpgprofile "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/profile"
	rpgpunish "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/punishment"
	rpgquest "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/quest"
	rpgrecharge "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/recharge"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/repo"
	rpgsocial "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/rpg/social"
	userrepo "github.com/Jiang-Xia/blog-server-go/services/monolith/internal/user/repo"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// Module RPG 子模块全部服务，供 wire / handler 注入。
type Module struct {
	Repo       *rpgrepo.RpgRepo
	Rpg        *rpgcore.RpgService
	Level      *rpglevel.LevelService
	Sign       *rpglevel.SignService
	Punishment *rpgpunish.PunishmentService
	Notify     *rpgnotify.RpgNotifyService

	Inventory   *inventory.Service
	Buff        *rpgbuff.Service
	Weather     *rpgbuff.WeatherService
	Achievement *rpgachievement.Service
	Quest       *rpgquest.Service
	Lottery     *rpglottery.Service
	Pet         *rpgpet.Service
	Leaderboard *rpgleaderboard.Service
	Guild       *rpgguild.Service
	Activity    *rpgactivity.Service
	Reputation  *rpgsocial.ReputationService
	Tip         *rpgsocial.TipService
	Social      *rpgsocial.InteractService
	Recharge    *rpgrecharge.Service
	Admin       *rpgadmin.Service
	Profile     *rpgprofile.Service
}

// NewModule 装配 RPG 子模块全部依赖。
func NewModule(
	client *ent.Client,
	pusher ws.Pusher,
	redis *redisutil.Store,
	rds rueidis.Client,
	users *userrepo.UserRepo,
	articles *blogrepo.ArticleRepo,
	publisher *event.Publisher,
	hub *ws.Hub,
	log *zap.Logger,
) *Module {
	repo := rpgrepo.NewRpgRepo(client)
	rpgSvc := rpgcore.NewRpgService(repo)

	var online rpgnotify.OnlineUIDsProvider
	if hub != nil {
		online = hub.OnlineUIDs
	}
	notify := rpgnotify.NewRpgNotifyService(pusher, redis, users, online)

	levelSvc := rpglevel.NewLevelService(rpgSvc, repo, notify, redis)
	signSvc := rpglevel.NewSignService(rpgSvc, levelSvc)
	punishSvc := rpgpunish.NewPunishmentService(rpgSvc)

	inventorySvc := inventory.NewService(repo, rpgSvc, log)
	buffSvc := rpgbuff.NewService(repo)
	weatherSvc := rpgbuff.NewWeatherService()
	achievementSvc := rpgachievement.NewService(repo, levelSvc, inventorySvc, rpgSvc, log)

	questSvc := rpgquest.NewService(repo, rpgSvc, levelSvc, inventorySvc, nil)
	lotterySvc := rpglottery.NewService(repo, rpgSvc, inventorySvc, levelSvc, buffSvc, questSvc)
	questSvc.SetLottery(lotterySvc)

	petSvc := rpgpet.NewService(repo, inventorySvc, questSvc, log)
	leaderboardSvc := rpgleaderboard.NewService(repo, users, rds)
	guildSvc := rpgguild.NewService(repo, achievementSvc, questSvc)
	activitySvc := rpgactivity.NewService(repo, publisher, log)
	reputationSvc := rpgsocial.NewReputationService(rpgSvc, achievementSvc)
	tipSvc := rpgsocial.NewTipService(articles, repo, inventorySvc, reputationSvc, publisher, notify)
	socialSvc := rpgsocial.NewInteractService(repo, rpgSvc, inventorySvc, reputationSvc, redis, achievementSvc, questSvc)
	rechargeSvc := rpgrecharge.NewService(repo, rpgSvc, inventorySvc, notify)
	adminSvc := rpgadmin.NewService(repo, rpgSvc, inventorySvc, lotterySvc, guildSvc, punishSvc, "./public/uploads/", "/static/")
	profileSvc := rpgprofile.NewService(users, repo, rpgSvc, inventorySvc, achievementSvc)

	return &Module{
		Repo:        repo,
		Rpg:         rpgSvc,
		Level:       levelSvc,
		Sign:        signSvc,
		Punishment:  punishSvc,
		Notify:      notify,
		Inventory:   inventorySvc,
		Buff:        buffSvc,
		Weather:     weatherSvc,
		Achievement: achievementSvc,
		Quest:       questSvc,
		Lottery:     lotterySvc,
		Pet:         petSvc,
		Leaderboard: leaderboardSvc,
		Guild:       guildSvc,
		Activity:    activitySvc,
		Reputation:  reputationSvc,
		Tip:         tipSvc,
		Social:      socialSvc,
		Recharge:    rechargeSvc,
		Admin:       adminSvc,
		Profile:     profileSvc,
	}
}
