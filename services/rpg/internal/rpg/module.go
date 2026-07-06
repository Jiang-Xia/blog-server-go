// Package rpg RPG 模块化域：repo + 全部 gameplay 服务 DI 聚合。
package rpg

import (
	"github.com/Jiang-Xia/blog-server-go/pkg/blogsvc"
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/articleport"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/event"
	rpgachievement "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/achievement"
	rpgactivity "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/activity"
	rpgadmin "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/admin"
	rpgbuff "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/buff"
	rpgcore "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/core"
	rpgguild "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/guild"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/inventory"
	rpgleaderboard "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/leaderboard"
	rpglevel "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/level"
	rpglottery "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/lottery"
	rpgnotify "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/notify"
	rpgpet "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/pet"
	rpgprofile "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/profile"
	rpgpunish "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/punishment"
	rpgquest "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/quest"
	rpgrecharge "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/recharge"
	rpgrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/repo"
	rpgsocial "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/social"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/userport"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/wspush"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

// Module RPG 子模块全部服务，供 wire / handler 注入。
type Module struct {
	Repo       *rpgrepo.RpgRepo
	Rpg        *rpgcore.RpgService
	Level      *rpglevel.LevelService
	ArticleLevel *rpglevel.ArticleLevelService
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
	pusher wspush.Pusher,
	redis *redisutil.Store,
	rds rueidis.Client,
	users userport.UserReader,
	articles articleport.ArticleReader,
	articleRPG blogsvc.ArticleRPGStore,
	publisher *event.Publisher,
	cfg *config.Config,
	log *zap.Logger,
) *Module {
	repo := rpgrepo.NewRpgRepo(client)
	rpgSvc := rpgcore.NewRpgService(repo)

	notify := rpgnotify.NewRpgNotifyService(pusher, redis, users)
	notify.SetItemConfigReader(repo)

	levelSvc := rpglevel.NewLevelService(rpgSvc, repo, notify, redis)
	signSvc := rpglevel.NewSignService(rpgSvc, levelSvc)

	inventorySvc := inventory.NewService(repo, rpgSvc, log)
	inventorySvc.SetNotify(notify)
	buffSvc := rpgbuff.NewService(repo)
	punishSvc := rpgpunish.NewPunishmentService(rpgSvc, buffSvc, notify)
	weatherSvc := rpgbuff.NewWeatherService()
	achievementSvc := rpgachievement.NewService(repo, levelSvc, inventorySvc, rpgSvc, log)
	achievementSvc.SetNotify(notify)
	levelSvc.SetAchievementTracker(achievementSvc)

	questSvc := rpgquest.NewService(repo, rpgSvc, levelSvc, inventorySvc, nil)
	questSvc.SetNotify(notify)
	lotterySvc := rpglottery.NewService(repo, rpgSvc, inventorySvc, levelSvc, buffSvc, questSvc)
	lotterySvc.SetAchievement(achievementSvc)
	lotterySvc.SetNotify(notify)
	questSvc.SetLottery(lotterySvc)

	petSvc := rpgpet.NewService(repo, inventorySvc, questSvc, log)
	petSvc.SetNotify(notify)
	leaderboardSvc := rpgleaderboard.NewService(repo, users, rds)
	guildSvc := rpgguild.NewService(repo, achievementSvc, questSvc)
	activitySvc := rpgactivity.NewService(repo, publisher, log)
	reputationSvc := rpgsocial.NewReputationService(rpgSvc, achievementSvc)
	articleLevelSvc := rpglevel.NewArticleLevelService(articleRPG, reputationSvc, achievementSvc)
	articleLevelSvc.SetNotify(notify)
	tipSvc := rpgsocial.NewTipService(articles, repo, inventorySvc, reputationSvc, articleLevelSvc, publisher, notify)
	socialSvc := rpgsocial.NewInteractService(repo, rpgSvc, inventorySvc, reputationSvc, redis, achievementSvc, questSvc)
	socialSvc.SetNotify(notify)
	rechargeSvc := rpgrecharge.NewService(repo, rpgSvc, inventorySvc, notify)
	adminSvc := rpgadmin.NewService(repo, rpgSvc, inventorySvc, lotterySvc, guildSvc, punishSvc, storageUploadRoot(cfg), cfg.Storage.PublicPrefixOrDefault())
	profileSvc := rpgprofile.NewService(users, repo, rpgSvc, inventorySvc, achievementSvc)

	return &Module{
		Repo:        repo,
		Rpg:         rpgSvc,
		Level:       levelSvc,
		ArticleLevel: articleLevelSvc,
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

func storageUploadRoot(cfg *config.Config) string {
	if cfg == nil || cfg.Storage.UploadPath == "" {
		return "./public/uploads/"
	}
	return cfg.Storage.UploadPath
}
