package games

import (
	"math/rand/v2"
	"testing"
)

func TestWheelRTP(t *testing.T) {
	rtp := DefaultWheel().RTP()
	if rtp < 0.93 || rtp > 0.98 {
		t.Fatalf("转盘理论回报率超出合理区间: %.4f", rtp)
	}
}

func TestWheelSpinDistribution(t *testing.T) {
	w := DefaultWheel()
	rng := rand.New(rand.NewPCG(1, 2))
	n := 200000
	counts := make([]int, len(w.Segments))
	for i := 0; i < n; i++ {
		counts[w.Spin(rng)]++
	}
	for i, s := range w.Segments {
		want := s.Weight / 92.0
		got := float64(counts[i]) / float64(n)
		if diff := got - want; diff > 0.01 || diff < -0.01 {
			t.Fatalf("扇区 %d (%sx) 频率偏离: want %.4f got %.4f", i, s.Label, want, got)
		}
	}
}

func TestSlotsRTP(t *testing.T) {
	rtp := DefaultSlots().RTP()
	t.Logf("老虎机理论回报率: %.4f", rtp)
	if rtp < 0.90 || rtp > 0.98 {
		t.Fatalf("老虎机理论回报率超出合理区间: %.4f", rtp)
	}
}

func TestSlotsSpinSanity(t *testing.T) {
	s := DefaultSlots()
	rng := rand.New(rand.NewPCG(3, 4))
	var hit, triple int
	const n = 100000
	for i := 0; i < n; i++ {
		reels, mult := s.Spin(rng)
		if mult < 0 {
			t.Fatalf("倍数不能为负: %v", mult)
		}
		if mult > 0 {
			hit++
			if reels[0].ID == reels[1].ID && reels[1].ID == reels[2].ID {
				triple++
			}
		}
	}
	if hit == 0 || triple == 0 {
		t.Fatalf("抽样异常: 中奖 %d 次, 三连 %d 次", hit, triple)
	}
}

func TestHandValue(t *testing.T) {
	cases := []struct {
		cards []Card
		total int
		soft  bool
	}{
		{[]Card{{"A", "s"}, {"K", "h"}}, 21, true},
		{[]Card{{"A", "s"}, {"A", "h"}, {"A", "d"}}, 13, true},
		{[]Card{{"A", "s"}, {"9", "h"}, {"A", "d"}}, 21, true},
		{[]Card{{"A", "s"}, {"9", "h"}, {"5", "d"}}, 15, false},
		{[]Card{{"K", "s"}, {"Q", "h"}, {"2", "d"}}, 22, false},
		{[]Card{{"10", "s"}, {"10", "h"}, {"A", "d"}, {"A", "c"}}, 22, false},
		{[]Card{{"7", "s"}}, 7, false},
	}
	for i, c := range cases {
		total, soft := HandValue(c.cards)
		if total != c.total || soft != c.soft {
			t.Errorf("case %d: want (%d,%v) got (%d,%v)", i, c.total, c.soft, total, soft)
		}
	}
}

func TestIsBlackjack(t *testing.T) {
	if !IsBlackjack([]Card{{"A", "s"}, {"J", "h"}}) {
		t.Error("A+J 应为 blackjack")
	}
	if IsBlackjack([]Card{{"10", "s"}, {"Q", "h"}, {"A", "d"}}) {
		t.Error("三张 21 不算自然 blackjack")
	}
}

func TestDealerRule(t *testing.T) {
	cfg := DefaultBlackjack() // S17
	if !cfg.DealerShouldHit(16, false) {
		t.Error("硬16必须牌")
	}
	if cfg.DealerShouldHit(17, false) {
		t.Error("硬17必须停")
	}
	if cfg.DealerShouldHit(17, true) {
		t.Error("S17 规则下软17应停牌")
	}
	hit := &BlackjackConfig{DealerStandsSoft17: false}
	if !hit.DealerShouldHit(17, true) {
		t.Error("H17 规则下软17应要牌")
	}
}

// TestBlackjackSimulation 用"硬17停牌"简单策略模拟庄家对局，验证期望回报在合理区间。
func TestBlackjackSimulation(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	cfg := DefaultBlackjack()
	const n = 50000
	var staked, returned float64
	for i := 0; i < n; i++ {
		deck := NewDeck(rng)
		staked += 1
		draw := func() Card {
			c, _ := Draw(&deck)
			return c
		}
		player := []Card{draw(), draw()}
		dealer := []Card{draw(), draw()}
		// 玩家策略：仅当手牌为一张A+8/9(软19/20)时视为停牌，否则 <17 要牌（简化策略）
		for {
			total, _ := HandValue(player)
			if total >= 17 {
				break
			}
			player = append(player, draw())
		}
		pt, _ := HandValue(player)
		playerBust := pt > 21
		if !playerBust && !IsBlackjack(player) && !IsBlackjack(dealer) {
			for {
				dt, soft := HandValue(dealer)
				if !cfg.DealerShouldHit(dt, soft) {
					break
				}
				dealer = append(dealer, draw())
			}
		}
		switch {
		case IsBlackjack(player) && !IsBlackjack(dealer):
			returned += cfg.BlackjackPays
		case IsBlackjack(dealer) && !IsBlackjack(player):
			// 输
		case playerBust:
			// 输
		default:
			dt, _ := HandValue(dealer)
			switch {
			case dt > 21 || pt > dt:
				returned += 2
			case pt == dt:
				returned += 1
			}
		}
	}
	rtp := returned / staked
	t.Logf("21点模拟回报率(硬17停牌策略): %.4f", rtp)
	if rtp < 0.85 || rtp > 1.05 {
		t.Fatalf("21点模拟回报率异常: %.4f", rtp)
	}
}

func TestDeckIntegrity(t *testing.T) {
	rng := rand.New(rand.NewPCG(9, 9))
	deck := NewDeck(rng)
	if len(deck) != 52 {
		t.Fatalf("牌堆应有52张, got %d", len(deck))
	}
	seen := map[Card]bool{}
	for _, c := range deck {
		if seen[c] {
			t.Fatalf("重复发牌: %v", c)
		}
		seen[c] = true
	}
	// 摸完 52 张后应报错
	for i := 0; i < 52; i++ {
		if _, err := Draw(&deck); err != nil {
			t.Fatalf("第%d张摸牌不应出错: %v", i+1, err)
		}
	}
	if _, err := Draw(&deck); err == nil {
		t.Error("空牌堆摸牌应报错")
	}
}
