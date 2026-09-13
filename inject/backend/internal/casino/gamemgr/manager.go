// Package gamemgr 管理游戏配置的加载/保存/热更新。
// 配置持久化在娱乐场自己的 casino_settings 表，管理页面保存后立即生效，无需重启。
package gamemgr

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/casino/games"
)

// SettingStore 配置持久化所需的最小存储接口（由 casino 包的 *Store 实现，
// 通过接口解耦避免 gamemgr → casino 的 import 环）。
type SettingStore interface {
	GetSettingJSON(ctx context.Context, key string) (string, error)
	PutSettingJSON(ctx context.Context, key, value string) error
}

// Config 完整游戏配置（管理页可编辑）。
type Config struct {
	// Enabled 娱乐模式总开关；指针语义：历史配置缺省该键时视为开启。
	Enabled        *bool                  `json:"enabled"`
	MinBet         float64                `json:"min_bet"`
	MaxBet         float64                `json:"max_bet"`
	DailyLossLimit float64                `json:"daily_loss_limit"` // 0 = 不限
	Wheel          *games.WheelConfig     `json:"wheel"`
	Slots          *games.SlotsConfig     `json:"slots"`
	Scratch        *games.ScratchConfig   `json:"scratch"`
	Blackjack      *games.BlackjackConfig `json:"blackjack"`
	Sicbo          *games.SicboConfig     `json:"sicbo"`
	Baccarat       *games.BaccaratConfig  `json:"baccarat"`
}

// IsEnabled 报告娱乐模式是否开启（Enabled 缺省/为 nil 时默认开启）。
func (c *Config) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}

const settingKey = "games_v1"

// Manager 持有当前配置，支持并发读取与热替换。
type Manager struct {
	mu  sync.RWMutex
	cfg *Config
	st  SettingStore
}

// New 创建管理器并从数据库加载配置；库里没有（首次启动）则写入默认配置。
func New(ctx context.Context, st SettingStore) (*Manager, error) {
	m := &Manager{st: st}
	raw, err := st.GetSettingJSON(ctx, settingKey)
	if err != nil {
		return nil, err
	}
	if raw != "" {
		cfg := Default()
		if err := json.Unmarshal([]byte(raw), cfg); err == nil {
			// 旧版持久化配置可能缺少 sicbo/baccarat 键（反序列化后为 nil），补默认值
			cfg.applyDefaults()
			if err := cfg.Validate(); err == nil {
				m.cfg = cfg
				return m, nil
			}
			fmt.Printf("[casino] 数据库中的游戏配置校验失败(%v)，回退默认配置\n", err)
		} else {
			fmt.Printf("[casino] 数据库中的游戏配置解析失败(%v)，回退默认配置\n", err)
		}
	}
	def := Default()
	if err := m.persist(ctx, def); err != nil {
		return nil, err
	}
	m.cfg = def
	return m, nil
}

// Default 默认配置：各游戏理论回报（骰宝 ≈97.2%、
// 百家乐按规则文件从简后约 85%~90%、刮刮乐 ≈93.9%、大转盘/老虎机/21点 ≈95%）。
func Default() *Config {
	return &Config{
		MinBet:         0.01,
		MaxBet:         10,
		DailyLossLimit: 0,
		Wheel:          games.DefaultWheel(),
		Slots:          games.DefaultSlots(),
		Scratch:        games.DefaultScratch(),
		Blackjack:      games.DefaultBlackjack(),
		Sicbo:          games.DefaultSicbo(),
		Baccarat:       games.DefaultBaccarat(),
	}
}

// applyDefaults 补齐缺省的骰宝/百家乐/刮刮乐配置。
// 旧版持久化配置或管理端旧载荷缺少对应键时自动填入默认值；
// 旧版刮刮乐为 3×3 符号配置（无 win_rate/tiers 键），反序列化后为退化结构，
// 此处整体替换为「面值+数量」模型默认值；缺 lucky7/lines 键时仅补对应玩法。
// 归一化清理概率的浮点误差，使各概率表之和恰为 1（加载与保存共用本函数，
// 见 New/Save）。
func (c *Config) applyDefaults() {
	if c.Scratch == nil || c.Scratch.WinRate <= 0 || len(c.Scratch.Tiers) == 0 {
		c.Scratch = games.DefaultScratch()
	}
	if len(c.Scratch.Lucky7.Tiers) == 0 {
		c.Scratch.Lucky7 = games.DefaultLucky7()
	}
	if len(c.Scratch.Lines.Symbols) == 0 {
		c.Scratch.Lines = games.DefaultLines()
	}
	c.Scratch.Normalize()
	if c.Sicbo == nil {
		c.Sicbo = games.DefaultSicbo()
	}
	if c.Baccarat == nil {
		c.Baccarat = games.DefaultBaccarat()
	}
}

// Get 返回当前配置的拷贝（深拷贝避免外部修改影响运行态）。
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.cfg
}

// Save 校验新配置 → 写库 → 原子替换运行态。
// GetConfig 返回当前配置指针（只读；用于指针方法如 IsEnabled）。
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) Save(ctx context.Context, cfg *Config) error {
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := m.persist(ctx, cfg); err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	return nil
}

func (m *Manager) persist(ctx context.Context, cfg *Config) error {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return m.st.PutSettingJSON(ctx, settingKey, string(raw))
}

// Validate 配置合法性校验。
func (c *Config) Validate() error {
	if c.MinBet <= 0 {
		return fmt.Errorf("最小下注必须大于 0")
	}
	if c.MaxBet < c.MinBet {
		return fmt.Errorf("最大下注不能小于最小下注")
	}
	if len(c.Wheel.Segments) < 2 || len(c.Wheel.Segments) > 24 {
		return fmt.Errorf("转盘扇区数量须在 2~24 之间")
	}
	for _, s := range c.Wheel.Segments {
		if s.Weight < 0 || s.Multiplier < 0 {
			return fmt.Errorf("转盘扇区的权重与倍数不能为负")
		}
	}
	if len(c.Slots.Symbols) < 3 || len(c.Slots.Symbols) > 12 {
		return fmt.Errorf("老虎机符号数量须在 3~12 之间")
	}
	for _, s := range c.Slots.Symbols {
		if s.Weight <= 0 {
			return fmt.Errorf("老虎机符号 %s 的权重必须大于 0", s.ID)
		}
	}
	for id, p := range c.Slots.TriplePays {
		if p < 0 {
			return fmt.Errorf("老虎机符号 %s 三连赔率不能为负", id)
		}
	}
	for id, p := range c.Slots.PairPays {
		if p < 0 {
			return fmt.Errorf("老虎机符号 %s 两连赔率不能为负", id)
		}
	}
	if err := c.Scratch.Validate(); err != nil {
		return err
	}
	if c.Blackjack.BlackjackPays < 2 || c.Blackjack.BlackjackPays > 5 {
		return fmt.Errorf("21点 BlackJack 赔率须在 2~5 之间")
	}
	for _, b := range []struct {
		name string
		m    float64
	}{{"大", c.Sicbo.Big}, {"小", c.Sicbo.Small}, {"单", c.Sicbo.Odd}, {"双", c.Sicbo.Even}} {
		if b.m < 1 {
			return fmt.Errorf("骰宝「%s」注项赔率必须不小于 1", b.name)
		}
	}
	for _, b := range []struct {
		name string
		m    float64
	}{{"闲", c.Baccarat.Player}, {"庄", c.Baccarat.Banker}, {"和", c.Baccarat.Tie}} {
		if b.m < 1 {
			return fmt.Errorf("百家乐「%s」方向赔率必须不小于 1", b.name)
		}
	}
	if c.DailyLossLimit < 0 {
		return fmt.Errorf("日亏损上限不能为负")
	}
	return nil
}

// RTPSummary 各游戏的理论回报率（管理页实时预览）。
type RTPSummary struct {
	Wheel     float64            `json:"wheel"`
	Slots     float64            `json:"slots"`
	Scratch   map[string]float64 `json:"scratch"` // 刮刮乐各玩法回报率（classic/lucky7/lines）
	Blackjack string             `json:"blackjack"`
	Sicbo     map[string]float64 `json:"sicbo"`   // 各注项回报率
	Baccarat  map[string]float64 `json:"baccarat"` // 各下注方向回报率
}

// RTP 计算当前配置的理论回报率。
func (c *Config) RTP() RTPSummary {
	return RTPSummary{
		Wheel:     c.Wheel.RTP(),
		Slots:     c.Slots.RTP(),
		Scratch:   c.Scratch.RTP(),
		Blackjack: "≈95% (策略相关)",
		Sicbo:     c.Sicbo.RTP(),
		Baccarat:  c.Baccarat.RTP(),
	}
}
