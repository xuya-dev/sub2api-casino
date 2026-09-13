package games

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
)

// ScratchMode 刮刮乐玩法模式。
type ScratchMode string

const (
	ScratchModeClassic = "classic" // 经典刮奖：单格涂层直接开出倍数
	ScratchModeLucky7  = "lucky7"  // 幸运7：7 格涂层寻找「幸运7」，命中格倍数叠加
	ScratchModeLines   = "lines"   // 幸运连线：3×3 宫格 8 条线，三连相同符号即中奖
)

// AllScratchModes 全部合法玩法（handler 校验用）。
var AllScratchModes = []ScratchMode{ScratchModeClassic, ScratchModeLucky7, ScratchModeLines}

// ScratchTier 刮刮乐奖品档位：倍数 + （中奖条件下）抽中概率。
type ScratchTier struct {
	Multiplier  float64 `json:"multiplier"`
	Probability float64 `json:"probability"`
}

// Lucky7Config 幸运7玩法：cells 个涂层格，每格独立按 HitRate 判定是否藏有
// 「幸运7」，命中后按 Tiers 抽倍数；一张卡上多个 7 的倍数累加。
//
// 默认参数 RTP = cells × HitRate × Σ(prob×mult)
//          = 7 × 0.085 × 1.595444 ≈ 0.9493（≈94.9%）。
type Lucky7Config struct {
	Cells   int           `json:"cells"`   // 涂层格数（默认 7）
	HitRate float64       `json:"hit_rate"` // 单格命中幸运7的概率 ∈ (0,1]
	Tiers   []ScratchTier `json:"tiers"`   // 命中后的倍数档位（概率和为 1）
}

// LineSymbol 幸运连线玩法的一个符号：三连成线时派 Multiplier 倍。
type LineSymbol struct {
	ID          string  `json:"id"`
	Multiplier  float64 `json:"multiplier"`
	Probability float64 `json:"probability"` // 中奖条件下抽中该符号的概率（全部和为 1）
}

// LinesConfig 幸运连线玩法：3×3 宫格共 8 条线（3 横 3 竖 2 对角）。
// 先按 WinRate 判定本卡是否出奖；出奖时随机选一个符号与一条线铺成三连，
// 其余格子防误中奖填充（保证至多一条中奖线）；未出奖时全盘防三连填充。
//
// 默认参数 RTP = WinRate × Σ(prob×mult) = 0.27 × 3.43388 ≈ 0.9271（≈92.7%）。
type LinesConfig struct {
	WinRate float64      `json:"win_rate"` // 单卡出现中奖连线的概率 ∈ (0,1]
	Symbols []LineSymbol `json:"symbols"`  // 符号表（≥2 个，概率和为 1）
}

// ScratchConfig 刮刮乐配置：「面值 + 数量」模型 × 三种玩法。
// 玩家选卡片面值（bet）与购买数量（1~10 张），一次性购买并逐张独立结算。
type ScratchConfig struct {
	WinRate float64       `json:"win_rate"` // 经典玩法单卡中奖率 ∈ (0,1]
	Tiers   []ScratchTier `json:"tiers"`    // 经典玩法奖品档位表
	Lucky7  Lucky7Config  `json:"lucky7"`   // 幸运7玩法
	Lines   LinesConfig   `json:"lines"`    // 幸运连线玩法
}

// DefaultScratch 默认配置（三种玩法 RTP 均落在 92%~96% 目标区间）：
//   - classic: 0.29 × 3.238888 ≈ 0.9393（≈93.9%）
//   - lucky7 : 7 × 0.085 × 1.595444 ≈ 0.9493（≈94.9%）
//   - lines  : 0.27 × 3.43388 ≈ 0.9271（≈92.7%）
//
// 最高单卡奖金（classic/lines）= 88.888 × 面值（面值 10 时为 888.88）。
func DefaultScratch() *ScratchConfig {
	s := &ScratchConfig{
		WinRate: 0.29,
		Tiers: []ScratchTier{
			{Multiplier: 1, Probability: 0.60},
			{Multiplier: 2, Probability: 0.20},
			{Multiplier: 5, Probability: 0.10},
			{Multiplier: 10, Probability: 0.06},
			{Multiplier: 20, Probability: 0.03},
			{Multiplier: 50, Probability: 0.009},
			{Multiplier: 88.888, Probability: 0.001},
		},
		Lucky7: DefaultLucky7(),
		Lines:  DefaultLines(),
	}
	s.Normalize()
	return s
}

// DefaultLucky7 幸运7默认参数。
// 档位期望 Σ(prob×mult) = 1×0.80 + 2×0.1155 + 3×0.045 + 5×0.022 + 10×0.011 +
// 20×0.0045 + 50×0.0015 + 88.888×0.0005 = 1.595444。
func DefaultLucky7() Lucky7Config {
	l := Lucky7Config{
		Cells:   7,
		HitRate: 0.085,
		Tiers: []ScratchTier{
			{Multiplier: 1, Probability: 0.80},
			{Multiplier: 2, Probability: 0.1155},
			{Multiplier: 3, Probability: 0.045},
			{Multiplier: 5, Probability: 0.022},
			{Multiplier: 10, Probability: 0.011},
			{Multiplier: 20, Probability: 0.0045},
			{Multiplier: 50, Probability: 0.0015},
			{Multiplier: 88.888, Probability: 0.0005},
		},
	}
	l.Normalize()
	return l
}

// DefaultLines 幸运连线默认参数。
// 符号期望 Σ(prob×mult) = 1×0.40 + 1.5×0.20 + 2×0.16 + 3×0.10 + 5×0.065 +
// 10×0.04 + 20×0.025 + 88.888×0.01 = 3.43388。
func DefaultLines() LinesConfig {
	l := LinesConfig{
		WinRate: 0.27,
		Symbols: []LineSymbol{
			{ID: "cherry", Multiplier: 1, Probability: 0.40},
			{ID: "lemon", Multiplier: 1.5, Probability: 0.20},
			{ID: "grape", Multiplier: 2, Probability: 0.16},
			{ID: "melon", Multiplier: 3, Probability: 0.10},
			{ID: "bell", Multiplier: 5, Probability: 0.065},
			{ID: "diamond", Multiplier: 10, Probability: 0.04},
			{ID: "crown", Multiplier: 20, Probability: 0.025},
			{ID: "seven", Multiplier: 88.888, Probability: 0.01},
		},
	}
	l.Normalize()
	return l
}

// ScratchCell 卡面单个涂层格刮开后的奖层内容。
type ScratchCell struct {
	// kind: "mult" 经典倍数格 / "seven" 幸运7格 / "symbol" 连线符号格 / "dud" 未中奖格
	Kind       string  `json:"kind"`
	Symbol     string  `json:"symbol,omitempty"`     // symbol 格的符号 ID
	Multiplier float64 `json:"multiplier,omitempty"` // 该格派彩倍数（0 时省略）
}

// ScratchCard 一张已刮开的卡。
type ScratchCard struct {
	Multiplier float64  `json:"multiplier"`          // 中奖倍数（未中奖为 0）
	Prize      float64  `json:"prize"`               // 奖金 = 倍数 × 卡片面值
	Cells      []ScratchCell `json:"cells"`           // 涂层格内容（classic 1 格 / lucky7 7 格 / lines 9 格）
	WinLines   [][3]int `json:"win_lines,omitempty"` // 连线玩法的下标（行优先 0-8）
}

// scratchLines 3×3 宫格的 8 条线（行优先下标）：三横、三竖、两对角。
var scratchLines = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

// scratchCellLines 每个格子参与的连线编号（防误中奖填充用）。
var scratchCellLines = func() map[int][]int {
	m := make(map[int][]int, 9)
	for i, ln := range scratchLines {
		for _, c := range ln {
			m[c] = append(m[c], i)
		}
	}
	return m
}()

// ---- 抽奖公共件 ----

// drawTier 按累积概率从档位表抽一个倍数（概率和须为 1，越界兜底最后一档）。
func drawTier(rng *rand.Rand, tiers []ScratchTier) float64 {
	roll := rng.Float64()
	var acc float64
	for _, tier := range tiers {
		acc += tier.Probability
		if roll < acc {
			return tier.Multiplier
		}
	}
	return tiers[len(tiers)-1].Multiplier
}

// normalizeProbList 通用概率归一：负值截 0 → 全零兜底首项为 1 → 等比归一 →
// 残余浮点误差并入最大项，使概率按序求和在 float64 意义下恰为 1。
func normalizeProbList(n int, prob func(int) float64, setProb func(int, float64)) {
	if n == 0 {
		return
	}
	total := 0.0
	for i := 0; i < n; i++ {
		p := prob(i)
		if p < 0 {
			p = 0
			setProb(i, 0)
		}
		total += p
	}
	switch {
	case total <= 0:
		setProb(0, 1)
	case total != 1:
		inv := 1 / total
		for i := 0; i < n; i++ {
			setProb(i, prob(i)*inv)
		}
	}
	big := 0
	for i := 1; i < n; i++ {
		if prob(i) > prob(big) {
			big = i
		}
	}
	for k := 0; k < 8; k++ {
		sum := 0.0
		for i := 0; i < n; i++ {
			sum += prob(i)
		}
		if sum == 1 {
			break
		}
		setProb(big, prob(big)+1-sum)
	}
}

// normalizeTiers 档位概率归一。
func normalizeTiers(tiers []ScratchTier) {
	normalizeProbList(len(tiers),
		func(i int) float64 { return tiers[i].Probability },
		func(i int, p float64) { tiers[i].Probability = p })
}

// Normalize 清理三种玩法概率的浮点误差，使各概率表之和恰为 1。
// 配置加载/保存时调用（见 gamemgr.Config.applyDefaults）。
func (s *ScratchConfig) Normalize() {
	normalizeTiers(s.Tiers)
	normalizeTiers(s.Lucky7.Tiers)
	normalizeLineSymbols(s.Lines.Symbols)
}

func normalizeLineSymbols(symbols []LineSymbol) {
	normalizeProbList(len(symbols),
		func(i int) float64 { return symbols[i].Probability },
		func(i int, p float64) { symbols[i].Probability = p })
}

// validateTiers 档位表合法性：非空、倍数与概率非负、概率和为 1（±1e-6）。
func validateTiers(label string, tiers []ScratchTier) error {
	if len(tiers) == 0 {
		return fmt.Errorf("%s奖品档位不能为空", label)
	}
	var sum float64
	for i, t := range tiers {
		if t.Multiplier < 0 {
			return fmt.Errorf("%s第 %d 档倍数不能为负", label, i+1)
		}
		if t.Probability < 0 {
			return fmt.Errorf("%s第 %d 档概率不能为负", label, i+1)
		}
		sum += t.Probability
	}
	if math.Abs(sum-1) > 1e-6 {
		return fmt.Errorf("%s奖品档位概率之和须为 1（当前 %.8f）", label, sum)
	}
	return nil
}

// Validate 三种玩法配置合法性。
func (s *ScratchConfig) Validate() error {
	if s.WinRate <= 0 || s.WinRate > 1 {
		return errors.New("刮刮乐中奖率须在 (0,1] 之内")
	}
	if err := validateTiers("刮刮乐", s.Tiers); err != nil {
		return err
	}
	if s.Lucky7.Cells < 1 || s.Lucky7.Cells > 12 {
		return errors.New("幸运7涂层格数须在 1~12 之间")
	}
	if s.Lucky7.HitRate <= 0 || s.Lucky7.HitRate > 1 {
		return errors.New("幸运7单格命中率须在 (0,1] 之内")
	}
	if err := validateTiers("幸运7", s.Lucky7.Tiers); err != nil {
		return err
	}
	if s.Lines.WinRate <= 0 || s.Lines.WinRate > 1 {
		return errors.New("幸运连线中奖率须在 (0,1] 之内")
	}
	if len(s.Lines.Symbols) < 2 {
		return errors.New("幸运连线符号数量不能少于 2")
	}
	var sum float64
	for i, sym := range s.Lines.Symbols {
		if sym.ID == "" {
			return fmt.Errorf("幸运连线第 %d 个符号缺少 ID", i+1)
		}
		if sym.Multiplier < 0 || sym.Probability < 0 {
			return fmt.Errorf("幸运连线符号 %s 的倍数与概率不能为负", sym.ID)
		}
		sum += sym.Probability
	}
	if math.Abs(sum-1) > 1e-6 {
		return fmt.Errorf("幸运连线符号概率之和须为 1（当前 %.8f）", sum)
	}
	return nil
}

// ---- 经典玩法 ----

// Play 刮开一张面值 faceValue 的经典卡：先按 WinRate 判定中奖，中奖再抽档。
func (s *ScratchConfig) Play(rng *rand.Rand, faceValue float64) ScratchCard {
	m := s.multiplierAt(rng.Float64(), rng.Float64())
	kind := "dud"
	cell := ScratchCell{Kind: kind}
	if m > 0 {
		cell = ScratchCell{Kind: "mult", Multiplier: m}
	}
	return ScratchCard{Multiplier: m, Prize: m * faceValue, Cells: []ScratchCell{cell}}
}

// multiplierAt 经典单卡结算的纯函数形式（便于逐档位精确测试）：
// winRoll 与 WinRate 比较决定是否中奖；中奖时 tierRoll 按档位累积概率选档。
func (s *ScratchConfig) multiplierAt(winRoll, tierRoll float64) float64 {
	if winRoll >= s.WinRate || len(s.Tiers) == 0 {
		return 0
	}
	var acc float64
	for _, tier := range s.Tiers {
		acc += tier.Probability
		if tierRoll < acc {
			return tier.Multiplier
		}
	}
	return s.Tiers[len(s.Tiers)-1].Multiplier
}

// ---- 幸运7玩法 ----

// Play 刮开一张面值 faceValue 的幸运7卡：每格独立判定，命中的倍数累加。
func (l *Lucky7Config) Play(rng *rand.Rand, faceValue float64) ScratchCard {
	cells := make([]ScratchCell, l.Cells)
	var total float64
	for i := range cells {
		if rng.Float64() < l.HitRate {
			m := drawTier(rng, l.Tiers)
			cells[i] = ScratchCell{Kind: "seven", Multiplier: m}
			total += m
		} else {
			cells[i] = ScratchCell{Kind: "dud"}
		}
	}
	return ScratchCard{Multiplier: total, Prize: total * faceValue, Cells: cells}
}

// Normalize 幸运7档位概率归一。
func (l *Lucky7Config) Normalize() { normalizeTiers(l.Tiers) }

// RTP 理论回报率 = Cells × HitRate × Σ(prob × mult)。
func (l *Lucky7Config) RTP() float64 {
	var ev float64
	for _, t := range l.Tiers {
		ev += t.Probability * t.Multiplier
	}
	return float64(l.Cells) * l.HitRate * ev
}

// ---- 幸运连线玩法 ----

// Play 刮开一张面值 faceValue 的连线卡：构造 3×3 宫格并结算中奖线。
func (l *LinesConfig) Play(rng *rand.Rand, faceValue float64) ScratchCard {
	ids, winLines := l.buildGrid(rng)
	cells := make([]ScratchCell, len(ids))
	for i, id := range ids {
		cells[i] = ScratchCell{Kind: "symbol", Symbol: id, Multiplier: l.symbolMult(id)}
	}
	var total float64
	for _, ln := range winLines {
		total += l.symbolMult(ids[ln[0]])
	}
	return ScratchCard{Multiplier: total, Prize: total * faceValue, Cells: cells, WinLines: winLines}
}

// Normalize 连线符号概率归一。
func (l *LinesConfig) Normalize() { normalizeLineSymbols(l.Symbols) }

// RTP 理论回报率 = WinRate × Σ(prob × mult)。
func (l *LinesConfig) RTP() float64 {
	var ev float64
	for _, sym := range l.Symbols {
		ev += sym.Probability * sym.Multiplier
	}
	return l.WinRate * ev
}

func (l *LinesConfig) symbolMult(id string) float64 {
	for _, sym := range l.Symbols {
		if sym.ID == id {
			return sym.Multiplier
		}
	}
	return 0
}

// pickSymbol 按累积概率抽中奖符号 ID。
func (l *LinesConfig) pickSymbol(rng *rand.Rand) string {
	roll := rng.Float64()
	var acc float64
	for _, sym := range l.Symbols {
		acc += sym.Probability
		if roll < acc {
			return sym.ID
		}
	}
	return l.Symbols[len(l.Symbols)-1].ID
}

// buildGrid 构造一局 3×3 宫格：出奖时恰有一条指定中奖线，其余空格做
// 防误中奖填充（禁止形成指定线之外的三连）；未出奖时全盘防三连。
// 返回符号 ID 矩阵（行优先）与实际结算的中奖线（按构造至多一条）。
func (l *LinesConfig) buildGrid(rng *rand.Rand) (ids []string, winLines [][3]int) {
	ids = make([]string, 9)
	if rng.Float64() < l.WinRate {
		sym := l.pickSymbol(rng)
		ln := scratchLines[rng.IntN(len(scratchLines))]
		for _, c := range ln {
			ids[c] = sym
		}
	}
	for c := 0; c < 9; c++ {
		if ids[c] != "" {
			continue
		}
		placed := false
		for attempt := 0; attempt < 64 && !placed; attempt++ {
			ids[c] = l.Symbols[rng.IntN(len(l.Symbols))].ID
			placed = !l.completesAnyLine(ids, c)
		}
		if !placed { // 线性兜底：每个格子至多被禁 len(Symbols)-1 个取值，必存在可行符号
			for _, sym := range l.Symbols {
				ids[c] = sym.ID
				if !l.completesAnyLine(ids, c) {
					break
				}
			}
		}
	}
	for _, ln := range scratchLines {
		if ids[ln[0]] != "" && ids[ln[0]] == ids[ln[1]] && ids[ln[1]] == ids[ln[2]] {
			winLines = append(winLines, ln)
		}
	}
	return ids, winLines
}

// completesAnyLine 报告填充 ids[c] 后是否形成了三连（用于防误中奖填充；
// 指定中奖线的格子预填后不再被填充，故无需排除逻辑）。
func (l *LinesConfig) completesAnyLine(ids []string, c int) bool {
	for _, li := range scratchCellLines[c] {
		ln := scratchLines[li]
		if a, b, d := ids[ln[0]], ids[ln[1]], ids[ln[2]]; a != "" && a == b && b == d {
			return true
		}
	}
	return false
}

// ---- 汇总 ----

// MaxMultiplier 全玩法最高奖品倍数（幸运7按全部格子同中最高档的理论上限）。
func (s *ScratchConfig) MaxMultiplier() float64 {
	max := 0.0
	for _, t := range s.Tiers {
		if t.Multiplier > max {
			max = t.Multiplier
		}
	}
	return max
}

// ClassicMaxMultiplier 经典玩法最高倍数。
func (s *ScratchConfig) ClassicMaxMultiplier() float64 { return s.MaxMultiplier() }

// Lucky7MaxMultiplier 幸运7单卡理论上限倍数（全部格子命中最高档）。
func (s *ScratchConfig) Lucky7MaxMultiplier() float64 {
	max := 0.0
	for _, t := range s.Lucky7.Tiers {
		if t.Multiplier > max {
			max = t.Multiplier
		}
	}
	return float64(s.Lucky7.Cells) * max
}

// RTP 各玩法理论回报率（管理端预览与测试用）。
func (s *ScratchConfig) RTP() map[string]float64 {
	return map[string]float64{
		string(ScratchModeClassic): s.classicRTP(),
		string(ScratchModeLucky7):  s.Lucky7.RTP(),
		string(ScratchModeLines):   s.Lines.RTP(),
	}
}

func (s *ScratchConfig) classicRTP() float64 {
	var ev float64
	for _, t := range s.Tiers {
		ev += t.Probability * t.Multiplier
	}
	return s.WinRate * ev
}
