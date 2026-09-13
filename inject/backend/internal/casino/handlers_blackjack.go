package casino

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/casino/games"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// bjRequest 21点操作请求（deal 只要 bet，其余要 game_id）。
type bjRequest struct {
	Bet    float64 `json:"bet"`
	GameID int64   `json:"game_id"`
}

// bjView 发给前端的牌局视图：庄家暗牌在结算前隐藏。
type bjView struct {
	GameID      int64        `json:"game_id"`
	Bet         float64      `json:"bet"`
	Doubled     bool         `json:"doubled"`
	PlayerCards []games.Card `json:"player_cards"`
	DealerCards []games.Card `json:"dealer_cards"` // active 时仅第一张（明牌）
	PlayerTotal int          `json:"player_total"`
	PlayerSoft  bool         `json:"player_soft"`
	Status      string       `json:"status"` // active / settled
	Result      string       `json:"result"`
	Payout      float64      `json:"payout"`
	CanDouble   bool         `json:"can_double"`
	Balance     float64      `json:"balance"`
}

func viewOf(g *BlackjackGame, settled, doubleAllowed bool) bjView {
	player := games.ParseCards(g.PlayerCards)
	dealer := games.ParseCards(g.DealerCards)
	total, soft := games.HandValue(player)
	v := bjView{
		GameID: g.ID, Bet: g.Bet, Doubled: g.Doubled,
		PlayerCards: player, PlayerTotal: total, PlayerSoft: soft,
		Status: g.Status, Result: g.Result, Payout: g.Payout,
	}
	if settled {
		v.DealerCards = dealer
	} else if len(dealer) > 0 {
		v.DealerCards = dealer[:1] // 只亮明牌
	}
	v.CanDouble = doubleAllowed && g.Status == "active" && !g.Doubled && len(player) == 2
	return v
}

// HandleBJCurrent 断线/换页恢复：返回当前进行中的牌局。
func (s *Server) HandleBJCurrent(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	g, err := s.st.GetActiveBlackjack(c.Request.Context(), uid)
	if err != nil {
		response.Success(c, gin.H{"game": nil})
		return
	}
	cfg := s.mgr.Get()
	response.Success(c, gin.H{"game": viewOf(g, false, cfg.Blackjack.DoubleAllowed)})
}

// HandleBJDeal 开局：扣注 → 洗牌发牌 → 玩家天牌直接结算。
func (s *Server) HandleBJDeal(c *gin.Context) {
	if !s.requireEnabled(c) {
		return
	}
	var req bjRequest
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
	ctx := c.Request.Context()
	cfg := s.mgr.Get()

	// 已有进行中的牌局则直接返回（前端恢复），避免误重复下注
	if existing, _ := s.st.GetActiveBlackjack(ctx, uid); existing != nil {
		response.Success(c, gin.H{"resumed": true, "game": viewOf(existing, false, cfg.Blackjack.DoubleAllowed)})
		return
	}

	deck := games.NewDeck(rng())
	draw := func() games.Card { c2, _ := games.Draw(&deck); return c2 }
	player := []games.Card{draw(), draw()}
	dealer := []games.Card{draw(), draw()}

	// 玩家自然 blackjack：peek 庄家，天牌局不建牌局行，即时原子结算
	if games.IsBlackjack(player) {
		payout, result := 0.0, "lose"
		if games.IsBlackjack(dealer) {
			payout, result = bet, "push"
		} else {
			payout, result = roundAmount(bet*cfg.Blackjack.BlackjackPays), "blackjack"
		}
		detail, _ := json.Marshal(map[string]any{"action": "deal", "result": result, "player": player, "dealer": dealer})
		balance, err := s.st.SettleInstantBet(ctx, uid, bet, payout, "blackjack", string(detail), cfg.DailyLossLimit)
		if err != nil {
			gameError(c, err)
			return
		}
		s.invalidateBalance(ctx, uid)
		g := &BlackjackGame{Bet: bet, PlayerCards: games.CardsJSON(player),
			DealerCards: games.CardsJSON(dealer), Status: "settled", Result: result, Payout: payout}
		v := viewOf(g, true, cfg.Blackjack.DoubleAllowed)
		v.Balance = balance
		response.Success(c, gin.H{"game": v})
		return
	}

	detail, _ := json.Marshal(map[string]any{"action": "deal"})
	gameID, balance, err := s.st.DealBlackjack(ctx, uid, bet,
		games.CardsJSON(player), games.CardsJSON(dealer), games.DeckJSON(deck), string(detail), cfg.DailyLossLimit)
	if err != nil {
		gameError(c, err)
		return
	}
	s.invalidateBalance(ctx, uid)
	pending := &BlackjackGame{ID: gameID, UserID: uid, Bet: bet,
		PlayerCards: games.CardsJSON(player), DealerCards: games.CardsJSON(dealer), Status: "active"}
	v := viewOf(pending, false, cfg.Blackjack.DoubleAllowed)
	v.Balance = balance
	response.Success(c, gin.H{"game": v})
}

// HandleBJHit 要牌：爆牌即输；到 21 自动停牌。
func (s *Server) HandleBJHit(c *gin.Context) { s.bjAction(c, actHit) }

// HandleBJStand 停牌：庄家按规则补牌后比较结算。
func (s *Server) HandleBJStand(c *gin.Context) { s.bjAction(c, actStand) }

// HandleBJDouble 加倍：再扣一倍注，只发一张牌后强制停牌。
func (s *Server) HandleBJDouble(c *gin.Context) { s.bjAction(c, actDouble) }

type bjAct int

const (
	actHit bjAct = iota
	actStand
	actDouble
)

// bjAction 统一状态机。加倍补扣与派奖在同一条结算事务里完成（原子）。
func (s *Server) bjAction(c *gin.Context, act bjAct) {
	if !s.requireEnabled(c) {
		return
	}
	var req bjRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求格式错误")
		return
	}
	uid, ok := currentUserID(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	cfg := s.mgr.Get()
	g, err := s.st.GetBlackjack(c.Request.Context(), req.GameID, uid)
	if err != nil {
		response.NotFound(c, "牌局不存在")
		return
	}
	if g.Status != "active" {
		response.Error(c, http.StatusConflict, "牌局已结束")
		return
	}
	if act == actDouble && (!cfg.Blackjack.DoubleAllowed || g.Doubled || len(games.ParseCards(g.PlayerCards)) != 2) {
		response.BadRequest(c, "当前不可加倍")
		return
	}

	player := games.ParseCards(g.PlayerCards)
	dealer := games.ParseCards(g.DealerCards)
	deck := games.ParseCards(g.Deck)
	draw := func() games.Card { c2, _ := games.Draw(&deck); return c2 }

	totalBet := g.Bet
	if act == actDouble {
		g.Doubled = true
		totalBet = g.Bet * 2
		player = append(player, draw())
	} else if act == actHit {
		player = append(player, draw())
	}

	pt, _ := games.HandValue(player)
	bust := pt > 21
	// 停牌条件：显式停牌、加倍发牌后、要牌到 21、爆牌
	goToDealer := act == actStand || act == actDouble || bust || pt == 21

	if !goToDealer {
		// 未结束：保存手牌与牌堆，返回进行中视图
		if err := s.st.UpdateBlackjackHands(c.Request.Context(), g.ID, games.CardsJSON(player), games.DeckJSON(deck)); err != nil {
			response.InternalError(c, "更新牌局失败")
			return
		}
		g.PlayerCards, g.Deck = games.CardsJSON(player), games.DeckJSON(deck)
		response.Success(c, gin.H{"game": viewOf(g, false, cfg.Blackjack.DoubleAllowed)})
		return
	}

	// 庄家按规则补牌（双方均非天牌时）
	if !games.IsBlackjack(player) && !games.IsBlackjack(dealer) {
		for {
			dt, soft := games.HandValue(dealer)
			if !cfg.Blackjack.DealerShouldHit(dt, soft) {
				break
			}
			dealer = append(dealer, draw())
		}
	}

	// 结算：totalBet 为含加倍的累计注
	payout, result := 0.0, "lose"
	dt, _ := games.HandValue(dealer)
	switch {
	case games.IsBlackjack(dealer) && games.IsBlackjack(player):
		payout, result = totalBet, "push"
	case games.IsBlackjack(player):
		payout, result = roundAmount(totalBet*cfg.Blackjack.BlackjackPays), "blackjack"
	case bust:
		result = "bust"
	case games.IsBlackjack(dealer):
		// 庄家天牌，玩家输
	case dt > 21 || pt > dt:
		payout, result = totalBet*2, "win"
	case pt == dt:
		payout, result = totalBet, "push"
	}
	payout = roundAmount(payout)

	action := map[bjAct]string{actHit: "hit", actStand: "stand", actDouble: "double"}[act]
	detail, _ := json.Marshal(map[string]any{
		"action": action, "result": result, "player": player, "dealer": dealer,
		"player_total": pt, "dealer_total": dt,
	})
	extraBet := 0.0
	if act == actDouble {
		extraBet = g.Bet
	}
	balance, err := s.st.SettleBlackjack(c.Request.Context(), g.ID, uid, extraBet, payout, g.Doubled, result,
		games.CardsJSON(player), games.CardsJSON(dealer), games.DeckJSON(deck), string(detail))
	if err != nil {
		gameError(c, err)
		return
	}
	s.invalidateBalance(c.Request.Context(), uid)
	g.Status, g.Result, g.Payout = "settled", result, payout
	g.PlayerCards, g.DealerCards = games.CardsJSON(player), games.CardsJSON(dealer)
	v := viewOf(g, true, cfg.Blackjack.DoubleAllowed)
	v.Balance = balance
	response.Success(c, gin.H{"game": v})
}

// SettleExpiredGames 周期任务：把超过 staleHours 未操作的牌局按"自动停牌"规则结算，
// 保证玩家断线不会永久锁死注金。由 RegisterCasinoRoutes 启动的后台 goroutine 周期调用。
func (s *Server) SettleExpiredGames(staleHours int) {
	ctx := context.Background()
	for {
		g, err := s.st.ExpiredBlackjack(ctx, staleHours)
		if err != nil || g == nil {
			return
		}
		cfg := s.mgr.Get()
		player := games.ParseCards(g.PlayerCards)
		dealer := games.ParseCards(g.DealerCards)
		deck := games.ParseCards(g.Deck)
		draw := func() games.Card { c2, _ := games.Draw(&deck); return c2 }
		for {
			dt, soft := games.HandValue(dealer)
			if !cfg.Blackjack.DealerShouldHit(dt, soft) {
				break
			}
			dealer = append(dealer, draw())
		}
		pt, _ := games.HandValue(player)
		dt, _ := games.HandValue(dealer)
		payout, result := 0.0, "lose"
		switch {
		case games.IsBlackjack(dealer) && games.IsBlackjack(player):
			payout, result = g.Bet, "push"
		case games.IsBlackjack(player):
			payout, result = roundAmount(g.Bet*cfg.Blackjack.BlackjackPays), "blackjack"
		case pt > 21:
			result = "bust"
		case dt > 21 || pt > dt:
			payout, result = g.Bet*2, "win"
		case pt == dt:
			payout, result = g.Bet, "push"
		}
		detail, _ := json.Marshal(map[string]any{"action": "timeout_stand", "result": result, "player": player, "dealer": dealer})
		if _, err := s.st.SettleBlackjack(ctx, g.ID, g.UserID, 0, roundAmount(payout), g.Doubled, result,
			games.CardsJSON(player), games.CardsJSON(dealer), games.DeckJSON(deck), string(detail)); err != nil {
			log.Printf("[casino] 超时牌局结算失败 game=%d: %v", g.ID, err)
			return
		}
		s.invalidateBalance(ctx, g.UserID)
		log.Printf("[casino] 超时牌局已自动停牌结算 game=%d user=%d result=%s", g.ID, g.UserID, result)
	}
}
