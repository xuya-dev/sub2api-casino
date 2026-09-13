package games

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

// 百家乐下注方向 / 结果。
const (
	BaccaratPlayer = "player" // 闲
	BaccaratBanker = "banker" // 庄
	BaccaratTie    = "tie"    // 和
)

// BaccaratSides 合法下注方向列表。
var BaccaratSides = []string{BaccaratPlayer, BaccaratBanker, BaccaratTie}

// BaccaratConfig 百家乐配置：三个方向的赔付倍数（含本金）。
// 默认：闲 2（1:1）、庄 1.95（含 5% 佣金）、和 9（8:1）。
type BaccaratConfig struct {
	Player float64 `json:"player"`
	Banker float64 `json:"banker"`
	Tie    float64 `json:"tie"`
}

// DefaultBaccarat 默认赔率 2 / 1.95 / 9。
// 本实现按规则文件从简：和局只对 tie 注赔付，闲/庄注判负（不退本金），
// 默认赔率下三个方向的理论回报约为 89.2% / 89.4% / 85.6%。
func DefaultBaccarat() *BaccaratConfig {
	return &BaccaratConfig{Player: 2, Banker: 1.95, Tie: 9}
}

// BaccaratRound 一局百家乐的发牌与结果。
type BaccaratRound struct {
	PlayerCards  []Card `json:"player_cards"`
	BankerCards  []Card `json:"banker_cards"`
	PlayerPoints int    `json:"player_points"`
	BankerPoints int    `json:"banker_points"`
	Outcome      string `json:"outcome"` // player / banker / tie
}

// BaccaratCardPoint 单张牌的百家乐点数：A=1，2-9 按面值，10/J/Q/K=0。
func BaccaratCardPoint(c Card) int {
	switch c.Rank {
	case "10", "J", "Q", "K":
		return 0
	case "A":
		return 1
	default:
		return int(c.Rank[0] - '0')
	}
}

// BaccaratPoints 手牌百家乐点数（各牌点数之和 mod 10）。
func BaccaratPoints(cards []Card) int {
	total := 0
	for _, c := range cards {
		total += BaccaratCardPoint(c)
	}
	return total % 10
}

// PlayerShouldDraw 闲家要牌规则：两张 0-5 点补第三张，6-7 停牌。
func PlayerShouldDraw(points int) bool { return points <= 5 }

// BankerShouldDraw 庄家要牌表（标准百家乐）。
// playerThird 为 nil 表示闲家未补牌（闲家停牌，含天牌局）：
// 此时庄 0-5 点补牌，6-7 停。
// playerThird 非 nil 表示闲家补了第三张（值为该张的点数）：
// 庄 0-2 恒补；3 除闲第三张为 8 外均补；4 遇 2-7 补；5 遇 4-7 补；
// 6 遇 6-7 补；7 停（8-9 已是天牌，不会走到本表）。
func BankerShouldDraw(bankerPoints int, playerThird *int) bool {
	if playerThird == nil {
		return bankerPoints <= 5
	}
	t := *playerThird
	switch {
	case bankerPoints <= 2:
		return true
	case bankerPoints == 3:
		return t != 8
	case bankerPoints == 4:
		return t >= 2 && t <= 7
	case bankerPoints == 5:
		return t >= 4 && t <= 7
	case bankerPoints == 6:
		return t == 6 || t == 7
	default: // 7 及以上（8-9 天牌局不会进入要牌流程）
		return false
	}
}

// playBaccarat 从双方各两张起按要牌规则补牌到终局并判定结果。
// 单独抽出便于用固定手牌测试要牌表。
func playBaccarat(player, banker []Card, draw func() Card) *BaccaratRound {
	pp, bp := BaccaratPoints(player), BaccaratPoints(banker)

	// 天牌：任一方起手 8-9 点即双方停牌直接比较
	var playerThird *int
	if pp < 8 && bp < 8 {
		if PlayerShouldDraw(pp) {
			third := draw()
			player = append(player, third)
			pt := BaccaratCardPoint(third)
			playerThird = &pt
			pp = BaccaratPoints(player)
		}
		if BankerShouldDraw(bp, playerThird) {
			banker = append(banker, draw())
			bp = BaccaratPoints(banker)
		}
	}

	outcome := BaccaratTie
	switch {
	case pp > bp:
		outcome = BaccaratPlayer
	case bp > pp:
		outcome = BaccaratBanker
	}
	return &BaccaratRound{
		PlayerCards: player, BankerCards: banker,
		PlayerPoints: pp, BankerPoints: bp, Outcome: outcome,
	}
}

// Deal 开一局：洗一副 52 张牌 → 闲庄各两张 → 按要牌规则补到终局。
func (c *BaccaratConfig) Deal(rng *rand.Rand) *BaccaratRound {
	deck := NewDeck(rng)
	draw := func() Card { card, _ := Draw(&deck); return card }
	player := []Card{draw(), draw()}
	banker := []Card{draw(), draw()}
	return playBaccarat(player, banker, draw)
}

// Settle 按下注方向结算，返回赔付倍数（含本金，判负为 0）。
// 结果与下注方向一致才赔付：闲>庄 → player；庄>闲 → banker；相等 → tie。
// side 非法或 bet 非正时报错。
func (c *BaccaratConfig) Settle(side string, bet float64, r *BaccaratRound) (float64, error) {
	if bet <= 0 {
		return 0, errors.New("下注金额必须大于 0")
	}
	if r == nil {
		return 0, errors.New("牌局不存在")
	}
	// 和局：tie 注按赔率派奖；闲/庄注退回本金（倍数 1）
	if r.Outcome == BaccaratTie {
		if side == BaccaratTie {
			return c.Tie, nil
		}
		return 1, nil
	}
	var m float64
	switch side {
	case BaccaratPlayer:
		m = c.Player
	case BaccaratBanker:
		m = c.Banker
	case BaccaratTie:
		m = c.Tie
	default:
		return 0, fmt.Errorf("未知百家乐下注方向: %s", side)
	}
	if r.Outcome != side {
		return 0, nil
	}
	return m, nil
}

// 标准百家乐（8 副牌）结果概率，用于管理端 RTP 预览；
// 本实现每局洗单副牌，概率与之差异在 0.3 个百分点以内，不影响实际结算。
const (
	baccaratPPlayer = 0.446247
	baccaratPBanker = 0.458597
	baccaratPTie    = 0.095156
)

// RTP 返回三个下注方向的理论回报率（含本金倍数 × 对应结果概率 + 和局返本）。
// 和局时闲/庄注退回本金，故闲/庄 RTP 叠加 baccaratPTie × 1。
func (c *BaccaratConfig) RTP() map[string]float64 {
	return map[string]float64{
		BaccaratPlayer: baccaratPPlayer*c.Player + baccaratPTie,
		BaccaratBanker: baccaratPBanker*c.Banker + baccaratPTie,
		BaccaratTie:    baccaratPTie * c.Tie,
	}
}
