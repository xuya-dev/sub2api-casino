package games

import (
	"fmt"
	"math/rand/v2"
)

// 骰宝四个注项。
const (
	SicboBig   = "big"   // 大：点数和 11-17
	SicboSmall = "small" // 小：点数和 4-10
	SicboOdd   = "odd"   // 单：点数和为奇数
	SicboEven  = "even"  // 双：点数和为偶数
)

// SicboBetTypes 合法注项列表。
var SicboBetTypes = []string{SicboBig, SicboSmall, SicboOdd, SicboEven}

// SicboConfig 骰宝配置：四个注项各自的赔付倍数（含本金，2 = 1:1）。
type SicboConfig struct {
	Big   float64 `json:"big"`
	Small float64 `json:"small"`
	Odd   float64 `json:"odd"`
	Even  float64 `json:"even"`
}

// DefaultSicbo 默认赔率：四个注项均为 2（1:1）。
// 216 种等概率组合中围骰占 6 种（判负），其余 210 种里每个注项恰好命中 105 种，
// 理论回报 = 105/216 × 2 ≈ 97.22%（庄家优势约 2.78%）。
func DefaultSicbo() *SicboConfig {
	return &SicboConfig{Big: 2, Small: 2, Odd: 2, Even: 2}
}

// Roll 掷三颗骰子（每颗 1-6），与其它游戏一致使用调用方传入的随机源。
func (c *SicboConfig) Roll(rng *rand.Rand) [3]int {
	return [3]int{rng.IntN(6) + 1, rng.IntN(6) + 1, rng.IntN(6) + 1}
}

// SicboSum 三颗骰子点数和。
func SicboSum(dice [3]int) int { return dice[0] + dice[1] + dice[2] }

// isTriple 围骰：三颗骰子点数相同。
func isTriple(dice [3]int) bool { return dice[0] == dice[1] && dice[1] == dice[2] }

// SicboResult 掷骰结果分类：围骰（三颗相同）返回 "triple"，
// 否则按点数和返回 SicboBig（11-17）或 SicboSmall（4-10）。
// 单/双由点数和的奇偶性直接可得，前端按需自行展示。
func SicboResult(dice [3]int) string {
	if isTriple(dice) {
		return "triple"
	}
	if SicboSum(dice) >= 11 {
		return SicboBig
	}
	return SicboSmall
}

// Settle 骰宝结算：返回下注注项的赔付倍数（含本金，判负为 0）。
// 规则：大 11-17 / 小 4-10 / 单 奇数和 / 双 偶数和；
// 围骰（三颗相同）通杀——四个注项一律判负。betType 非法或 bet 非正时报错。
func (c *SicboConfig) Settle(betType string, bet float64, dice [3]int) (float64, error) {
	if bet <= 0 {
		return 0, fmt.Errorf("下注金额必须大于 0")
	}
	var m float64
	switch betType {
	case SicboBig:
		m = c.Big
	case SicboSmall:
		m = c.Small
	case SicboOdd:
		m = c.Odd
	case SicboEven:
		m = c.Even
	default:
		return 0, fmt.Errorf("未知骰宝注项: %s", betType)
	}
	if isTriple(dice) {
		return 0, nil // 围骰通杀
	}
	sum := SicboSum(dice)
	win := false
	switch betType {
	case SicboBig:
		win = sum >= 11
	case SicboSmall:
		win = sum <= 10
	case SicboOdd:
		win = sum%2 == 1
	case SicboEven:
		win = sum%2 == 0
	}
	if !win {
		return 0, nil
	}
	return m, nil
}

// RTP 返回各注项的理论回报率。
// 每个注项在 216 种等概率组合中恰好命中 105 种（围骰判负），
// 回报率 = 105/216 × 赔率倍数。
func (c *SicboConfig) RTP() map[string]float64 {
	const p = 105.0 / 216.0
	return map[string]float64{
		SicboBig:   p * c.Big,
		SicboSmall: p * c.Small,
		SicboOdd:   p * c.Odd,
		SicboEven:  p * c.Even,
	}
}
