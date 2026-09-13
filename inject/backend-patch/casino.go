package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/casino"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterCasinoRoutes 注册娱乐场模块路由（源码注入形态的薄封装层）。
//
// 全部路由挂 v1（/api/v1）之下的 /casino 组：
//   - /casino/**：登录用户（jwtAuth）；
//   - /casino/admin/**：管理员（adminAuth，自带管理员角色校验）。
//
// 首次注册会建立 pgx 连接池（复用主站 cfg.Database.DSN()）、
// 幂等创建娱乐场账本表并加载游戏配置；失败返回错误由调用方记日志兜底。
func RegisterCasinoRoutes(
	v1 *gin.RouterGroup,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	cfg *config.Config,
	rdb *redis.Client,
) error {
	return casino.RegisterCasinoRoutes(v1, jwtAuth, adminAuth, cfg, rdb)
}
