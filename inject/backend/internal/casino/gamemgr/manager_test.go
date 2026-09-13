package gamemgr

import (
	"context"
	"encoding/json"
	"testing"
)

// fakeStore 返回预设的持久化 JSON（模拟旧版配置）。
type fakeStore struct {
	raw string
}

func (f fakeStore) GetSettingJSON(ctx context.Context, key string) (string, error) {
	return f.raw, nil
}
func (f fakeStore) PutSettingJSON(ctx context.Context, key, value string) error { return nil }

const legacyConfigJSON = `{"enabled":true,"min_bet":0.01,"max_bet":10,"daily_loss_limit":0,
"wheel":{"segments":[{"multiplier":2,"weight":1,"label":"2x"},{"multiplier":0,"weight":1,"label":"0x"}]},
"slots":{"symbols":[{"id":"a","emoji":"🍎","weight":1},{"id":"b","emoji":"🍌","weight":1},{"id":"c","emoji":"🍇","weight":1}],"triple_pays":{"a":3},"pair_pays":{"a":1}},
"blackjack":{"blackjack_pays":2.5,"dealer_stands_soft17":true,"double_allowed":true}}`

// legacyScratchJSON 携带旧版 3×3 符号刮刮乐配置的持久化载荷（无 win_rate/tiers 键）。
const legacyScratchJSON = `{"enabled":true,"min_bet":0.01,"max_bet":10,"daily_loss_limit":0,
"wheel":{"segments":[{"multiplier":2,"weight":1,"label":"2x"},{"multiplier":0,"weight":1,"label":"0x"}]},
"slots":{"symbols":[{"id":"a","emoji":"🍎","weight":1},{"id":"b","emoji":"🍌","weight":1},{"id":"c","emoji":"🍇","weight":1}],"triple_pays":{"a":3},"pair_pays":{"a":1}},
"blackjack":{"blackjack_pays":2.5,"dealer_stands_soft17":true,"double_allowed":true},
"scratch":{"symbols":[{"id":"cherry","weight":73,"prize":2},{"id":"lemon","weight":30,"prize":2}]}}`

// TestNewFillsLegacyConfig 旧版持久化配置缺少 sicbo/baccarat 键时应自动填默认值，
// 且整体校验通过、正常加载。
func TestNewFillsLegacyConfig(t *testing.T) {
	m, err := New(context.Background(), fakeStore{raw: legacyConfigJSON})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	cfg := m.Get()
	if cfg.Sicbo == nil {
		t.Fatal("旧配置缺 sicbo 键时应补默认值")
	}
	if cfg.Sicbo.Big != 2 || cfg.Sicbo.Small != 2 || cfg.Sicbo.Odd != 2 || cfg.Sicbo.Even != 2 {
		t.Fatalf("骰宝默认赔率应为 2/2/2/2, got %+v", cfg.Sicbo)
	}
	if cfg.Baccarat == nil {
		t.Fatal("旧配置缺 baccarat 键时应补默认值")
	}
	if cfg.Baccarat.Player != 2 || cfg.Baccarat.Banker != 1.95 || cfg.Baccarat.Tie != 9 {
		t.Fatalf("百家乐默认赔率应为 2/1.95/9, got %+v", cfg.Baccarat)
	}
	if cfg.Scratch == nil {
		t.Fatal("旧配置缺 scratch 键时应补默认值")
	}
	// 旧载荷无 scratch 键 → 补「面值+数量」默认；归一化后概率和恰为 1
	if cfg.Scratch.WinRate != 0.29 || len(cfg.Scratch.Tiers) != 7 {
		t.Fatalf("刮刮乐默认应为中奖率 0.29 + 7 档奖品, got %+v", cfg.Scratch)
	}
	var probSum float64
	for _, tier := range cfg.Scratch.Tiers {
		probSum += tier.Probability
	}
	if probSum != 1.0 {
		t.Fatalf("刮刮乐档位概率和应恰为 1, got %.20f", probSum)
	}
}

// TestValidateNewGames 新游戏赔率校验：低于 1 拒绝，缺省子配置补齐后通过。
func TestValidateNewGames(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("默认配置应通过校验: %v", err)
	}
	bad := Default()
	bad.Sicbo.Odd = 0.99
	if err := bad.Validate(); err == nil {
		t.Fatal("骰宝赔率 <1 应校验失败")
	}
	bad2 := Default()
	bad2.Baccarat.Tie = 0.5
	if err := bad2.Validate(); err == nil {
		t.Fatal("百家乐赔率 <1 应校验失败")
	}
	bad3 := Default()
	bad3.Scratch.WinRate = 1.5
	if err := bad3.Validate(); err == nil {
		t.Fatal("刮刮乐中奖率 >1 应校验失败")
	}
	// 旧载荷（无 sicbo/baccarat 键）经 applyDefaults 后应通过
	var legacy Config
	if err := json.Unmarshal([]byte(legacyConfigJSON), &legacy); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	legacy.applyDefaults()
	if err := legacy.Validate(); err != nil {
		t.Fatalf("补默认值后应通过校验: %v", err)
	}
	// 旧版 3×3 符号刮刮乐配置（无 win_rate/tiers 键）应在 applyDefaults 时
	// 整体替换为「面值+数量」模型默认值
	var oldScratch Config
	if err := json.Unmarshal([]byte(legacyScratchJSON), &oldScratch); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	oldScratch.applyDefaults()
	if oldScratch.Scratch.WinRate != 0.29 || len(oldScratch.Scratch.Tiers) != 7 {
		t.Fatalf("旧版符号刮刮乐应替换为新模型默认值, got %+v", oldScratch.Scratch)
	}
	if err := oldScratch.Validate(); err != nil {
		t.Fatalf("迁移后应通过校验: %v", err)
	}
}

// TestRTPSummaryContainsNewGames RTP 汇总应包含新游戏的各注项回报率。
func TestRTPSummaryContainsNewGames(t *testing.T) {
	cfg := Default()
	rtp := cfg.RTP()
	if len(rtp.Sicbo) != 4 {
		t.Fatalf("骰宝 RTP 应有 4 个注项, got %d", len(rtp.Sicbo))
	}
	if len(rtp.Baccarat) != 3 {
		t.Fatalf("百家乐 RTP 应有 3 个方向, got %d", len(rtp.Baccarat))
	}
	if len(rtp.Scratch) != 3 {
		t.Fatalf("刮刮乐 RTP 应含 3 种玩法, got %d (%v)", len(rtp.Scratch), rtp.Scratch)
	}
	for mode, v := range rtp.Scratch {
		if v <= 0 || v > 1 {
			t.Fatalf("刮刮乐玩法 %s 回报率异常: %v", mode, v)
		}
	}
	for side, v := range rtp.Baccarat {
		if v <= 0 || v > 3 {
			t.Fatalf("百家乐 %s 回报率异常: %v", side, v)
		}
	}
}
