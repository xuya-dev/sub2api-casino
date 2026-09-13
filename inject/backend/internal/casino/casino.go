// Package casino sub2api 娱乐场（大转盘 / 老虎机 / 刮刮乐 / 21点 / 骰宝 / 百家乐）内部包。
//
// 由独立插件改造为源码注入形态：
//   - 不再持有自有登录/会话/JWT，身份一律取自主站鉴权中间件注入的上下文
//     （JWTAuthMiddleware 注入的 AuthSubject 与用户角色）；
//   - 路由统一挂在 v1（/api/v1）之下的 /casino 组；
//   - 响应使用主站标准信封 {code,message,data}（internal/pkg/response）；
//   - 数据库复用主站连接配置（cfg.Database.DSN()，pgx 直连），
//     自有账本表 casino_bets / casino_blackjack_games / casino_settings。
package casino

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/casino/gamemgr"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	// billingBalanceKeyPrefix 主站网关余额缓存键前缀（与 repository 的 BillingCache 保持一致）。
	billingBalanceKeyPrefix = "billing:balance:"
	// defaultLimitPerSecond 每用户游戏接口默认限流（次/秒）。
	defaultLimitPerSecond = 5
	// blackjackStaleHours 21点牌局闲置超过该时长按"自动停牌"规则结算。
	blackjackStaleHours = 12
	// settleExpiredInterval 超时牌局后台巡检间隔。
	settleExpiredInterval = 10 * time.Minute
)

// Server 娱乐场服务：存储层 + 配置管理器 + 可选 Redis。
type Server struct {
	st          *Store
	mgr         *gamemgr.Manager
	rdb         *redis.Client // 可为 nil：不可用时仅跳过余额缓存失效
	limitPerSec int           // 每用户游戏接口限流（次/秒），默认 5

	limMu   sync.Mutex
	limHits map[int64][]time.Time
}

// RegisterCasinoRoutes 把娱乐场路由挂到 v1（/api/v1）之下：
//   - /casino/**：登录用户（jwtAuth）；
//   - /casino/admin/**：管理员（adminAuth，中间件自带管理员角色校验）。
//
// 首次调用会建立数据库连接池、创建娱乐场账本表并加载游戏配置。
func RegisterCasinoRoutes(
	v1 *gin.RouterGroup,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	cfg *config.Config,
	rdb *redis.Client,
) error {
	ctx := context.Background()
	st, err := NewStore(ctx, cfg.Database.DSN())
	if err != nil {
		return err
	}
	if err := st.EnsureSchema(ctx); err != nil {
		return err
	}
	mgr, err := gamemgr.New(ctx, st)
	if err != nil {
		return err
	}
	srv := &Server{st: st, mgr: mgr, rdb: rdb, limitPerSec: defaultLimitPerSecond}

	// 公共状态端点（无需登录）：侧边栏据此决定是否展示娱乐场入口。
	v1.GET("/casino/status", srv.HandleStatus)

	pub := v1.Group("/casino")
	pub.Use(gin.HandlerFunc(jwtAuth))
	{
		pub.GET("/meta", srv.HandleMeta)
		pub.GET("/me", srv.HandleMe)
		pub.GET("/history", srv.HandleHistory)
		pub.GET("/leaderboard", srv.HandleLeaderboard)
		pub.POST("/games/wheel/spin", srv.rateLimit(), srv.HandleWheelSpin)
		pub.POST("/games/slots/spin", srv.rateLimit(), srv.HandleSlotsSpin)
		pub.POST("/games/sicbo/roll", srv.rateLimit(), srv.HandleSicboRoll)
		pub.POST("/games/baccarat/deal", srv.rateLimit(), srv.HandleBaccaratDeal)
		pub.POST("/games/scratch/reveal", srv.rateLimit(), srv.HandleScratchReveal)
		pub.POST("/blackjack/deal", srv.rateLimit(), srv.HandleBJDeal)
		pub.POST("/blackjack/hit", srv.rateLimit(), srv.HandleBJHit)
		pub.POST("/blackjack/stand", srv.rateLimit(), srv.HandleBJStand)
		pub.POST("/blackjack/double", srv.rateLimit(), srv.HandleBJDouble)
		pub.GET("/blackjack/current", srv.HandleBJCurrent)
	}

	admin := v1.Group("/casino/admin")
	admin.Use(gin.HandlerFunc(adminAuth))
	{
		admin.GET("/config", srv.HandleGetConfig)
		admin.PUT("/config", srv.HandlePutConfig)
		admin.GET("/stats", srv.HandleAdminStats)
	}

	// 后台巡检：闲置超时的 21 点牌局按自动停牌结算（保证断线玩家注金不被永久锁死）
	go srv.settleExpiredLoop()
	return nil
}

// SetLimitPerSecond 调整每用户游戏接口限流阈值（测试用）。
func (s *Server) SetLimitPerSecond(n int) { s.limitPerSec = n }

// settleExpiredLoop 周期结算超时牌局。
func (s *Server) settleExpiredLoop() {
	ticker := time.NewTicker(settleExpiredInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.SettleExpiredGames(blackjackStaleHours)
	}
}

// ---- 中间件与辅助 ----

// rateLimit 游戏接口限流：按用户 ID 的内存滑动窗口（1 秒 limitPerSec 次）。
func (s *Server) rateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		if !s.allow(uid) {
			response.Error(c, http.StatusTooManyRequests, "操作过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// allow 滑动窗口限流判定。
func (s *Server) allow(uid int64) bool {
	s.limMu.Lock()
	defer s.limMu.Unlock()
	if s.limHits == nil {
		s.limHits = make(map[int64][]time.Time)
	}
	now := time.Now()
	hits := s.limHits[uid][:0]
	for _, t := range s.limHits[uid] {
		if now.Sub(t) < time.Second {
			hits = append(hits, t)
		}
	}
	if len(hits) >= s.limitPerSec {
		s.limHits[uid] = hits
		return false
	}
	s.limHits[uid] = append(hits, now)
	return true
}

// currentUserID 从主站鉴权中间件注入的上下文中取用户 ID。
func currentUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		return 0, false
	}
	return subject.UserID, true
}

// currentRole 取当前用户角色（中间件注入；缺失返回空串）。
func currentRole(c *gin.Context) string {
	role, _ := middleware.GetUserRoleFromContext(c)
	return role
}

// gameError 按业务错误类型映射响应（信封 code != 0，message 为中文文案）。
func gameError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInsufficient):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrUserNotFound):
		response.Unauthorized(c, err.Error())
	case errors.Is(err, ErrActiveGame), errors.Is(err, ErrGameNotFound):
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.BadRequest(c, err.Error())
	}
}

// invalidateBalance 尽力失效主站网关的余额缓存，失败仅记日志（主站缓存有 TTL 兜底），
// 对日亏损限额与账本逻辑均无影响。
func (s *Server) invalidateBalance(ctx context.Context, uid int64) {
	if s.rdb == nil {
		return
	}
	if err := s.rdb.Del(ctx, billingBalanceKeyPrefix+strconv.FormatInt(uid, 10)).Err(); err != nil {
		log.Printf("[casino] redis 余额缓存失效失败(不影响下注): %v", err)
	}
}

// rng 每次请求独立随机源（PCG，种子来自 crypto/rand 播种的全局源）。
func rng() *rand.Rand { return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())) }

// HandleStatus GET /api/v1/casino/status —— 娱乐模式是否开启（公共，供前端菜单使用）。
func (s *Server) HandleStatus(c *gin.Context) {
	response.Success(c, gin.H{"enabled": s.mgr.GetConfig().IsEnabled()})
}
