// Package casino 存储层：与主站共享 PostgreSQL（users 表），
// 娱乐场自有账本表 casino_bets / casino_blackjack_games / casino_settings。
// 余额变更全部使用单事务原子完成，并在流水表留痕。
package casino

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficient = errors.New("余额不足")
	ErrUserNotFound = errors.New("用户不存在或已被禁用")
	ErrActiveGame   = errors.New("已有进行中的21点牌局")
	ErrGameNotFound = errors.New("牌局不存在或已结束")
)

// UserInfo 用户信息（余额实时读库）。
type UserInfo struct {
	ID      int64   `json:"id"`
	Email   string  `json:"email"`
	Role    string  `json:"role"`
	Balance float64 `json:"balance"`
}

// BetRow 一条下注流水。
type BetRow struct {
	ID           int64   `json:"id"`
	Game         string  `json:"game"`
	Bet          float64 `json:"bet"`
	Payout       float64 `json:"payout"`
	Delta        float64 `json:"delta"`
	BalanceAfter float64 `json:"balance_after"`
	Detail       string  `json:"detail"` // JSON 字符串
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

// BlackjackGame 21点牌局持久化状态。
type BlackjackGame struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	Bet         float64 `json:"bet"`
	PlayerCards string  `json:"player_cards"` // JSON
	DealerCards string  `json:"dealer_cards"` // JSON
	Deck        string  `json:"deck"`         // JSON
	Doubled     bool    `json:"doubled"`
	Status      string  `json:"status"` // active / settled
	Result      string  `json:"result"` // win/lose/push/blackjack/bust
	Payout      float64 `json:"payout"`
}

// Store 数据库访问层。
type Store struct {
	pool *pgxpool.Pool
}

// NewStore 建立连接池。dsn 为主站 cfg.Database.DSN() 的 libpq keyword/value 格式，pgx 可直接解析。
func NewStore(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("解析数据库配置: %w", err)
	}
	cfg.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("连接数据库: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("数据库连通性检查失败: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close 关闭连接池。
func (s *Store) Close() { s.pool.Close() }

// EnsureSchema 创建娱乐场所需表结构（幂等，可安全重复执行）。
// 不触碰主站自身的任何表结构。
func (s *Store) EnsureSchema(ctx context.Context) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS casino_bets (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL,
    game          VARCHAR(20) NOT NULL,
    bet           DECIMAL(20,8) NOT NULL,
    payout        DECIMAL(20,8) NOT NULL DEFAULT 0,
    delta         DECIMAL(20,8) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    detail        JSONB NOT NULL DEFAULT '{}',
    status        VARCHAR(20) NOT NULL DEFAULT 'settled',
    game_ref      BIGINT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_casino_bets_user_time ON casino_bets (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_casino_bets_game_time ON casino_bets (game, created_at DESC);

CREATE TABLE IF NOT EXISTS casino_blackjack_games (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL,
    bet          DECIMAL(20,8) NOT NULL,
    player_cards JSONB NOT NULL,
    dealer_cards JSONB NOT NULL,
    deck         JSONB NOT NULL,
    doubled      BOOLEAN NOT NULL DEFAULT FALSE,
    status       VARCHAR(20) NOT NULL DEFAULT 'active',
    result       VARCHAR(20),
    payout       DECIMAL(20,8),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_casino_bj_active
    ON casino_blackjack_games (user_id) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS casino_settings (
    key        VARCHAR(64) PRIMARY KEY,
    value      JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
	_, err := s.pool.Exec(ctx, ddl)
	return err
}

// round8 金额统一保留 8 位小数，避免浮点尾差写入 DECIMAL(20,8)。
func round8(v float64) float64 { return math.Round(v*1e8) / 1e8 }

// GetUserByID 按 id 取活跃用户（主站中间件已完成鉴权，这里只补全邮箱/角色/余额）。
func (s *Store) GetUserByID(ctx context.Context, id int64) (*UserInfo, error) {
	u := &UserInfo{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, role, balance FROM users
		 WHERE id = $1 AND deleted_at IS NULL AND status = 'active'`,
		id,
	).Scan(&u.ID, &u.Email, &u.Role, &u.Balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

// checkDailyLoss 事务内执行：当日累计净输超过限额时拒绝下注（limit<=0 表示不限）。
func checkDailyLoss(ctx context.Context, tx pgx.Tx, userID int64, limit float64) error {
	if limit <= 0 {
		return nil
	}
	var net float64
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(delta), 0) FROM casino_bets
		 WHERE user_id = $1 AND status <> 'pending' AND created_at >= date_trunc('day', NOW())`,
		userID).Scan(&net)
	if err != nil {
		return err
	}
	if -net >= limit {
		return fmt.Errorf("已达今日亏损上限 (%.2f)，今日无法继续下注", limit)
	}
	return nil
}

// SettleInstantBet 即时类游戏（转盘/老虎机/天牌21点）原子结算：
// 校验余额 → 扣注+派奖（单条 UPDATE 原子完成）→ 写流水。返回结算后余额。
func (s *Store) SettleInstantBet(ctx context.Context, userID int64, bet, payout float64, game, detail string, dailyLossLimit float64) (float64, error) {
	bet, payout = round8(bet), round8(payout)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	if err := checkDailyLoss(ctx, tx, userID, dailyLossLimit); err != nil {
		return 0, err
	}
	var balance float64
	err = tx.QueryRow(ctx,
		`UPDATE users SET balance = balance - $1 + $2, updated_at = NOW()
		 WHERE id = $3 AND deleted_at IS NULL AND status = 'active' AND balance >= $1
		 RETURNING balance`,
		bet, payout, userID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		// 区分用户不可用与余额不足
		var n int
		_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE id=$1 AND deleted_at IS NULL AND status='active'`, userID).Scan(&n)
		if n == 0 {
			return 0, ErrUserNotFound
		}
		return 0, ErrInsufficient
	}
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO casino_bets (user_id, game, bet, payout, delta, balance_after, detail)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`,
		userID, game, bet, payout, round8(payout-bet), balance, detail); err != nil {
		return 0, err
	}
	return balance, tx.Commit(ctx)
}

// DealBlackjack 开局：原子扣注并创建牌局。每用户同时仅允许一局（部分唯一索引约束）。
func (s *Store) DealBlackjack(ctx context.Context, userID int64, bet float64, playerCards, dealerCards, deck, detail string, dailyLossLimit float64) (int64, float64, error) {
	bet = round8(bet)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	if err := checkDailyLoss(ctx, tx, userID, dailyLossLimit); err != nil {
		return 0, 0, err
	}
	var balance float64
	err = tx.QueryRow(ctx,
		`UPDATE users SET balance = balance - $1, updated_at = NOW()
		 WHERE id = $2 AND deleted_at IS NULL AND status = 'active' AND balance >= $1
		 RETURNING balance`,
		bet, userID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		var n int
		_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE id=$1 AND deleted_at IS NULL AND status='active'`, userID).Scan(&n)
		if n == 0 {
			return 0, 0, ErrUserNotFound
		}
		return 0, 0, ErrInsufficient
	}
	if err != nil {
		return 0, 0, err
	}
	var gameID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO casino_blackjack_games (user_id, bet, player_cards, dealer_cards, deck)
		 VALUES ($1, $2, $3::jsonb, $4::jsonb, $5::jsonb) RETURNING id`,
		userID, bet, playerCards, dealerCards, deck).Scan(&gameID)
	if err != nil {
		if strings.Contains(err.Error(), "uq_casino_bj_active") {
			return 0, 0, ErrActiveGame
		}
		return 0, 0, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO casino_bets (user_id, game, bet, delta, balance_after, detail, status, game_ref)
		 VALUES ($1, 'blackjack', $2, $3, $4, $5::jsonb, 'pending', $6)`,
		userID, bet, round8(-bet), balance, detail, gameID); err != nil {
		return 0, 0, err
	}
	return gameID, balance, tx.Commit(ctx)
}

// GetActiveBlackjack 取用户当前进行中的牌局。
func (s *Store) GetActiveBlackjack(ctx context.Context, userID int64) (*BlackjackGame, error) {
	return s.scanBlackjack(s.pool.QueryRow(ctx,
		`SELECT id, user_id, bet, player_cards::text, dealer_cards::text, deck::text, doubled, status, COALESCE(result,''), COALESCE(payout,0)
		 FROM casino_blackjack_games WHERE user_id = $1 AND status = 'active'`, userID))
}

// GetBlackjack 按 id 取牌局（限本人）。
func (s *Store) GetBlackjack(ctx context.Context, id, userID int64) (*BlackjackGame, error) {
	g, err := s.scanBlackjack(s.pool.QueryRow(ctx,
		`SELECT id, user_id, bet, player_cards::text, dealer_cards::text, deck::text, doubled, status, COALESCE(result,''), COALESCE(payout,0)
		 FROM casino_blackjack_games WHERE id = $1 AND user_id = $2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGameNotFound
	}
	return g, err
}

func (s *Store) scanBlackjack(row pgx.Row) (*BlackjackGame, error) {
	g := &BlackjackGame{}
	err := row.Scan(&g.ID, &g.UserID, &g.Bet, &g.PlayerCards, &g.DealerCards, &g.Deck, &g.Doubled, &g.Status, &g.Result, &g.Payout)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGameNotFound
	}
	return g, err
}

// SettleBlackjack 结算 21点：extraBet 用于加倍补扣（在原注基础上再扣一倍），
// payout 为返还总额（含本金）。事务内锁定牌局行防止并发操作，完成后更新对应 pending 流水。
func (s *Store) SettleBlackjack(ctx context.Context, gameID, userID int64, extraBet, payout float64, doubled bool, result, playerCards, dealerCards, deck, detail string) (float64, error) {
	extraBet, payout = round8(extraBet), round8(payout)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// 锁定并校验牌局仍为 active，同时取原始注额用于累计
	var status string
	var origBet float64
	err = tx.QueryRow(ctx,
		`SELECT status, bet FROM casino_blackjack_games WHERE id = $1 AND user_id = $2 FOR UPDATE`,
		gameID, userID).Scan(&status, &origBet)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrGameNotFound
	}
	if err != nil {
		return 0, err
	}
	if status != "active" {
		return 0, ErrGameNotFound
	}

	balance := 0.0
	if extraBet > 0 { // 加倍补扣
		err = tx.QueryRow(ctx,
			`UPDATE users SET balance = balance - $1, updated_at = NOW()
			 WHERE id = $2 AND deleted_at IS NULL AND status='active' AND balance >= $1
			 RETURNING balance`,
			extraBet, userID).Scan(&balance)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInsufficient
		}
		if err != nil {
			return 0, err
		}
	}
	if payout > 0 {
		err = tx.QueryRow(ctx,
			`UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2 RETURNING balance`,
			payout, userID).Scan(&balance)
		if err != nil {
			return 0, err
		}
	}
	if _, err := tx.Exec(ctx,
		`UPDATE casino_blackjack_games
		 SET status='settled', result=$1, payout=$2, doubled=$3, player_cards=$4::jsonb, dealer_cards=$5::jsonb, deck=$6::jsonb, updated_at=NOW()
		 WHERE id=$7`,
		result, payout, doubled, playerCards, dealerCards, deck, gameID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE casino_bets
		 SET bet = $1, payout = $2, delta = $3, balance_after = $4, status='settled', detail = $5::jsonb
		 WHERE game_ref = $6 AND status = 'pending'`,
		round8(origBet+extraBet), payout, round8(payout-origBet-extraBet), balance, detail, gameID); err != nil {
		return 0, err
	}
	return balance, tx.Commit(ctx)
}

// UpdateBlackjackHands 要牌未结束时保存手牌与剩余牌堆（仅 active 局）。
func (s *Store) UpdateBlackjackHands(ctx context.Context, gameID int64, playerCards, deck string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE casino_blackjack_games SET player_cards=$1::jsonb, deck=$2::jsonb, updated_at=NOW()
		 WHERE id=$3 AND status='active'`,
		playerCards, deck, gameID)
	return err
}

// UserTodayStats 用户当天已结算流水统计，不受历史列表分页限制。
type UserTodayStats struct {
	Profit float64 `json:"today_profit"`
	Rounds int64   `json:"today_rounds"`
}

// TodayStats 按 Asia/Shanghai 的完整自然日统计用户已结算账本。
func (s *Store) TodayStats(ctx context.Context, userID int64) (*UserTodayStats, error) {
	st := &UserTodayStats{}
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(delta), 0), COUNT(*)
		 FROM casino_bets
		 WHERE user_id = $1 AND status = 'settled'
		   AND created_at >= ((NOW() AT TIME ZONE 'Asia/Shanghai')::date AT TIME ZONE 'Asia/Shanghai')
		   AND created_at < (((NOW() AT TIME ZONE 'Asia/Shanghai')::date + INTERVAL '1 day') AT TIME ZONE 'Asia/Shanghai')`,
		userID).Scan(&st.Profit, &st.Rounds)
	if err != nil {
		return nil, err
	}
	st.Profit = round8(st.Profit)
	return st, nil
}

// ListHistory 分页取用户流水。
func (s *Store) ListHistory(ctx context.Context, userID int64, limit, offset int) ([]BetRow, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, game, bet, payout, delta, balance_after, detail::text, status,
		        to_char(created_at AT TIME ZONE 'Asia/Shanghai', 'YYYY-MM-DD HH24:MI:SS')
		 FROM casino_bets WHERE user_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BetRow
	for rows.Next() {
		var r BetRow
		if err := rows.Scan(&r.ID, &r.Game, &r.Bet, &r.Payout, &r.Delta, &r.BalanceAfter, &r.Detail, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ExpiredBlackjack 结算前先读取一张超时牌局（每次处理一张，由后台定时任务循环调用）。
func (s *Store) ExpiredBlackjack(ctx context.Context, olderThanHours int) (*BlackjackGame, error) {
	g, err := s.scanBlackjack(s.pool.QueryRow(ctx,
		`SELECT id, user_id, bet, player_cards::text, dealer_cards::text, deck::text, doubled, status, COALESCE(result,''), COALESCE(payout,0)
		 FROM casino_blackjack_games
		 WHERE status='active' AND updated_at < NOW() - make_interval(hours => $1)
		 ORDER BY id LIMIT 1`, olderThanHours))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return g, err
}

// ---- 配置存储（gamemgr.SettingStore 接口实现）----

// GetSettingJSON 读取 casino_settings 中 key 对应的 JSON 字符串，不存在返回 ("", nil)。
func (s *Store) GetSettingJSON(ctx context.Context, key string) (string, error) {
	var v string
	err := s.pool.QueryRow(ctx, `SELECT value::text FROM casino_settings WHERE key = $1`, key).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return v, err
}

// PutSettingJSON 写入（upsert）游戏配置。
func (s *Store) PutSettingJSON(ctx context.Context, key, value string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO casino_settings (key, value) VALUES ($1, $2::jsonb)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		key, value)
	return err
}

// ---- 管理端：统计 ----

// StatsSummary 管理看板统计。
type StatsSummary struct {
	TotalBets   int64   `json:"total_bets"`
	TotalBet    float64 `json:"total_bet"`    // 总下注额
	TotalPayout float64 `json:"total_payout"` // 总派奖额
	Profit      float64 `json:"profit"`       // 平台盈余 = 总下注 - 总派奖
	TodayBets   int64   `json:"today_bets"`
	TodayProfit float64 `json:"today_profit"`
	Players     int64   `json:"players"`
}

// LeaderboardEntry 赢家榜一行。
type LeaderboardEntry struct {
	Rank   int     `json:"rank"`
	Player string  `json:"player"` // 打码邮箱
	Profit float64 `json:"profit"`
}

// Leaderboard 最近 7 天盈利榜前 5（只统计已结算流水，邮箱打码）。
func (s *Store) Leaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT u.email, COALESCE(SUM(b.delta), 0) AS profit
		 FROM casino_bets b JOIN users u ON u.id = b.user_id
		 WHERE b.status = 'settled' AND b.created_at >= NOW() - interval '7 days'
		 GROUP BY b.user_id, u.email
		 HAVING SUM(b.delta) > 0
		 ORDER BY profit DESC
		 LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var email string
		var profit float64
		if err := rows.Scan(&email, &profit); err != nil {
			return nil, err
		}
		// 邮箱打码：保留前 2 位 + *** + 域名
		masked := email
		if at := strings.Index(email, "@"); at > 2 {
			masked = email[:2] + "***" + email[at:]
		} else if len(email) > 2 {
			masked = email[:2] + "***"
		}
		out = append(out, LeaderboardEntry{Rank: rank, Player: masked, Profit: round8(profit)})
		rank++
	}
	return out, rows.Err()
}

// AdminStats 汇总平台盈亏。
func (s *Store) AdminStats(ctx context.Context) (*StatsSummary, error) {
	st := &StatsSummary{}
	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(COUNT(*),0), COALESCE(SUM(bet),0), COALESCE(SUM(payout),0),
		        COALESCE(SUM(bet)-SUM(payout),0), COALESCE(COUNT(DISTINCT user_id),0)
		 FROM casino_bets WHERE status <> 'pending'`).Scan(
		&st.TotalBets, &st.TotalBet, &st.TotalPayout, &st.Profit, &st.Players)
	if err != nil {
		return nil, err
	}
	err = s.pool.QueryRow(ctx,
		`SELECT COALESCE(COUNT(*),0), COALESCE(SUM(bet)-SUM(payout),0)
		 FROM casino_bets WHERE status <> 'pending' AND created_at >= date_trunc('day', NOW())`).
		Scan(&st.TodayBets, &st.TodayProfit)
	return st, err
}

// AdminRecentBets 管理端最近下注（含用户邮箱）。
func (s *Store) AdminRecentBets(ctx context.Context, limit int) ([]map[string]any, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT b.id, u.email, b.game, b.bet, b.payout, b.delta, b.status,
		        to_char(b.created_at AT TIME ZONE 'Asia/Shanghai', 'MM-DD HH24:MI:SS')
		 FROM casino_bets b JOIN users u ON u.id = b.user_id
		 ORDER BY b.id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var email, game, status, created string
		var bet, payout, delta float64
		if err := rows.Scan(&id, &email, &game, &bet, &payout, &delta, &status, &created); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id": id, "email": email, "game": game, "bet": bet,
			"payout": payout, "delta": delta, "status": status, "created_at": created,
		})
	}
	return out, rows.Err()
}

// GameStats 分游戏统计（管理端配置页参考）。
func (s *Store) GameStats(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT game, COUNT(*), COALESCE(SUM(bet),0), COALESCE(SUM(payout),0),
		        CASE WHEN SUM(bet) > 0 THEN (SUM(bet)-SUM(payout))/SUM(bet) ELSE 0 END
		 FROM casino_bets WHERE status <> 'pending' GROUP BY game ORDER BY game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var game string
		var n int64
		var bet, payout, margin float64
		if err := rows.Scan(&game, &n, &bet, &payout, &margin); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"game": game, "rounds": n, "bet": bet, "payout": payout, "house_margin": margin,
		})
	}
	return out, rows.Err()
}
