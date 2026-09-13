package games

import (
	"math"
	"math/rand/v2"
	"testing"
)

// probSum 按序累加档位概率（与 Normalize/Validate 的求和顺序一致）。
func probSum(tiers []ScratchTier) float64 {
	var s float64
	for _, t := range tiers {
		s += t.Probability
	}
	return s
}

// TestScratchDefaultConfig 默认配置：校验通过、各概率表和恰为 1、
// 三种玩法 RTP 均与理论值一致。
func TestScratchDefaultConfig(t *testing.T) {
	s := DefaultScratch()
	if err := s.Validate(); err != nil {
		t.Fatalf("默认配置应通过校验: %v", err)
	}
	if got := probSum(s.Tiers); got != 1.0 {
		t.Fatalf("经典档位概率和应恰为 1, got %.20f", got)
	}
	if got := probSum(s.Lucky7.Tiers); got != 1.0 {
		t.Fatalf("幸运7档位概率和应恰为 1, got %.20f", got)
	}
	var lineSum float64
	for _, sym := range s.Lines.Symbols {
		lineSum += sym.Probability
	}
	if lineSum != 1.0 {
		t.Fatalf("连线符号概率和应恰为 1, got %.20f", lineSum)
	}
	if max := s.MaxMultiplier(); max != 88.888 {
		t.Fatalf("最高倍数应为 88.888, got %v", max)
	}
	rtp := s.RTP()
	// classic = 0.29 × 3.238888 = 0.93927752
	if got := rtp[string(ScratchModeClassic)]; math.Abs(got-0.93927752) > 1e-9 {
		t.Fatalf("classic RTP = %.12f, want 0.93927752", got)
	}
	// lucky7 = 7 × 0.085 × 1.595444 = 0.94928918
	if got := rtp[string(ScratchModeLucky7)]; math.Abs(got-0.94928918) > 1e-9 {
		t.Fatalf("lucky7 RTP = %.12f, want 0.94928918", got)
	}
	// lines = 0.27 × 3.43388 = 0.9271476
	if got := rtp[string(ScratchModeLines)]; math.Abs(got-0.9271476) > 1e-9 {
		t.Fatalf("lines RTP = %.12f, want 0.9271476", got)
	}
}

// TestScratchNormalize 归一化：浮点误差/脏输入 → 概率和恰为 1、可过校验、幂等。
func TestScratchNormalize(t *testing.T) {
	// 10 × 0.1 按序求和在 float64 下存在误差（≠ 1）
	s := &ScratchConfig{WinRate: 0.3, Tiers: []ScratchTier{
		{1, 0.1}, {2, 0.1}, {5, 0.1}, {10, 0.1}, {20, 0.1},
		{50, 0.1}, {88.888, 0.1}, {3, 0.1}, {4, 0.1}, {8, 0.1},
	}}
	if probSum(s.Tiers) == 1.0 {
		t.Fatal("前置条件失效: 10×0.1 按序求和应存在浮点误差")
	}
	s.Normalize()
	s.Lucky7 = DefaultLucky7() // Validate 会校验幸运7/连线子配置，此处补默认值
	s.Lines = DefaultLines()
	if got := probSum(s.Tiers); got != 1.0 {
		t.Fatalf("归一化后概率和应恰为 1, got %.20f", got)
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("归一化后应通过校验: %v", err)
	}
	before := probSum(s.Tiers)
	s.Normalize()
	if probSum(s.Tiers) != before {
		t.Fatal("Normalize 应幂等")
	}

	// 负概率截 0 后等比归一
	s2 := &ScratchConfig{Tiers: []ScratchTier{{1, 1.5}, {2, -0.5}}}
	s2.Normalize()
	if probSum(s2.Tiers) != 1.0 || s2.Tiers[1].Probability != 0 {
		t.Fatalf("负概率应截 0 并归一: %+v", s2.Tiers)
	}
	// 全部非正：兜底为首档概率 1
	s3 := &ScratchConfig{Tiers: []ScratchTier{{1, -1}, {2, 0}}}
	s3.Normalize()
	if s3.Tiers[0].Probability != 1 || probSum(s3.Tiers) != 1.0 {
		t.Fatalf("全负概率应兜底为首档 1: %+v", s3.Tiers)
	}
	// 空档位：Normalize 为空操作
	(&ScratchConfig{}).Normalize()

	// 连线符号概率同样归一
	l := &LinesConfig{WinRate: 0.3, Symbols: []LineSymbol{
		{ID: "a", Multiplier: 1, Probability: 0.7},
		{ID: "b", Multiplier: 2, Probability: 0.7},
	}}
	l.Normalize()
	var sum float64
	for _, sym := range l.Symbols {
		sum += sym.Probability
	}
	if sum != 1.0 {
		t.Fatalf("连线符号概率归一后和应恰为 1, got %.20f", sum)
	}
}

// TestScratchValidate 配置校验：中奖率范围、档位非空、概率和为 1（±1e-6）、倍数与概率非负。
func TestScratchValidate(t *testing.T) {
	if err := DefaultScratch().Validate(); err != nil {
		t.Fatalf("默认配置应通过校验: %v", err)
	}
	base := []ScratchTier{{1, 0.6}, {2, 0.4}}
	bad := []struct {
		name string
		cfg  *ScratchConfig
	}{
		{"中奖率 0", &ScratchConfig{WinRate: 0, Tiers: base}},
		{"中奖率为负", &ScratchConfig{WinRate: -0.1, Tiers: base}},
		{"中奖率 >1", &ScratchConfig{WinRate: 1.0001, Tiers: base}},
		{"档位为空", &ScratchConfig{WinRate: 0.3}},
		{"概率和 0.98", &ScratchConfig{WinRate: 0.3, Tiers: []ScratchTier{{1, 0.98}}}},
		{"概率和 1.5", &ScratchConfig{WinRate: 0.3, Tiers: []ScratchTier{{1, 0.6}, {2, 0.9}}}},
		{"概率和超容差", &ScratchConfig{WinRate: 0.3, Tiers: []ScratchTier{{1, 0.5}, {2, 0.50001}}}},
		{"倍数为负", &ScratchConfig{WinRate: 0.3, Tiers: []ScratchTier{{1, 0.5}, {-2, 0.5}}}},
		{"概率为负", &ScratchConfig{WinRate: 0.3, Tiers: []ScratchTier{{1, 1.5}, {2, -0.5}}}},
	}
	for _, tc := range bad {
		if err := tc.cfg.Validate(); err == nil {
			t.Fatalf("%s 应校验失败", tc.name)
		}
	}
	// 幸运7格数为 0 需单独构造（上面 shared Lucky7 已带默认值）
	zeroCells := DefaultScratch()
	zeroCells.Lucky7.Cells = 0
	if err := zeroCells.Validate(); err == nil {
		t.Fatal("幸运7格数为 0 应校验失败")
	}
	// 幸运7命中率越界
	badHit := DefaultScratch()
	badHit.Lucky7.HitRate = 0
	if err := badHit.Validate(); err == nil {
		t.Fatal("幸运7命中率 0 应校验失败")
	}
	// 连线符号不足
	fewSymbols := DefaultScratch()
	fewSymbols.Lines.Symbols = fewSymbols.Lines.Symbols[:1]
	if err := fewSymbols.Validate(); err == nil {
		t.Fatal("连线符号少于 2 应校验失败")
	}
	ok := []struct {
		name string
		cfg  *ScratchConfig
	}{
		{"中奖率 =1", func() *ScratchConfig { s := DefaultScratch(); s.WinRate = 1; return s }()},
		{"概率和在容差内", func() *ScratchConfig {
			s := DefaultScratch()
			s.Tiers = []ScratchTier{{1, 0.5}, {2, 0.5 + 1e-7}}
			return s
		}()},
		{"倍数为 0", func() *ScratchConfig {
			s := DefaultScratch()
			s.Tiers = []ScratchTier{{0, 1}}
			return s
		}()},
	}
	for _, tc := range ok {
		if err := tc.cfg.Validate(); err != nil {
			t.Fatalf("%s 应通过校验: %v", tc.name, err)
		}
	}
}

// TestScratchCardTiers 经典单卡逐档位结算：边界值精确命中各档，未中奖倍数为 0。
func TestScratchCardTiers(t *testing.T) {
	s := DefaultScratch()
	// winRoll ∈ [0, 0.29) 中奖；winRoll ≥ WinRate 判不中奖
	if m := s.multiplierAt(0.29, 0); m != 0 {
		t.Fatalf("winRoll=WinRate 应判不中奖, got %v", m)
	}
	if m := s.multiplierAt(0.999, 0.5); m != 0 {
		t.Fatalf("winRoll 超过 WinRate 应判不中奖, got %v", m)
	}
	// 中奖条件下按累积概率选档，档位边界精确命中：
	// [0,0.60)→1x [0.60,0.80)→2x [0.80,0.90)→5x [0.90,0.96)→10x
	// [0.96,0.99)→20x [0.99,0.999)→50x [0.999,1)→88.888x
	cases := []struct {
		tierRoll float64
		want     float64
	}{
		{0, 1}, {0.5999, 1},
		{0.6, 2}, {0.7999, 2},
		{0.8, 5}, {0.8999, 5},
		{0.9, 10}, {0.9599, 10},
		{0.96, 20}, {0.9899, 20},
		{0.99, 50}, {0.9989, 50},
		{0.999, 88.888}, {0.9999, 88.888},
	}
	for _, tc := range cases {
		if m := s.multiplierAt(0, tc.tierRoll); m != tc.want {
			t.Fatalf("tierRoll=%v: 倍数 want %v got %v", tc.tierRoll, tc.want, m)
		}
	}
}

// TestScratchPlay 面值折算：prize = 倍数 × 面值；必中/必不中配置的行为；
// 卡面涂层格内容（cells）与玩法一致。
func TestScratchPlay(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	// WinRate=1 且单档：必中该档，奖金 = 88.888 × 10 = 888.88
	allWin := &ScratchConfig{WinRate: 1, Tiers: []ScratchTier{{Multiplier: 88.888, Probability: 1}}}
	for i := 0; i < 100; i++ {
		card := allWin.Play(rng, 10)
		if card.Multiplier != 88.888 || math.Abs(card.Prize-888.88) > 1e-9 {
			t.Fatalf("面值 10 必中 88.888 倍: got %+v", card)
		}
		if len(card.Cells) != 1 || card.Cells[0].Kind != "mult" {
			t.Fatalf("经典卡应为单 mult 格: got %+v", card.Cells)
		}
	}
	// WinRate=0 恒不中奖（仅行为测试；Validate 不接受 0）
	never := &ScratchConfig{WinRate: 0, Tiers: []ScratchTier{{Multiplier: 5, Probability: 0.6}, {Multiplier: 2, Probability: 0.4}}}
	for i := 0; i < 100; i++ {
		if card := never.Play(rng, 2); card.Multiplier != 0 || card.Prize != 0 || card.Cells[0].Kind != "dud" {
			t.Fatalf("WinRate=0 应恒不中奖: got %+v", card)
		}
	}
}

// TestLucky7Play 幸运7：7 格、命中率边界（必中/必不中）、倍数累加、面值折算。
func TestLucky7Play(t *testing.T) {
	rng := rand.New(rand.NewPCG(11, 12))
	// 必中且单档 88.888：7 格全中，倍数 = 7 × 88.888
	allHit := &Lucky7Config{Cells: 7, HitRate: 1, Tiers: []ScratchTier{{Multiplier: 88.888, Probability: 1}}}
	card := allHit.Play(rng, 10)
	if math.Abs(card.Multiplier-7*88.888) > 1e-9 || math.Abs(card.Prize-7*888.88) > 1e-9 {
		t.Fatalf("全中 88.888 档应得 7×88.888 倍: got %+v", card)
	}
	if len(card.Cells) != 7 {
		t.Fatalf("幸运7卡应有 7 格: got %d", len(card.Cells))
	}
	// 必不中：全 dud、倍数与奖金为 0
	neverHit := &Lucky7Config{Cells: 7, HitRate: 0, Tiers: []ScratchTier{{Multiplier: 5, Probability: 1}}}
	card = neverHit.Play(rng, 10)
	if card.Multiplier != 0 || card.Prize != 0 {
		t.Fatalf("命中率 0 应恒不中奖: got %+v", card)
	}
	for _, cell := range card.Cells {
		if cell.Kind != "dud" {
			t.Fatalf("未命中格应为 dud: got %+v", cell)
		}
	}
	// 命中率 0.5 抽样：命中格 kind=seven 且倍数大于 0，与总量对账
	half := &Lucky7Config{Cells: 7, HitRate: 0.5, Tiers: []ScratchTier{
		{Multiplier: 1, Probability: 0.5}, {Multiplier: 3, Probability: 0.5},
	}}
	for i := 0; i < 200; i++ {
		card := half.Play(rng, 2)
		var sum float64
		var sevens int
		for _, cell := range card.Cells {
			if cell.Kind == "seven" {
				sevens++
				if cell.Multiplier != 1 && cell.Multiplier != 3 {
					t.Fatalf("命中格倍数只能为 1 或 3: got %+v", cell)
				}
				sum += cell.Multiplier
			}
		}
		if sum != card.Multiplier || math.Abs(card.Prize-card.Multiplier*2) > 1e-9 {
			t.Fatalf("倍数累加与奖金折算不符: card=%+v sum=%v", card, sum)
		}
		if (card.Multiplier > 0) != (sevens > 0) {
			t.Fatalf("中奖判定与 seven 格数不一致: card=%+v sevens=%d", card, sevens)
		}
	}
}

// TestLinesGrid 连线宫格：出奖时恰一条中奖线且倍数与符号一致；未出奖时无线；
// 所有格子的符号均来自符号表。
func TestLinesGrid(t *testing.T) {
	cfg := DefaultLines()
	rng := rand.New(rand.NewPCG(21, 22))
	valid := make(map[string]bool)
	for _, sym := range cfg.Symbols {
		valid[sym.ID] = true
	}
	var wins, losses int
	for i := 0; i < 3000; i++ {
		card := cfg.Play(rng, 1)
		if len(card.Cells) != 9 {
			t.Fatalf("连线卡应有 9 格: got %d", len(card.Cells))
		}
		for _, cell := range card.Cells {
			if !valid[cell.Symbol] {
				t.Fatalf("宫格出现未知符号 %q", cell.Symbol)
			}
		}
		// 独立重扫中奖线，必须与返回的 win_lines 一致
		var found [][3]int
		for _, ln := range scratchLines {
			a, b, c := card.Cells[ln[0]].Symbol, card.Cells[ln[1]].Symbol, card.Cells[ln[2]].Symbol
			if a == b && b == c {
				found = append(found, ln)
			}
		}
		if len(found) > 1 {
			t.Fatalf("构造保证至多一条中奖线: got %d (%+v)", len(found), found)
		}
		if len(found) != len(card.WinLines) {
			t.Fatalf("win_lines 与重扫不一致: got %+v want %+v", card.WinLines, found)
		}
		if len(found) == 1 {
			wins++
			wantMult := cfg.symbolMult(card.Cells[found[0][0]].Symbol)
			if card.Multiplier != wantMult {
				t.Fatalf("中奖倍数应等于该符号倍数: got %v want %v", card.Multiplier, wantMult)
			}
		} else {
			losses++
			if card.Multiplier != 0 || card.Prize != 0 || len(card.WinLines) != 0 {
				t.Fatalf("无中奖线时倍数应为 0: got %+v", card)
			}
		}
	}
	winRate := float64(wins) / float64(wins+losses)
	t.Logf("连线出奖率: %.4f（理论 %.4f）", winRate, cfg.WinRate)
	if math.Abs(winRate-cfg.WinRate) > 0.02 {
		t.Fatalf("出奖率偏离 WinRate 超过 2pp: got %.4f want %.4f", winRate, cfg.WinRate)
	}
}

// TestScratchRTP 三种玩法默认 RTP 均落在 92%~96% 目标区间。
func TestScratchRTP(t *testing.T) {
	for _, m := range AllScratchModes {
		rtp := DefaultScratch().RTP()[string(m)]
		t.Logf("刮刮乐[%s]理论回报率: %.6f", m, rtp)
		if rtp < 0.92 || rtp > 0.96 {
			t.Fatalf("玩法 %s RTP 超出 92%%~96%%: %.5f", m, rtp)
		}
	}
}

// TestScratchMonteCarlo 10 万张 Monte-Carlo（经典）：中奖率应落在 WinRate ± 0.5 个
// 百分点；实际回报率与理论 RTP 一致（容差按单卡回报方差取 2.5σ ≈ 0.03）。
func TestScratchMonteCarlo(t *testing.T) {
	s := DefaultScratch()
	rng := rand.New(rand.NewPCG(42, 43))
	const n = 100000
	var wins int
	var returned float64
	for i := 0; i < n; i++ {
		card := s.Play(rng, 1)
		if card.Multiplier > 0 {
			wins++
		}
		returned += card.Prize
	}
	winRate := float64(wins) / n
	t.Logf("10 万张模拟中奖率: %.4f（理论 %.4f）", winRate, s.WinRate)
	if math.Abs(winRate-s.WinRate) > 0.005 {
		t.Fatalf("中奖率偏离 WinRate 超过 0.5pp: got %.4f want %.4f", winRate, s.WinRate)
	}
	rtp := returned / n
	t.Logf("10 万张模拟 RTP: %.4f（理论 %.5f）", rtp, s.classicRTP())
	if math.Abs(rtp-s.classicRTP()) > 0.03 {
		t.Fatalf("模拟 RTP 偏离理论值: sim %.4f exact %.5f", rtp, s.classicRTP())
	}
}
