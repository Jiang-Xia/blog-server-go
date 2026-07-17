// Package app wire provider 集合。
package app

import (
	"database/sql"

	"github.com/Jiang-Xia/blog-server-go/pkg/blogsvc"
	"github.com/Jiang-Xia/blog-server-go/pkg/config"
	"github.com/Jiang-Xia/blog-server-go/pkg/publicprofile"
	"github.com/Jiang-Xia/blog-server-go/pkg/redisutil"
	"github.com/Jiang-Xia/blog-server-go/pkg/usersvc"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/ent"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/articleport"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/auth"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/data"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/event"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/handler"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/middleware"
	payrepo "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/pay/repo"
	paysvc "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/pay/service"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg"
	rpgactivity "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/activity"
	rpgevent "github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/event"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/rpg/kitexserver"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/scheduler"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/userport"
	"github.com/Jiang-Xia/blog-server-go/services/rpg/internal/wspush"
	"github.com/redis/rueidis"
	"go.uber.org/zap"
)

func provideDataStore(cfg *config.Config, log *zap.Logger) (*data.Data, error) {
	return data.NewData(cfg, log)
}

func provideEntClient(d *data.Data) *ent.Client {
	return d.Ent
}

func provideSQLDB(d *data.Data) *sql.DB {
	return d.SQL
}

func provideRedisClient(d *data.Data) rueidis.Client {
	return d.Redis
}

func provideRedisStore(client rueidis.Client) *redisutil.Store {
	return redisutil.New(client)
}

func provideUserService(cfg *config.Config) (usersvc.CrossClient, error) {
	return userport.ProvideUserService(cfg)
}

func provideUserReader(users usersvc.UserService) userport.UserReader {
	return userport.NewGRPCUserReader(users)
}

func provideArticleReader(db *sql.DB) articleport.ArticleReader {
	return articleport.NewSQLArticleReader(db)
}

func provideBlogPublicProfileLister(cfg *config.Config) (blogsvc.PublicProfileLister, error) {
	return blogsvc.NewKitexPublicProfileLister(cfg.Registry.EtcdEndpointsOrEmpty())
}

func providePublicProfileRepo(cfg *config.Config) (*publicprofile.Repo, error) {
	return publicprofile.NewRepo(cfg)
}

func provideBlogArticleRPGStore(cfg *config.Config) (blogsvc.ArticleRPGStore, error) {
	return blogsvc.NewKitexArticleRPGStore(cfg.Registry.EtcdEndpointsOrEmpty())
}

func provideWSPusher(rds rueidis.Client) wspush.Pusher {
	return wspush.NewRedisPusher(rds)
}

func provideJWT(cfg *config.Config) *auth.JWTService {
	return auth.NewJWTService(cfg)
}

func providePayOrderRepo(client *ent.Client) *payrepo.PayOrderRepo {
	return payrepo.NewPayOrderRepo(client)
}

func providePayService(cfg *config.Config, orderRepo *payrepo.PayOrderRepo, log *zap.Logger) (*paysvc.PayService, error) {
	return paysvc.NewPayService(cfg, orderRepo, log)
}

func providePayOrderService(orderRepo *payrepo.PayOrderRepo, pay *paysvc.PayService, mod *rpg.Module, log *zap.Logger) *paysvc.PayOrderService {
	return providePayOrderServiceWithRecharge(orderRepo, pay, mod, log)
}

func provideRPGKitexServer(mod *rpg.Module) *kitexserver.Server {
	if mod == nil {
		return kitexserver.New(nil, nil, nil)
	}
	return kitexserver.New(mod.Rpg, mod.Profile, mod.Punishment)
}

func provideRPGHandler(mod *rpg.Module, game *handler.RPGGameplay, jwt *auth.JWTService) *handler.RPGHandler {
	return handler.NewRPGHandler(mod, game, jwt)
}

func provideActivityNotifyScheduler(mod *rpg.Module, log *zap.Logger) scheduler.ActivityNotifyRunner {
	if mod == nil {
		return nil
	}
	return rpgactivity.NewNotifyScheduler(mod.Repo, mod.Notify, log)
}

func provideRPGEventHandlers(mod *rpg.Module, redis *redisutil.Store) rpgevent.Handlers {
	return provideRPGEventHandlersFull(mod, redis)
}

func provideRPGEventConsumer(rds rueidis.Client, log *zap.Logger, handlers rpgevent.Handlers) RPGEventConsumer {
	c := event.NewConsumer(rds, log, event.ConsumerGroupRPG)
	rpgevent.RegisterRPGHandlers(c, handlers)
	return RPGEventConsumer{c}
}

func provideRegisterDeps(
	health *handler.HealthHandler,
	rpgH *handler.RPGHandler,
	rpgAdmin *handler.RPGAdminHandler,
	rpgProfile *handler.RPGProfileHandler,
	payH *handler.PayHandler,
	payOrder *handler.PayOrderHandler,
	jwt *auth.JWTService,
	users usersvc.UserService,
	cfg *config.Config,
	redis *redisutil.Store,
	log *zap.Logger,
) handler.RegisterDeps {
	return handler.RegisterDeps{
		Health:     health,
		RPG:        rpgH,
		RPGAdmin:   rpgAdmin,
		RPGProfile: rpgProfile,
		Pay:        payH,
		PayOrder:   payOrder,
		JWT:        jwt,
		Users:      users,
		Permission: middleware.PermissionDeps{Cfg: cfg, Redis: redis, JWT: jwt, Log: log},
	}
}

// RPGEventConsumer RPG Stream 消费器。
type RPGEventConsumer struct {
	*event.Consumer
}
