package casino

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// HandleMeta 玩家端游戏展示信息（不含奖品档位概率分布，仅暴露中奖率与
// 最高倍数供「最高奖金 = 最高倍数 × 面值」展示；档位表仅管理端可见）。
// requireEnabled 娱乐模式总开关校验；关闭时用户端点一律 403。
func (s *Server) requireEnabled(c *gin.Context) bool {
	if s.mgr.GetConfig().IsEnabled() {
		return true
	}
	response.Error(c, 403, "娱乐模式已关闭")
	return false
}

func (s *Server) HandleMeta(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	cfg := s.mgr.Get()
	segments := make([]map[string]any, 0, len(cfg.Wheel.Segments))
	for _, seg := range cfg.Wheel.Segments {
		segments = append(segments, map[string]any{"multiplier": seg.Multiplier, "label": seg.Label})
	}
	symbols := make([]map[string]any, 0, len(cfg.Slots.Symbols))
	for _, sym := range cfg.Slots.Symbols {
		symbols = append(symbols, map[string]any{"id": sym.ID, "emoji": sym.Emoji})
	}
	lineSymbols := make([]map[string]any, 0, len(cfg.Scratch.Lines.Symbols))
	for _, sym := range cfg.Scratch.Lines.Symbols {
		lineSymbols = append(lineSymbols, map[string]any{"id": sym.ID, "multiplier": sym.Multiplier})
	}
	response.Success(c, gin.H{
		"min_bet": cfg.MinBet, "max_bet": cfg.MaxBet,
		"daily_loss_limit": cfg.DailyLossLimit,
		"wheel":            gin.H{"segments": segments},
		"slots":            gin.H{"symbols": symbols},
		"scratch": gin.H{
			"win_rate":       cfg.Scratch.WinRate,
			"max_multiplier": cfg.Scratch.ClassicMaxMultiplier(),
			"lucky7": gin.H{
				"cells":       cfg.Scratch.Lucky7.Cells,
				"hit_rate":    cfg.Scratch.Lucky7.HitRate,
				"max_cells_mult": cfg.Scratch.Lucky7MaxMultiplier(),
			},
			"lines": gin.H{
				"win_rate": cfg.Scratch.Lines.WinRate,
				"symbols":  lineSymbols,
			},
		},
		"blackjack": cfg.Blackjack,
		"sicbo":     cfg.Sicbo,
		"baccarat":  cfg.Baccarat,
	})
}

// HandleMe 当前用户与实时余额。
func (s *Server) HandleMe(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	ctx := c.Request.Context()
	u, err := s.st.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Unauthorized(c, err.Error())
		} else {
			response.InternalError(c, "查询用户失败")
		}
		return
	}
	stats, err := s.st.TodayStats(ctx, uid)
	if err != nil {
		response.InternalError(c, "查询今日统计失败")
		return
	}
	role := u.Role
	if role == "" {
		role = currentRole(c)
	}
	data := gin.H{
		"user":         gin.H{"id": u.ID, "email": u.Email, "role": role, "balance": u.Balance},
		"is_admin":     role == "admin",
		"today_profit": stats.Profit,
		"today_rounds": stats.Rounds,
	}
	if game, _ := s.st.GetActiveBlackjack(ctx, uid); game != nil {
		data["active_blackjack"] = game.ID
	}
	response.Success(c, data)
}

// HandleHistory 用户流水。
func (s *Server) HandleHistory(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	rows, err := s.st.ListHistory(c.Request.Context(), uid, limit, offset)
	if err != nil {
		response.InternalError(c, "查询失败")
		return
	}
	if rows == nil {
		rows = []BetRow{}
	}
	response.Success(c, gin.H{"items": rows})
}

// HandleLeaderboard 最近 7 天盈利榜（邮箱打码，展示在游戏化大厅）。
func (s *Server) HandleLeaderboard(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	items, err := s.st.Leaderboard(c.Request.Context())
	if err != nil {
		response.InternalError(c, "查询失败")
		return
	}
	if items == nil {
		items = []LeaderboardEntry{}
	}
	response.Success(c, gin.H{"items": items})
}

// spinRequest 通用下注请求。
type spinRequest struct {
	Bet float64 `json:"bet"`
}

// checkBet 校验下注范围，失败时已写响应并返回 false。
func (s *Server) checkBet(c *gin.Context, bet float64) (float64, bool) {
	if bet <= 0 {
		response.BadRequest(c, "下注金额必须大于 0")
		return 0, false
	}
	bet = roundAmount(bet)
	cfg := s.mgr.Get()
	if bet < cfg.MinBet {
		response.BadRequest(c, "低于最小下注")
		return 0, false
	}
	if bet > cfg.MaxBet {
		response.BadRequest(c, "超过最大下注")
		return 0, false
	}
	return bet, true
}

func roundAmount(v float64) float64 { return float64(int64(v*1e8+0.5)) / 1e8 }

// HandleWheelSpin 大转盘：权重抽扇区 → 原子结算。
func (s *Server) HandleWheelSpin(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	var req spinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	bet, ok := s.checkBet(c, req.Bet)
	if !ok {
		return
	}
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	cfg := s.mgr.Get()

	idx := cfg.Wheel.Spin(rng())
	seg := cfg.Wheel.Segments[idx]
	payout := roundAmount(bet * seg.Multiplier)

	detail, _ := json.Marshal(map[string]any{"index": idx, "multiplier": seg.Multiplier, "label": seg.Label})
	balance, err := s.st.SettleInstantBet(c.Request.Context(), uid, bet, payout, "wheel", string(detail), cfg.DailyLossLimit)
	if err != nil {
		gameError(c, err)
		return
	}
	s.invalidateBalance(c.Request.Context(), uid)
	response.Success(c, gin.H{
		"segment": idx, "multiplier": seg.Multiplier,
		"payout": payout, "bet": bet, "balance": balance,
	})
}

// HandleSlotsSpin 老虎机：三卷轴抽签 → 原子结算。
func (s *Server) HandleSlotsSpin(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	var req spinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	bet, ok := s.checkBet(c, req.Bet)
	if !ok {
		return
	}
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	cfg := s.mgr.Get()

	reels, mult := cfg.Slots.Spin(rng())
	payout := roundAmount(bet * mult)

	ids := []string{reels[0].ID, reels[1].ID, reels[2].ID}
	emojis := []string{reels[0].Emoji, reels[1].Emoji, reels[2].Emoji}
	detail, _ := json.Marshal(map[string]any{"reels": ids, "multiplier": mult})

	balance, err := s.st.SettleInstantBet(c.Request.Context(), uid, bet, payout, "slots", string(detail), cfg.DailyLossLimit)
	if err != nil {
		gameError(c, err)
		return
	}
	s.invalidateBalance(c.Request.Context(), uid)
	response.Success(c, gin.H{
		"reels": emojis, "reel_ids": ids,
		"multiplier": mult, "payout": payout, "bet": bet, "balance": balance,
	})
}
