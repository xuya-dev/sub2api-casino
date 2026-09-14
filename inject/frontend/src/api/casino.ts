/**
 * Casino API endpoints
 * Entertainment module (wheel / slots / blackjack) backed by the main-site
 * axios client — the response interceptor unwraps the { code, message, data }
 * envelope, so business functions return `data` directly.
 */

import { apiClient } from './client'

// ==================== Types ====================

export interface CasinoWheelSegment {
  label: string
  multiplier: number
  /** 出现权重（仅管理端 /admin/config 返回） */
  weight?: number
}

export interface CasinoSlotSymbol {
  id: string
  emoji: string
  /** 出现权重（仅管理端 /admin/config 返回） */
  weight?: number
}

export interface CasinoBlackjackConfig {
  /** Blackjack payout multiplier (includes stake), e.g. 2.5 => 3:2 */
  blackjack_pays: number
  dealer_stands_soft17: boolean
  double_allowed: boolean
}

export interface CasinoMeta {
  min_bet: number
  max_bet: number
  daily_loss_limit: number
  wheel: { segments: CasinoWheelSegment[] }
  slots: {
    symbols: CasinoSlotSymbol[]
    triple_pays?: Record<string, number>
    pair_pays?: Record<string, number>
  }
  blackjack: CasinoBlackjackConfig
}

export interface CasinoUser {
  id: number
  email: string
  role: string
  balance: number
}

export interface CasinoMe {
  user: CasinoUser
  /** Today's net profit (may be negative) */
  today_profit: number
  today_rounds: number
}

export interface CasinoBetRow {
  id: number
  /** 'wheel' | 'slots' | 'blackjack' */
  game: string
  bet: number
  payout: number
  /** Net delta applied to balance (payout - bet) */
  delta: number
  balance_after: number
  status: string
  created_at: string
}

export interface CasinoLeaderboardRow {
  rank: number
  player: string
  profit: number
}

export interface CasinoWheelSpinResult {
  /** Index of the winning segment (or its label, depending on backend build) */
  segment: number | string
  multiplier: number
  payout: number
  bet: number
  balance: number
}

export interface CasinoSlotsSpinResult {
  reels: string[]
  reel_ids: string[]
  multiplier: number
  payout: number
  bet: number
  balance: number
}

export interface CasinoBlackjackCard {
  /** Rank: A, 2..10, J, Q, K */
  r: string
  /** Suit: s / h / d / c */
  s: string
}

export type CasinoSicboBetType = 'big' | 'small' | 'odd' | 'even'

export interface CasinoSicboRollResult {
  /** Three dice values (1-6 each) */
  dice: number[]
  /** Sum of the three dice */
  sum: number
  /** 'big' | 'small' | 'odd' | 'even' | 'triple' (triple wipes all above bets) */
  result: CasinoSicboBetType | 'triple'
  multiplier: number
  payout: number
  bet: number
  balance: number
}

export type CasinoScratchMode = 'classic' | 'lucky7' | 'lines'

/** 一格涂层刮开后的奖层内容 */
export interface CasinoScratchCell {
  /** mult: 倍数格 / seven: 幸运7格 / symbol: 连线符号格 / dud: 未中奖格 */
  kind: 'mult' | 'seven' | 'symbol' | 'dud'
  /** symbol 格的符号 id（bell/cherry/crown/diamond/grape/lemon/melon/seven） */
  symbol?: string
  /** 该格派彩倍数（0 时省略） */
  multiplier?: number
}

export interface CasinoScratchCardResult {
  multiplier: number
  prize: number
  /** 涂层格：classic 1 格 / lucky7 7 格 / lines 9 格（行优先） */
  cells: CasinoScratchCell[]
  /** 连线玩法的中奖线（每项为 3 个行优先下标） */
  win_lines?: number[][]
}

export interface CasinoScratchRevealResult {
  mode: CasinoScratchMode
  face_value: number
  count: number
  cards: CasinoScratchCardResult[]
  win_count: number
  total_payout: number
  total_bet: number
  balance: number
}

/**
 * Buy & reveal scratch cards in bulk: face_value × count cards are settled
 * atomically; the client plays the reveal animation card by card while the
 * user (or the auto mode) uncovers the foil. mode: classic / lucky7 / lines.
 */
export async function scratchReveal(
  faceValue: number,
  count: number,
  mode: CasinoScratchMode = 'classic'
): Promise<CasinoScratchRevealResult> {
  const { data } = await apiClient.post<CasinoScratchRevealResult>('/casino/games/scratch/reveal', {
    face_value: faceValue,
    count,
    mode
  })
  return data
}

export type CasinoBaccaratSide = 'player' | 'banker' | 'tie'

export interface CasinoBaccaratDealResult {
  player_cards: CasinoBlackjackCard[]
  banker_cards: CasinoBlackjackCard[]
  /** Points modulo 10 */
  player_points: number
  banker_points: number
  outcome: CasinoBaccaratSide
  multiplier: number
  payout: number
  bet: number
  balance: number
}

export interface CasinoBlackjackGame {
  game_id: string
  bet: number
  doubled: boolean
  player_cards: CasinoBlackjackCard[]
  dealer_cards: CasinoBlackjackCard[]
  player_total: number
  player_soft: boolean
  status: 'active' | 'settled'
  /** 'win' | 'blackjack' | 'push' | 'bust' | 'lose' | null when still active */
  result?: string | null
  payout: number
  can_double: boolean
  balance?: number
}

export interface CasinoAdminConfig {
  /** 娱乐模式总开关（关闭时用户侧隐藏入口并停止服务） */
  enabled?: boolean
  min_bet: number
  max_bet: number
  daily_loss_limit: number
  wheel: { segments: CasinoWheelSegment[] }
  slots: {
    symbols: CasinoSlotSymbol[]
    triple_pays?: Record<string, number>
    pair_pays?: Record<string, number>
  }
  blackjack: CasinoBlackjackConfig
  sicbo?: { big: number; small: number; odd: number; even: number }
  baccarat?: { player: number; banker: number; tie: number }
  scratch?: {
    win_rate: number
    tiers: { multiplier: number; probability: number }[]
    lucky7?: { cells: number; hit_rate: number }
    lines?: { win_rate: number; symbols: { id: string; multiplier: number; probability: number }[] }
  }
}

export interface CasinoAdminStats {
  summary: Record<string, unknown>
  games: Record<string, unknown>[]
  recent: Record<string, unknown>[]
}

// ==================== Endpoints ====================

/**
 * Game metadata: bet limits, wheel segments, slot symbols, blackjack rules.
 */
export async function getMeta(): Promise<CasinoMeta> {
  const { data } = await apiClient.get<CasinoMeta>('/casino/meta')
  return data
}

/**
 * Casino view of the current user: balance + today's profit/rounds.
 */
export async function getMe(): Promise<CasinoMe> {
  const { data } = await apiClient.get<CasinoMe>('/casino/me')
  return data
}

/**
 * Recent bet history (newest first).
 */
export async function getHistory(limit = 20): Promise<{ items: CasinoBetRow[] }> {
  const { data } = await apiClient.get<{ items: CasinoBetRow[] }>('/casino/history', {
    params: { limit }
  })
  return data
}

/**
 * 7-day profit leaderboard.
 */
export async function getLeaderboard(): Promise<{ items: CasinoLeaderboardRow[] }> {
  const { data } = await apiClient.get<{ items: CasinoLeaderboardRow[] }>('/casino/leaderboard')
  return data
}

/**
 * Spin the lucky wheel.
 */
export async function spinWheel(bet: number): Promise<CasinoWheelSpinResult> {
  const { data } = await apiClient.post<CasinoWheelSpinResult>('/casino/games/wheel/spin', { bet })
  return data
}

/**
 * Spin the slot machine.
 */
export async function spinSlots(bet: number): Promise<CasinoSlotsSpinResult> {
  const { data } = await apiClient.post<CasinoSlotsSpinResult>('/casino/games/slots/spin', { bet })
  return data
}

/**
 * Deal a new blackjack hand.
 */
export async function blackjackDeal(bet: number): Promise<{ game: CasinoBlackjackGame }> {
  const { data } = await apiClient.post<{ game: CasinoBlackjackGame }>('/casino/blackjack/deal', { bet })
  return data
}

/**
 * Ask for another card.
 */
export async function blackjackHit(gameId: string): Promise<{ game: CasinoBlackjackGame }> {
  const { data } = await apiClient.post<{ game: CasinoBlackjackGame }>('/casino/blackjack/hit', { game_id: gameId })
  return data
}

/**
 * Stand on the current hand.
 */
export async function blackjackStand(gameId: string): Promise<{ game: CasinoBlackjackGame }> {
  const { data } = await apiClient.post<{ game: CasinoBlackjackGame }>('/casino/blackjack/stand', { game_id: gameId })
  return data
}

/**
 * Double down (one extra card, then stand).
 */
export async function blackjackDouble(gameId: string): Promise<{ game: CasinoBlackjackGame }> {
  const { data } = await apiClient.post<{ game: CasinoBlackjackGame }>('/casino/blackjack/double', { game_id: gameId })
  return data
}

/**
 * Restore an in-flight blackjack hand after a page refresh.
 */
export async function blackjackCurrent(): Promise<{ game: CasinoBlackjackGame | null }> {
  const { data } = await apiClient.get<{ game: CasinoBlackjackGame | null }>('/casino/blackjack/current')
  return data
}

/**
 * Roll the sic bo dice (bet on big / small / odd / even).
 */
export async function sicboRoll(bet: number, betType: CasinoSicboBetType): Promise<CasinoSicboRollResult> {
  const { data } = await apiClient.post<CasinoSicboRollResult>('/casino/games/sicbo/roll', { bet, bet_type: betType })
  return data
}

/**
 * Deal a baccarat round (bet on player / banker / tie).
 */
export async function baccaratDeal(bet: number, side: CasinoBaccaratSide): Promise<CasinoBaccaratDealResult> {
  const { data } = await apiClient.post<CasinoBaccaratDealResult>('/casino/games/baccarat/deal', { bet, side })
  return data
}

// ==================== 状态 ====================

/** 娱乐模式是否开启（公共接口，供侧边栏菜单使用） */
export async function getCasinoStatus(): Promise<{ enabled: boolean }> {
  const { data } = await apiClient.get<{ enabled: boolean }>('/casino/status')
  return data
}

// ==================== Admin (placeholder for a future admin page) ====================

/**
 * Admin: read casino configuration.
 */
export async function getAdminConfig(): Promise<{ config: CasinoAdminConfig; rtp: Record<string, unknown> }> {
  const { data } = await apiClient.get<{ config: CasinoAdminConfig; rtp: Record<string, unknown> }>(
    '/casino/admin/config'
  )
  return data
}

/**
 * Admin: update casino configuration.
 */
export async function updateAdminConfig(config: Partial<CasinoAdminConfig>): Promise<{ config: CasinoAdminConfig }> {
  const { data } = await apiClient.put<{ config: CasinoAdminConfig }>('/casino/admin/config', config)
  return data
}

/**
 * Admin: aggregate profit stats.
 */
export async function getAdminStats(): Promise<CasinoAdminStats> {
  const { data } = await apiClient.get<CasinoAdminStats>('/casino/admin/stats')
  return data
}

export const casinoAPI = {
  getCasinoStatus,
  getMeta,
  getMe,
  getHistory,
  getLeaderboard,
  spinWheel,
  spinSlots,
  blackjackDeal,
  blackjackHit,
  blackjackStand,
  blackjackDouble,
  blackjackCurrent,
  sicboRoll,
  baccaratDeal,
  scratchReveal,
  getAdminConfig,
  updateAdminConfig,
  getAdminStats
}

export default casinoAPI
