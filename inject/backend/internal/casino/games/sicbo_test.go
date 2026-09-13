package games

import (
	"math/rand/v2"
	"testing"
)

func TestSicboSettle(t *testing.T) {
	cfg := DefaultSicbo()
	cases := []struct {
		name    string
		betType string
		dice    [3]int
		want    float64
	}{
		{"大-下界11", SicboBig, [3]int{5, 4, 2}, 2},
		{"大-上界17", SicboBig, [3]int{6, 6, 5}, 2},
		{"大-10判负", SicboBig, [3]int{5, 4, 1}, 0},
		{"小-下界4", SicboSmall, [3]int{2, 1, 1}, 2},
		{"小-上界10", SicboSmall, [3]int{5, 4, 1}, 2},
		{"小-11判负", SicboSmall, [3]int{6, 4, 1}, 0},
		{"单-奇数和9", SicboOdd, [3]int{5, 3, 1}, 2},
		{"单-偶数和判负", SicboOdd, [3]int{5, 4, 1}, 0},
		{"双-偶数和10", SicboEven, [3]int{5, 4, 1}, 2},
		{"双-奇数和判负", SicboEven, [3]int{5, 3, 1}, 0},
		{"围骰通杀-大(和12本属大)", SicboBig, [3]int{4, 4, 4}, 0},
		{"围骰通杀-小(和6)", SicboSmall, [3]int{2, 2, 2}, 0},
		{"围骰通杀-单(和15)", SicboOdd, [3]int{5, 5, 5}, 0},
		{"围骰通杀-双(和18)", SicboEven, [3]int{6, 6, 6}, 0},
		{"围骰通杀-三个1", SicboSmall, [3]int{1, 1, 1}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cfg.Settle(tc.betType, 1, tc.dice)
			if err != nil {
				t.Fatalf("不应报错: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestSicboSettleCustomOddsAndErrors(t *testing.T) {
	cfg := &SicboConfig{Big: 2.5, Small: 1.5, Odd: 1, Even: 3}
	cases := []struct {
		name    string
		betType string
		bet     float64
		dice    [3]int
		want    float64
		wantErr bool
	}{
		{"自定义大赔率", SicboBig, 1, [3]int{6, 6, 5}, 2.5, false},
		{"自定义小赔率", SicboSmall, 1, [3]int{2, 1, 1}, 1.5, false},
		{"自定义单赔率下限1", SicboOdd, 1, [3]int{5, 3, 1}, 1, false},
		{"自定义双赔率", SicboEven, 1, [3]int{2, 2, 4}, 3, false},
		{"非法注项", "huge", 1, [3]int{1, 2, 3}, 0, true},
		{"空注项", "", 1, [3]int{1, 2, 3}, 0, true},
		{"bet为0", SicboBig, 0, [3]int{1, 2, 3}, 0, true},
		{"bet为负", SicboBig, -1, [3]int{1, 2, 3}, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cfg.Settle(tc.betType, tc.bet, tc.dice)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("应报错, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("不应报错: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestSicboResultAndSum(t *testing.T) {
	cases := []struct {
		dice [3]int
		want string
	}{
		{[3]int{4, 4, 4}, "triple"},
		{[3]int{1, 1, 1}, "triple"},
		{[3]int{6, 6, 5}, SicboBig},   // 17
		{[3]int{2, 1, 1}, SicboSmall}, // 4
		{[3]int{6, 3, 2}, SicboBig},   // 11
		{[3]int{5, 3, 2}, SicboSmall}, // 10
	}
	for _, tc := range cases {
		if got := SicboResult(tc.dice); got != tc.want {
			t.Errorf("SicboResult(%v) = %q, want %q", tc.dice, got, tc.want)
		}
	}
	if got := SicboSum([3]int{2, 3, 4}); got != 9 {
		t.Errorf("SicboSum = %d, want 9", got)
	}
}

func TestSicboRTP(t *testing.T) {
	rtps := DefaultSicbo().RTP()
	if len(rtps) != len(SicboBetTypes) {
		t.Fatalf("RTP 应覆盖全部 %d 个注项, got %d", len(SicboBetTypes), len(rtps))
	}
	want := 105.0 / 216.0 * 2 // 每注项命中 105/216，默认赔率 2
	for _, bet := range SicboBetTypes {
		rtp, ok := rtps[bet]
		if !ok {
			t.Fatalf("缺少注项 %s 的回报率", bet)
		}
		if diff := rtp - want; diff > 1e-9 || diff < -1e-9 {
			t.Fatalf("注项 %s 理论回报 = %v, want %v", bet, rtp, want)
		}
	}
}

func TestSicboRollDistribution(t *testing.T) {
	cfg := DefaultSicbo()
	rng := rand.New(rand.NewPCG(11, 12))
	const n = 216000
	wins := make(map[string]int, len(SicboBetTypes))
	for _, bet := range SicboBetTypes {
		wins[bet] = 0
	}
	var triples int
	for i := 0; i < n; i++ {
		dice := cfg.Roll(rng)
		for _, d := range dice {
			if d < 1 || d > 6 {
				t.Fatalf("骰子点数越界: %v", dice)
			}
		}
		if SicboResult(dice) == "triple" {
			triples++
		}
		for _, bet := range SicboBetTypes {
			m, err := cfg.Settle(bet, 1, dice)
			if err != nil {
				t.Fatalf("Settle(%s): %v", bet, err)
			}
			if m > 0 {
				wins[bet]++
			}
		}
	}
	// 每注项命中率 ≈ 105/216，围骰率 ≈ 6/216，容差 0.01
	wantWin := 105.0 / 216.0
	for _, bet := range SicboBetTypes {
		got := float64(wins[bet]) / n
		if diff := got - wantWin; diff > 0.01 || diff < -0.01 {
			t.Fatalf("注项 %s 命中率偏离: want %.4f, got %.4f", bet, wantWin, got)
		}
	}
	gotTriple := float64(triples) / n
	if diff := gotTriple - 6.0/216.0; diff > 0.005 || diff < -0.005 {
		t.Fatalf("围骰频率偏离: want %.4f, got %.4f", 6.0/216.0, gotTriple)
	}
}
