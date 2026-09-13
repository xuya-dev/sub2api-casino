package games

import "math/rand/v2"

// SlotSymbol 老虎机单个符号。
type SlotSymbol struct {
	ID     string `json:"id"`     // 结算用标识
	Emoji  string `json:"emoji"`  // 前端展示
	Weight int    `json:"weight"` // 每个卷轴的出现权重（仅管理接口暴露）
}

// SlotsConfig 老虎机配置：3 个卷轴单赔付线。
// 三连按 TriplePays 赔付；恰好两连按 PairPays 赔付（两连与三连互斥，只取其一）。
type SlotsConfig struct {
	Symbols    []SlotSymbol       `json:"symbols"`
	TriplePays map[string]float64 `json:"triple_pays"`
	PairPays   map[string]float64 `json:"pair_pays"`
}

// DefaultSlots 默认符号表，理论回报约 95%。
func DefaultSlots() *SlotsConfig {
	return &SlotsConfig{
		Symbols: []SlotSymbol{
			{ID: "cherry", Emoji: "🍒", Weight: 40},
			{ID: "lemon", Emoji: "🍋", Weight: 24},
			{ID: "bell", Emoji: "🔔", Weight: 15},
			{ID: "star", Emoji: "⭐", Weight: 9},
			{ID: "diamond", Emoji: "💎", Weight: 5},
			{ID: "seven", Emoji: "7️⃣", Weight: 2},
		},
		TriplePays: map[string]float64{
			"cherry": 3.5, "lemon": 4, "bell": 12, "star": 30, "diamond": 60, "seven": 200,
		},
		PairPays: map[string]float64{
			"cherry": 1, "lemon": 0.5, "bell": 1, "star": 2, "diamond": 5, "seven": 10,
		},
	}
}

// pickSymbol 按权重抽取单个符号。
func (s *SlotsConfig) pickSymbol(rng *rand.Rand) SlotSymbol {
	total := 0
	for _, sym := range s.Symbols {
		total += sym.Weight
	}
	r := rng.IntN(total)
	for _, sym := range s.Symbols {
		r -= sym.Weight
		if r < 0 {
			return sym
		}
	}
	return s.Symbols[len(s.Symbols)-1]
}

// Spin 转一次：返回三个卷轴符号与总倍数（0 表示未中奖）。
func (s *SlotsConfig) Spin(rng *rand.Rand) ([3]SlotSymbol, float64) {
	reels := [3]SlotSymbol{s.pickSymbol(rng), s.pickSymbol(rng), s.pickSymbol(rng)}

	if reels[0].ID == reels[1].ID && reels[1].ID == reels[2].ID {
		return reels, s.TriplePays[reels[0].ID]
	}
	// 恰好两连：任意两个位置相同且第三格不同
	for _, sym := range s.Symbols {
		n := 0
		for _, r := range reels {
			if r.ID == sym.ID {
				n++
			}
		}
		if n == 2 {
			return reels, s.PairPays[sym.ID]
		}
	}
	return reels, 0
}

// RTP 返回理论回报率（对符号分布求解析期望），用于配置校验与自检。
func (s *SlotsConfig) RTP() float64 {
	total := 0
	for _, sym := range s.Symbols {
		total += sym.Weight
	}
	if total == 0 {
		return 0
	}
	var ev float64
	for _, a := range s.Symbols {
		pa := float64(a.Weight) / float64(total)
		for _, b := range s.Symbols {
			pb := float64(b.Weight) / float64(total)
			for _, c := range s.Symbols {
				pc := float64(c.Weight) / float64(total)
				var mult float64
				if a.ID == b.ID && b.ID == c.ID {
					mult = s.TriplePays[a.ID]
				} else if a.ID == b.ID && b.ID != c.ID {
					mult = s.PairPays[a.ID]
				} else if a.ID == c.ID && a.ID != b.ID {
					mult = s.PairPays[a.ID]
				} else if b.ID == c.ID && b.ID != a.ID {
					mult = s.PairPays[b.ID]
				}
				ev += pa * pb * pc * mult
			}
		}
	}
	return ev
}
