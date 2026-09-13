// Package games 实现三个娱乐玩法的核心逻辑：大转盘、老虎机、21点。
// 所有随机与结算均在服务端完成，前端只负责动画展示。
package games

import "math/rand/v2"

// Segment 转盘的一个扇区。
type Segment struct {
	Multiplier float64 `json:"multiplier"` // 中奖倍数（相对下注额）
	Weight     float64 `json:"weight"`     // 抽中权重（仅管理接口暴露，玩家接口会脱敏）
	Label      string  `json:"label"`      // 扇区显示文本
}

// WheelConfig 大转盘配置。
type WheelConfig struct {
	Segments []Segment `json:"segments"`
}

// DefaultWheel 默认 12 扇区转盘，总权重 92，期望回报 88/92 ≈ 95.65%。
// 扇区面积相同而概率不同；0x 大区与 10x 大奖相对摆放，视觉上符合直觉。
func DefaultWheel() *WheelConfig {
	return &WheelConfig{
		Segments: []Segment{
			{Multiplier: 0, Weight: 18, Label: "谢谢参与"},
			{Multiplier: 0.5, Weight: 7, Label: "0.5x"},
			{Multiplier: 1.2, Weight: 8, Label: "1.2x"},
			{Multiplier: 0.5, Weight: 7, Label: "0.5x"},
			{Multiplier: 2, Weight: 7, Label: "2x"},
			{Multiplier: 0, Weight: 18, Label: "谢谢参与"},
			{Multiplier: 3, Weight: 4, Label: "3x"},
			{Multiplier: 1.2, Weight: 7, Label: "1.2x"},
			{Multiplier: 0.5, Weight: 6, Label: "0.5x"},
			{Multiplier: 10, Weight: 1, Label: "10x"},
			{Multiplier: 2, Weight: 7, Label: "2x"},
			{Multiplier: 5, Weight: 2, Label: "5x"},
		},
	}
}

// Spin 按权重随机抽一个扇区，返回扇区下标。
func (w *WheelConfig) Spin(rng *rand.Rand) int {
	total := 0.0
	for _, s := range w.Segments {
		total += s.Weight
	}
	r := rng.Float64() * total
	var acc float64
	for i, s := range w.Segments {
		acc += s.Weight
		if r < acc {
			return i
		}
	}
	return len(w.Segments) - 1
}

// RTP 返回理论回报率，用于配置校验与自检。
func (w *WheelConfig) RTP() float64 {
	var ev, total float64
	for _, s := range w.Segments {
		ev += s.Weight * s.Multiplier
		total += s.Weight
	}
	if total == 0 {
		return 0
	}
	return ev / total
}
