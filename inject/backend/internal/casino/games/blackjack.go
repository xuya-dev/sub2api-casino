package games

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
)

// Card 一张扑克牌。Rank: A,2..10,J,Q,K；Suit: s(黑桃) h(红心) d(方块) c(梅花)。
type Card struct {
	Rank string `json:"r"`
	Suit string `json:"s"`
}

var suits = []string{"s", "h", "d", "c"}
var ranks = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

// NewDeck 生成一副洗好的 52 张牌（牌堆头部在切片末尾，摸牌即 pop）。
func NewDeck(rng *rand.Rand) []Card {
	deck := make([]Card, 0, 52)
	for _, s := range suits {
		for _, r := range ranks {
			deck = append(deck, Card{Rank: r, Suit: s})
		}
	}
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// Draw 从牌堆摸一张牌。
func Draw(deck *[]Card) (Card, error) {
	if len(*deck) == 0 {
		return Card{}, fmt.Errorf("deck exhausted")
	}
	c := (*deck)[len(*deck)-1]
	*deck = (*deck)[:len(*deck)-1]
	return c, nil
}

// HandValue 计算手牌点数。soft 表示存在按 11 计算且未爆的 A。
func HandValue(cards []Card) (total int, soft bool) {
	total, aces := 0, 0
	for _, c := range cards {
		switch c.Rank {
		case "A":
			total += 11
			aces++
		case "K", "Q", "J", "10":
			total += 10
		default:
			// 2-9: Rank 字符即点数
			total += int(c.Rank[0] - '0')
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total, aces > 0
}

// IsBlackjack 前两张即 21 点（自然 blackjack）。
func IsBlackjack(cards []Card) bool {
	if len(cards) != 2 {
		return false
	}
	t, _ := HandValue(cards)
	return t == 21
}

// BlackjackConfig 21点规则配置。
type BlackjackConfig struct {
	BlackjackPays      float64 `json:"blackjack_pays"`       // 自然 blackjack 赔率（2.5 = 返还本金+1.5倍）
	DealerStandsSoft17 bool    `json:"dealer_stands_soft17"` // true: S17 庄家软17停牌
	DoubleAllowed      bool    `json:"double_allowed"`
}

// DefaultBlackjack 标准规则：BJ 赔 3:2，庄家软 17 停牌，允许加倍。
func DefaultBlackjack() *BlackjackConfig {
	return &BlackjackConfig{BlackjackPays: 2.5, DealerStandsSoft17: true, DoubleAllowed: true}
}

// DealerShouldHit 庄家是否要牌：硬点 <17 必须要；软 17 视规则。
func (c *BlackjackConfig) DealerShouldHit(total int, soft bool) bool {
	if total < 17 {
		return true
	}
	return total == 17 && soft && !c.DealerStandsSoft17
}

// DeckJSON / CardsJSON 序列化辅助（存库用）。
func DeckJSON(deck []Card) string { b, _ := json.Marshal(deck); return string(b) }
func CardsJSON(cards []Card) string {
	if cards == nil {
		cards = []Card{}
	}
	b, _ := json.Marshal(cards)
	return string(b)
}
func ParseCards(s string) []Card {
	var cs []Card
	_ = json.Unmarshal([]byte(s), &cs)
	return cs
}
