package casino

import (
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/casino/games"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// baccaratRequest 百家乐下注请求。
type baccaratRequest struct {
	Bet  float64 `json:"bet"`
	Side string  `json:"side"`
}

// validBaccaratSide 校验下注方向合法性。
func validBaccaratSide(side string) bool {
	switch side {
	case games.BaccaratPlayer, games.BaccaratBanker, games.BaccaratTie:
		return true
	}
	return false
}

// HandleBaccaratDeal 百家乐：单副牌洗牌发牌到终局 → 原子结算（倍数含本金：
// 闲 2、庄 1.95 含 5% 佣金、和 9；和局只对 tie 注赔付）。
func (s *Server) HandleBaccaratDeal(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	var req baccaratRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	if !validBaccaratSide(req.Side) {
		response.BadRequest(c, "下注方向必须是 player/banker/tie")
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

	round := cfg.Baccarat.Deal(rng())
	multiplier, err := cfg.Baccarat.Settle(req.Side, bet, round)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	payout := roundAmount(bet * multiplier)

	detail, _ := json.Marshal(map[string]any{
		"side": req.Side, "outcome": round.Outcome, "multiplier": multiplier,
		"player_cards": round.PlayerCards, "banker_cards": round.BankerCards,
		"player_points": round.PlayerPoints, "banker_points": round.BankerPoints,
	})
	balance, err := s.st.SettleInstantBet(c.Request.Context(), uid, bet, payout, "baccarat", string(detail), cfg.DailyLossLimit)
	if err != nil {
		gameError(c, err)
		return
	}
	s.invalidateBalance(c.Request.Context(), uid)
	response.Success(c, gin.H{
		"player_cards": round.PlayerCards, "banker_cards": round.BankerCards,
		"player_points": round.PlayerPoints, "banker_points": round.BankerPoints,
		"outcome":    round.Outcome,
		"multiplier": multiplier, "payout": payout, "bet": bet, "balance": balance,
	})
}
