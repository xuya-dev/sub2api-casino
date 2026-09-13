package casino

import (
	"encoding/json"

	"github.com/Wei-Shaw/sub2api/internal/casino/games"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// scratchRequest 刮刮乐请求：玩法 + 卡片面值 + 一次购买的张数。
type scratchRequest struct {
	Mode      string `json:"mode"`
	FaceValue float64 `json:"face_value"`
	Count     int    `json:"count"`
}

// scratchCountMin/Max 单次购买张数上下限。
const (
	scratchCountMin = 1
	scratchCountMax = 10
)

// HandleScratchReveal 刮刮乐（面值+数量模型）：选玩法、卡片面值与购买数量，
// 一次性购买 count 张卡并按 total = face_value × count 原子结算（game 记 "scratch"）。
// 三种玩法逐张独立生成：
//   - classic：单格涂层按 WinRate 判定中奖，中奖按档位抽倍数；
//   - lucky7 ：7 格涂层每格独立命中幸运7，命中格倍数累加；
//   - lines  ：3×3 宫格按 WinRate 出一条三连线，派该符号倍数。
//
// 响应含逐卡涂层格内容（cells）、中奖线、中奖张数与总派奖。
func (s *Server) HandleScratchReveal(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	var req scratchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	if req.Count < scratchCountMin || req.Count > scratchCountMax {
		response.BadRequest(c, "购买数量须在 1~10 张之间")
		return
	}
	mode := games.ScratchMode(req.Mode)
	if mode == "" {
		mode = games.ScratchModeClassic
	}
	valid := false
	for _, m := range games.AllScratchModes {
		if mode == m {
			valid = true
			break
		}
	}
	if !valid {
		response.BadRequest(c, "未知玩法")
		return
	}
	face, ok := s.checkBet(c, req.FaceValue)
	if !ok {
		return
	}
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	cfg := s.mgr.Get()

	cards := make([]games.ScratchCard, req.Count)
	winCount, totalPayout := 0, 0.0
	for i := range cards {
		var card games.ScratchCard
		switch mode {
		case games.ScratchModeLucky7:
			card = cfg.Scratch.Lucky7.Play(rng(), face)
		case games.ScratchModeLines:
			card = cfg.Scratch.Lines.Play(rng(), face)
		default:
			card = cfg.Scratch.Play(rng(), face)
		}
		card.Prize = roundAmount(card.Prize)
		if card.Multiplier > 0 {
			winCount++
		}
		totalPayout += card.Prize
		cards[i] = card
	}
	totalBet := roundAmount(face * float64(req.Count))
	totalPayout = roundAmount(totalPayout)

	detail, _ := json.Marshal(map[string]any{
		"mode": mode, "face_value": face, "count": req.Count, "cards": cards,
		"win_count": winCount, "total_payout": totalPayout, "total_bet": totalBet,
	})
	balance, err := s.st.SettleInstantBet(c.Request.Context(), uid, totalBet, totalPayout, "scratch", string(detail), cfg.DailyLossLimit)
	if err != nil {
		gameError(c, err)
		return
	}
	s.invalidateBalance(c.Request.Context(), uid)
	response.Success(c, gin.H{
		"mode": mode, "face_value": face, "count": req.Count, "cards": cards,
		"win_count": winCount, "total_payout": totalPayout,
		"total_bet": totalBet, "balance": balance,
	})
}
