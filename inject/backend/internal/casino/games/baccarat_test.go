package games

import (
	"math/rand/v2"
	"testing"
)

func TestBaccaratCardPoint(t *testing.T) {
	cases := []struct {
		card Card
		want int
	}{
		{Card{"A", "s"}, 1},
		{Card{"2", "s"}, 2},
		{Card{"9", "h"}, 9},
		{Card{"10", "d"}, 0},
		{Card{"J", "c"}, 0},
		{Card{"Q", "s"}, 0},
		{Card{"K", "h"}, 0},
	}
	for _, tc := range cases {
		if got := BaccaratCardPoint(tc.card); got != tc.want {
			t.Errorf("BaccaratCardPoint(%v) = %d, want %d", tc.card, got, tc.want)
		}
	}
}

func TestBaccaratPoints(t *testing.T) {
	cases := []struct {
		cards []Card
		want  int
	}{
		{[]Card{{"9", "s"}, {"K", "h"}}, 9},             // 天牌9
		{[]Card{{"5", "s"}, {"4", "h"}}, 9},             // 天牌9
		{[]Card{{"9", "s"}, {"9", "h"}}, 8},             // 18 mod 10
		{[]Card{{"5", "s"}, {"5", "h"}, {"5", "d"}}, 5}, // 15 mod 10
		{[]Card{{"10", "s"}, {"A", "h"}}, 1},
		{[]Card{{"A", "s"}, {"A", "h"}}, 2},
		{[]Card{{"K", "s"}, {"Q", "h"}, {"J", "d"}}, 0},
		{[]Card{{"A", "s"}, {"8", "h"}}, 9},
	}
	for _, tc := range cases {
		if got := BaccaratPoints(tc.cards); got != tc.want {
			t.Errorf("BaccaratPoints(%v) = %d, want %d", tc.cards, got, tc.want)
		}
	}
}

func TestPlayerShouldDraw(t *testing.T) {
	cases := []struct {
		points int
		want   bool
	}{
		{0, true}, {1, true}, {5, true},
		{6, false}, {7, false},
		{8, false}, {9, false}, // 8-9 为天牌，理应不会进入要牌流程
	}
	for _, tc := range cases {
		if got := PlayerShouldDraw(tc.points); got != tc.want {
			t.Errorf("PlayerShouldDraw(%d) = %v, want %v", tc.points, got, tc.want)
		}
	}
}

func TestBankerShouldDraw(t *testing.T) {
	pt := func(v int) *int { return &v }
	cases := []struct {
		name        string
		banker      int
		playerThird *int // nil = 闲家未补牌
		want        bool
	}{
		// 闲家停牌：庄 0-5 补，6-7 停
		{"停牌-庄0", 0, nil, true},
		{"停牌-庄5", 5, nil, true},
		{"停牌-庄6", 6, nil, false},
		{"停牌-庄7", 7, nil, false},
		// 闲家补牌后的查表
		{"庄0-恒补", 0, pt(8), true},
		{"庄2-恒补", 2, pt(0), true},
		{"庄3-遇8停", 3, pt(8), false},
		{"庄3-遇9补", 3, pt(9), true},
		{"庄3-遇1补", 3, pt(1), true},
		{"庄4-遇2补", 4, pt(2), true},
		{"庄4-遇7补", 4, pt(7), true},
		{"庄4-遇1停", 4, pt(1), false},
		{"庄4-遇8停", 4, pt(8), false},
		{"庄5-遇4补", 5, pt(4), true},
		{"庄5-遇7补", 5, pt(7), true},
		{"庄5-遇3停", 5, pt(3), false},
		{"庄5-遇8停", 5, pt(8), false},
		{"庄6-遇6补", 6, pt(6), true},
		{"庄6-遇7补", 6, pt(7), true},
		{"庄6-遇5停", 6, pt(5), false},
		{"庄6-遇8停", 6, pt(8), false},
		{"庄7-恒停", 7, pt(0), false},
		{"庄7-遇6停", 7, pt(6), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BankerShouldDraw(tc.banker, tc.playerThird); got != tc.want {
				t.Fatalf("BankerShouldDraw(%d, %v) = %v, want %v", tc.banker, tc.playerThird, got, tc.want)
			}
		})
	}
}

// drawFrom 按给定牌序依次发牌（测试 playBaccarat 用）。
func drawFrom(cards ...Card) func() Card {
	i := 0
	return func() Card {
		c := cards[i]
		i++
		return c
	}
}

func TestPlayBaccarat(t *testing.T) {
	cases := []struct {
		name        string
		player      []Card
		banker      []Card
		third       []Card // 按需补的牌序（闲先庄后）
		wantPlayerP int
		wantBankerP int
		wantOutcome string
		wantPCards  int
		wantBCards  int
	}{
		{
			// 双天牌：闲 9 庄 8，均不补牌
			name: "天牌9对8", player: []Card{{"9", "s"}, {"K", "h"}},
			banker:      []Card{{"5", "s"}, {"3", "h"}},
			wantPlayerP: 9, wantBankerP: 8, wantOutcome: BaccaratPlayer,
			wantPCards: 2, wantBCards: 2,
		},
		{
			// 庄天牌 9（9+K），闲 5 也不再补牌
			name: "庄天牌停牌", player: []Card{{"5", "s"}, {"K", "h"}},
			banker:      []Card{{"9", "s"}, {"K", "h"}},
			wantPlayerP: 5, wantBankerP: 9, wantOutcome: BaccaratBanker,
			wantPCards: 2, wantBCards: 2,
		},
		{
			// 闲 5 补 2 → 7；庄 6 遇闲第三张点 2 停牌 → 闲胜
			name: "闲5补2庄6停", player: []Card{{"5", "s"}, {"K", "h"}},
			banker:      []Card{{"6", "s"}, {"K", "h"}},
			third:       []Card{{"2", "d"}},
			wantPlayerP: 7, wantBankerP: 6, wantOutcome: BaccaratPlayer,
			wantPCards: 3, wantBCards: 2,
		},
		{
			// 闲 5 补 9(点9) → 4；庄 5 遇闲第三张点 9 不补 → 庄胜
			name: "庄5遇9停", player: []Card{{"5", "s"}, {"K", "h"}},
			banker:      []Card{{"5", "s"}, {"K", "h"}},
			third:       []Card{{"9", "d"}},
			wantPlayerP: 4, wantBankerP: 5, wantOutcome: BaccaratBanker,
			wantPCards: 3, wantBCards: 2,
		},
		{
			// 闲 4 补 3 → 7；庄 0-5 必补：庄 5 遇闲第三张点 3 不补…
			// 但庄 1 遇任意点均补 → 庄 1 补 4 → 5；闲 7 胜
			name: "庄1必补", player: []Card{{"4", "s"}, {"K", "h"}},
			banker:      []Card{{"1", "s"}, {"K", "h"}},
			third:       []Card{{"3", "d"}, {"4", "c"}},
			wantPlayerP: 7, wantBankerP: 5, wantOutcome: BaccaratPlayer,
			wantPCards: 3, wantBCards: 3,
		},
		{
			// 闲 6 停牌；庄 3 且闲未补牌 → 庄补：补 5 → 8 庄胜
			name: "闲停庄5内补", player: []Card{{"6", "s"}, {"K", "h"}},
			banker:      []Card{{"3", "s"}, {"K", "h"}},
			third:       []Card{{"5", "c"}},
			wantPlayerP: 6, wantBankerP: 8, wantOutcome: BaccaratBanker,
			wantPCards: 2, wantBCards: 3,
		},
		{
			// 闲 4 补 5 → 9；庄 5 遇闲第三张点 5 补 4 → 9 → 和
			name: "和局", player: []Card{{"4", "s"}, {"K", "h"}},
			banker:      []Card{{"5", "s"}, {"K", "h"}},
			third:       []Card{{"5", "d"}, {"4", "c"}},
			wantPlayerP: 9, wantBankerP: 9, wantOutcome: BaccaratTie,
			wantPCards: 3, wantBCards: 3,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			draw := drawFrom(tc.third...)
			r := playBaccarat(tc.player, tc.banker, draw)
			if r.PlayerPoints != tc.wantPlayerP || r.BankerPoints != tc.wantBankerP {
				t.Fatalf("点数 want (%d,%d), got (%d,%d)", tc.wantPlayerP, tc.wantBankerP, r.PlayerPoints, r.BankerPoints)
			}
			if r.Outcome != tc.wantOutcome {
				t.Fatalf("结果 want %q, got %q", tc.wantOutcome, r.Outcome)
			}
			if len(r.PlayerCards) != tc.wantPCards || len(r.BankerCards) != tc.wantBCards {
				t.Fatalf("牌数 want (%d,%d), got (%d,%d)", tc.wantPCards, tc.wantBCards, len(r.PlayerCards), len(r.BankerCards))
			}
		})
	}
}

func TestBaccaratSettle(t *testing.T) {
	cfg := DefaultBaccarat()
	cases := []struct {
		name    string
		side    string
		outcome string
		bet     float64
		want    float64
		wantErr bool
	}{
		{"买闲闲胜", BaccaratPlayer, BaccaratPlayer, 1, 2, false},
		{"买庄庄胜", BaccaratBanker, BaccaratBanker, 1, 1.95, false},
		{"买和和局", BaccaratTie, BaccaratTie, 1, 9, false},
		{"买闲庄胜判负", BaccaratPlayer, BaccaratBanker, 1, 0, false},
		{"买闲和局退本", BaccaratPlayer, BaccaratTie, 1, 1, false},
		{"买庄和局退本", BaccaratBanker, BaccaratTie, 1, 1, false},
		{"买和闲胜判负", BaccaratTie, BaccaratPlayer, 1, 0, false},
		{"非法方向", "dragon", BaccaratPlayer, 1, 0, true},
		{"bet为0", BaccaratPlayer, BaccaratPlayer, 0, 0, true},
		{"bet为负", BaccaratPlayer, BaccaratPlayer, -1, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cfg.Settle(tc.side, tc.bet, &BaccaratRound{Outcome: tc.outcome})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("应报错, got 倍数 %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("不应报错: %v", err)
			}
			if got != tc.want {
				t.Fatalf("倍数 want %v, got %v", tc.want, got)
			}
		})
	}
	if _, err := cfg.Settle(BaccaratPlayer, 1, nil); err == nil {
		t.Fatal("空牌局应报错")
	}
	// 自定义赔率
	custom := &BaccaratConfig{Player: 2.1, Banker: 2, Tie: 10}
	if m, _ := custom.Settle(BaccaratPlayer, 1, &BaccaratRound{Outcome: BaccaratPlayer}); m != 2.1 {
		t.Fatalf("自定义闲赔率 = %v, want 2.1", m)
	}
}

func TestBaccaratDealSimulation(t *testing.T) {
	cfg := DefaultBaccarat()
	rng := rand.New(rand.NewPCG(21, 22))
	const n = 50000
	counts := map[string]int{BaccaratPlayer: 0, BaccaratBanker: 0, BaccaratTie: 0}
	var cardTotal int
	for i := 0; i < n; i++ {
		r := cfg.Deal(rng)
		if r.PlayerPoints < 0 || r.PlayerPoints > 9 || r.BankerPoints < 0 || r.BankerPoints > 9 {
			t.Fatalf("点数越界: %v", r)
		}
		// 每手至多 3 张；天牌局（起手任一方 8-9）必为各 2 张；
		// 非天牌局闲家 0-5 必补、6-7 必停
		if len(r.PlayerCards) > 3 || len(r.BankerCards) > 3 {
			t.Fatalf("手牌数量非法: %v", r)
		}
		pInit, bInit := BaccaratPoints(r.PlayerCards[:2]), BaccaratPoints(r.BankerCards[:2])
		if (pInit >= 8 || bInit >= 8) && (len(r.PlayerCards) != 2 || len(r.BankerCards) != 2) {
			t.Fatalf("天牌局不应补牌: %v", r)
		}
		if pInit < 8 && bInit < 8 {
			if PlayerShouldDraw(pInit) && len(r.PlayerCards) != 3 {
				t.Fatalf("闲家应补牌: %v", r)
			}
			if !PlayerShouldDraw(pInit) && len(r.PlayerCards) != 2 {
				t.Fatalf("闲家不应补牌: %v", r)
			}
		}
		// 结果与点数一致
		want := BaccaratTie
		if r.PlayerPoints > r.BankerPoints {
			want = BaccaratPlayer
		} else if r.BankerPoints > r.PlayerPoints {
			want = BaccaratBanker
		}
		if r.Outcome != want {
			t.Fatalf("结果与点数不符: %+v", r)
		}
		counts[r.Outcome]++
		cardTotal += len(r.PlayerCards) + len(r.BankerCards)
	}
	// 单副牌概率与标准 8 副牌接近，容差 3 个百分点
	tolerances := map[string]float64{BaccaratPlayer: baccaratPPlayer, BaccaratBanker: baccaratPBanker, BaccaratTie: baccaratPTie}
	for side, want := range tolerances {
		got := float64(counts[side]) / n
		if diff := got - want; diff > 0.03 || diff < -0.03 {
			t.Fatalf("%s 频率偏离: want %.4f, got %.4f", side, want, got)
		}
	}
	// 平均每局总牌数 4~5 张（部分局面补 1-2 张）
	if avg := float64(cardTotal) / n; avg < 4 || avg > 5 {
		t.Fatalf("平均牌数异常: %.3f", avg)
	}
}

func TestBaccaratRTP(t *testing.T) {
	rtps := DefaultBaccarat().RTP()
	cases := []struct {
		side string
		want float64
	}{
		{BaccaratPlayer, baccaratPPlayer*2 + baccaratPTie},
		{BaccaratBanker, baccaratPBanker*1.95 + baccaratPTie},
		{BaccaratTie, baccaratPTie * 9},
	}
	for _, tc := range cases {
		got := rtps[tc.side]
		if diff := got - tc.want; diff > 1e-9 || diff < -1e-9 {
			t.Fatalf("%s 理论回报 = %v, want %v", tc.side, got, tc.want)
		}
	}
}
