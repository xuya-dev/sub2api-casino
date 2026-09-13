package casino

import (
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/casino/games"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// sicboRequest 骰宝下注请求。
type sicboRequest struct {
	Bet     float64 `json:"bet"`
	BetType string  `json:"bet_type"`
}

// validSicboBetType 校验注项合法性。
func validSicboBetType(betType string) bool {
	switch betType {
	case games.SicboBig, games.SicboSmall, games.SicboOdd, games.SicboEven:
		return true
	}
	return false
}

// HandleSicboRoll 骰宝：掷三颗骰子 → 原子结算（围骰通杀，倍数含本金）。
func (s *Server) HandleSicboRoll(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	var req sicboRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	if !validSicboBetType(req.BetType) {
		response.BadRequest(c, "注项必须是 big/small/odd/even")
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

	dice := cfg.Sicbo.Roll(rng())
	multiplier, err := cfg.Sicbo.Settle(req.BetType, bet, dice)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	payout := roundAmount(bet * multiplier)
	sum := games.SicboSum(dice)
	result := games.SicboResult(dice)

	detail, _ := json.Marshal(map[string]any{
		"bet_type": req.BetType, "dice": dice, "sum": sum,
		"result": result, "multiplier": multiplier,
	})
	balance, err := s.st.SettleInstantBet(c.Request.Context(), uid, bet, payout, "sicbo", string(detail), cfg.DailyLossLimit)
	if err != nil {
		gameError(c, err)
		return
	}
	s.invalidateBalance(c.Request.Context(), uid)
	response.Success(c, gin.H{
		"dice": dice, "sum": sum, "result": result,
		"multiplier": multiplier, "payout": payout, "bet": bet, "balance": balance,
	})
}
