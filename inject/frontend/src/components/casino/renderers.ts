/**
 * Casino scene renderers (ported from the legacy plugin games.js).
 *
 * Plain object/classic style on purpose: the canvas wheel, slot reels and
 * playing-card markup are DOM-driven, so the Vue stage components mount them
 * in onMounted and dispose/cancel in onUnmounted.
 */

import '@/assets/casino/game-scenes.css'
import bellSymbolUrl from '@/assets/casino/symbols/bell.png'
import sevenSymbolUrl from '@/assets/casino/symbols/seven.png'
import barSymbolUrl from '@/assets/casino/symbols/bar.png'
import grapesSymbolUrl from '@/assets/casino/symbols/grapes.png'
import watermelonSymbolUrl from '@/assets/casino/symbols/watermelon.png'
import cherrySymbolUrl from '@/assets/casino/symbols/cherry.png'
import diamondSymbolUrl from '@/assets/casino/symbols/diamond.png'
import starSymbolUrl from '@/assets/casino/symbols/star.png'
import plumSymbolUrl from '@/assets/casino/symbols/plum.png'
import horseshoeSymbolUrl from '@/assets/casino/symbols/horseshoe.png'
import lemonSymbolUrl from '@/assets/casino/symbols/lemon.png'
import type { CasinoBlackjackCard, CasinoBlackjackGame, CasinoWheelSegment } from '@/api/casino'

const TAU = Math.PI * 2

const DEFAULT_SEGMENTS: CasinoWheelSegment[] = [
  { label: '特等奖', multiplier: 10 }, { label: '一等奖', multiplier: 4 },
  { label: '二等奖', multiplier: 2 }, { label: '三等奖', multiplier: 1 },
  { label: '谢谢参与', multiplier: 0 }, { label: '四等奖', multiplier: 0.4 },
  { label: '五等奖', multiplier: 0.2 }, { label: '六等奖', multiplier: 0.1 },
  { label: '幸运奖', multiplier: 0.8 }, { label: '三等奖', multiplier: 1 },
  { label: '二等奖', multiplier: 2 }, { label: '一等奖', multiplier: 4 }
]

const PASTELS = ['#fff2c9', '#e9dcff', '#cbdcff', '#d8f1cf', '#d8ebfa', '#f9d6e8', '#fde8cf', '#eadbff']

export function casinoMoney(value: number | null | undefined): string {
  return Number(value || 0).toFixed(2)
}

function textOf(segment: CasinoWheelSegment | null | undefined): string {
  return String(segment && (segment.label || segment.multiplier + 'x') || '奖项')
}

function multiplierOf(segment: unknown): number {
  const value = Number((segment as { multiplier?: unknown } | null)?.multiplier)
  return Number.isFinite(value) ? value : 0
}

function pick(arr: string[]): string {
  return arr && arr.length ? arr[Math.floor(Math.random() * arr.length)] : ''
}

export interface WheelHubLabels {
  /** Main hub label, e.g. "开始" */
  label: string
  /** Cost template, rendered as "· {prefix} 10.00 ·" */
  costPrefix: string
}

const DEFAULT_HUB_LABELS: WheelHubLabels = { label: '开始', costPrefix: '消耗' }

/* ==================== Lucky Wheel (canvas) ==================== */

export const Wheel = {
  canvas: null as HTMLCanvasElement | null,
  hub: null as HTMLButtonElement | null,
  segments: DEFAULT_SEGMENTS.slice(),
  angle: 0,
  bet: 10,
  spinning: false,
  _raf: 0,
  _timer: 0,
  _token: 0,
  _container: null as HTMLElement | null,

  render(container: HTMLElement | null, segments?: CasinoWheelSegment[], hubLabels?: WheelHubLabels) {
    if (!container) return
    this.cleanup()
    this.segments = Array.isArray(segments) && segments.length ? segments : DEFAULT_SEGMENTS.slice()
    const labels = hubLabels || DEFAULT_HUB_LABELS
    container.classList.add('casino-scene-vars', 'casino-wheel-stage')
    container.innerHTML =
      '<div class="wheel-scene" role="img" aria-label="紫金灯饰幸运大转盘">' +
      '<canvas class="wheel-canvas" width="1280" height="1280"></canvas>' +
      '<div class="wheel-pointer" aria-hidden="true"></div>' +
      '<button type="button" class="wheel-hub" aria-label="' + labels.label + '"><span>' + labels.label + '</span>' +
      '<small>· ' + labels.costPrefix + ' <b class="wheel-bet-value">' + casinoMoney(this.bet) + '</b> ·</small></button></div>'
    const scene = container.querySelector('.wheel-scene')
    this.canvas = scene ? scene.querySelector('canvas') : null
    this.hub = scene ? scene.querySelector('.wheel-hub') : null
    this._container = container
    this.draw()
  },

  setBet(bet: number) {
    const value = Number(bet)
    if (Number.isFinite(value)) this.bet = value
    if (this.hub) {
      const valueNode = this.hub.querySelector('.wheel-bet-value')
      if (valueNode) valueNode.textContent = casinoMoney(this.bet)
    }
    this.draw()
  },

  draw() {
    const canvas = this.canvas
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    const scale = canvas.width / 640
    const cx = canvas.width / 2
    const cy = canvas.height / 2
    const rim = 286 * scale
    const radius = 266 * scale
    const list = this.segments.length ? this.segments : DEFAULT_SEGMENTS
    const step = TAU / list.length
    // Award names are ranked by multiplier; multiplier 0 renders as "谢谢参与".
    const tiers = [...new Set(list.map(multiplierOf).filter((v) => v > 0))].sort((a, b) => b - a)
    const TIER_NAMES = ['特等奖', '一等奖', '二等奖', '三等奖', '四等奖', '五等奖', '六等奖']
    const tierName = (v: number) => {
      const i = tiers.indexOf(v)
      return i < 0 ? '幸运奖' : TIER_NAMES[i] || i + 1 + ' 等奖'
    }
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.save()
    ctx.shadowColor = 'rgba(36, 15, 86, .35)'
    ctx.shadowBlur = 22 * scale
    ctx.shadowOffsetY = 7 * scale
    ctx.beginPath()
    ctx.arc(cx, cy, rim + 17 * scale, 0, TAU)
    ctx.fillStyle = '#f7d487'
    ctx.fill()
    ctx.restore()
    for (let i = 0; i < list.length; i++) {
      const a0 = this.angle + i * step - Math.PI / 2
      const a1 = a0 + step
      const gradient = ctx.createRadialGradient(cx - 20 * scale, cy - 22 * scale, 20 * scale, cx, cy, radius)
      const color = PASTELS[i % PASTELS.length]
      gradient.addColorStop(0, '#fffdf5')
      gradient.addColorStop(0.3, color)
      gradient.addColorStop(1, color)
      ctx.beginPath()
      ctx.moveTo(cx, cy)
      ctx.arc(cx, cy, radius, a0, a1)
      ctx.closePath()
      ctx.fillStyle = gradient
      ctx.fill()
      ctx.strokeStyle = 'rgba(137, 89, 181, .22)'
      ctx.lineWidth = 1.5 * scale
      ctx.stroke()
      const mid = a0 + step / 2
      const tx = cx + Math.cos(mid) * radius * 0.67
      const ty = cy + Math.sin(mid) * radius * 0.67
      const mult = multiplierOf(list[i])
      const name = mult === 0 ? textOf(list[i]) : tierName(mult)
      const sub = mult === 0 ? '' : casinoMoney(this.bet * mult)
      ctx.save()
      ctx.translate(tx, ty)
      ctx.textAlign = 'center'
      ctx.textBaseline = 'middle'
      ctx.fillStyle = '#5a277d'
      ctx.font = '700 ' + 16 * scale + 'px "Microsoft YaHei", sans-serif'
      ctx.fillText(name, 0, sub ? -11 * scale : 0)
      if (sub) {
        ctx.fillStyle = '#7343c6'
        ctx.font = '800 ' + 20 * scale + 'px system-ui, sans-serif'
        ctx.fillText(sub, 0, 14 * scale)
      }
      ctx.restore()
    }
    const metal = ctx.createLinearGradient(cx, cy - rim, cx, cy + rim)
    metal.addColorStop(0, '#fff2b5')
    metal.addColorStop(0.22, '#d89122')
    metal.addColorStop(0.48, '#fff0a2')
    metal.addColorStop(0.74, '#a45b11')
    metal.addColorStop(1, '#f5c64f')
    ;[rim + 4 * scale, rim + 11 * scale, rim + 20 * scale].forEach((r, n) => {
      ctx.beginPath()
      ctx.arc(cx, cy, r, 0, TAU)
      ctx.strokeStyle = n === 1 ? metal : 'rgba(96,38,129,.8)'
      ctx.lineWidth = (n === 1 ? 8 : 3) * scale
      ctx.stroke()
    })
    for (let i = 0; i < 30; i++) {
      const a = (i * TAU) / 30
      const x = cx + Math.cos(a) * (rim + 11 * scale)
      const y = cy + Math.sin(a) * (rim + 11 * scale)
      ctx.beginPath()
      ctx.arc(x, y, (i % 5 === 0 ? 5 : 3) * scale, 0, TAU)
      ctx.fillStyle = i % 5 === 0 ? '#fff6bf' : '#f6d86d'
      ctx.shadowColor = '#ffe48a'
      ctx.shadowBlur = (i % 5 === 0 ? 11 : 5) * scale
      ctx.fill()
      ctx.shadowBlur = 0
    }
  },

  /** Spin so that `index` (0-based) stops under the top pointer, then call done(). */
  spinTo(index: number, done?: () => void, quick?: boolean) {
    if (!this.canvas || this.spinning) return
    const n = this.segments.length || DEFAULT_SEGMENTS.length
    index = Math.max(0, Math.min(n - 1, Number(index) || 0))
    this.cleanupAnimationOnly()
    this.spinning = true
    const token = ++this._token
    const step = TAU / n
    const desired = -(index + 0.5) * step // segment center at the top pointer
    const current = this.angle
    let delta = desired - current
    delta = ((delta % TAU) + TAU) % TAU
    if (delta < TAU * 0.15) delta += TAU
    const target = current + delta + TAU * (quick ? 2 : 6)
    const duration = quick ? 850 : 4200
    const start = performance.now()
    const finish = () => {
      if (token !== this._token) return
      this.angle = target
      this.spinning = false
      this._raf = 0
      if (done) done()
    }
    const frame = (now: number) => {
      if (token !== this._token || !this.canvas || !this.canvas.isConnected) return
      const p = Math.min(1, (now - start) / duration)
      const e = 1 - Math.pow(1 - p, 4)
      this.angle = current + (target - current) * e
      this.draw()
      if (p < 1) this._raf = requestAnimationFrame(frame)
      else finish()
    }
    this._raf = requestAnimationFrame(frame)
    this._timer = window.setTimeout(finish, duration + 500)
  },

  cleanupAnimationOnly() {
    if (this._raf) cancelAnimationFrame(this._raf)
    if (this._timer) clearTimeout(this._timer)
    this._raf = this._timer = 0
    this._token++
    this.spinning = false
  },

  cleanup() {
    this.cleanupAnimationOnly()
    this.canvas = null
    this.hub = null
    this._container = null
  },

  dispose() {
    this.cleanup()
  }
}

/* ==================== Slot Machine (sprite reels) ==================== */

/* 透明底符号图标（由 reference-slots.jpg 抠图 + 生图补柠檬） */
const SYMBOL_URLS: Record<string, string> = {
  bell: bellSymbolUrl, seven: sevenSymbolUrl, bar: barSymbolUrl, grapes: grapesSymbolUrl,
  watermelon: watermelonSymbolUrl, cherry: cherrySymbolUrl, diamond: diamondSymbolUrl,
  star: starSymbolUrl, plum: plumSymbolUrl, horseshoe: horseshoeSymbolUrl, lemon: lemonSymbolUrl
}

const SPRITE_ALIAS: Record<string, string> = {
  '🍒': 'cherry', '🔔': 'bell', '7': 'seven', '💎': 'diamond', '⭐': 'star', '🍇': 'grapes',
  '🍉': 'watermelon', '🟣': 'plum', '🟪': 'plum', '🧲': 'horseshoe',
  bar: 'bar', BAR: 'bar', cherry: 'cherry', bell: 'bell', seven: 'seven', diamond: 'diamond',
  star: 'star', grapes: 'grapes', watermelon: 'watermelon', plum: 'plum', horseshoe: 'horseshoe'
}

function symbolKey(value: unknown): string {
  let raw: string
  if (typeof value === 'object' && value !== null) {
    const obj = value as { id?: string; key?: string; name?: string; symbol?: string; emoji?: string }
    raw = String(obj.id || obj.key || obj.name || obj.symbol || obj.emoji || '')
  } else {
    raw = String(value ?? '')
  }
  return SPRITE_ALIAS[raw] || SPRITE_ALIAS[raw.toLowerCase()] || 'lemon'
}

export const Slots = {
  symbols: [] as string[],
  _stopTimers: [] as number[],
  _running: false,
  ROLL_ROWS: 10,
  STOP_ROWS: 6,
  _container: null as HTMLElement | null,
  _strips: [] as HTMLElement[],

  render(container: HTMLElement | null, symbols?: string[]) {
    this.cancel()
    if (!container) return
    this.symbols = Array.isArray(symbols) ? symbols : []
    container.classList.add('casino-scene-vars', 'slot-grid5')
    // One vertical strip per column: scrolling is done with GPU transforms to avoid flicker.
    container.innerHTML =
      '<div class="slot-frame" aria-label="五列三行老虎机窗口">' +
      Array.from({ length: 5 }, (_, c) => '<div class="slot-col" data-col="' + c + '"><div class="slot-strip"></div></div>').join('') +
      '<div class="slot-payline" aria-hidden="true"></div></div>'
    this._container = container
    this._strips = Array.from(container.querySelectorAll('.slot-strip')) as HTMLElement[]
    this._fillStatic()
  },

  _setCell(cell: HTMLElement, value: string) {
    const key = symbolKey(value)
    cell.className = 'slot-cell'
    cell.dataset.symbol = key
    cell.textContent = ''
    cell.removeAttribute('style')
    const url = SYMBOL_URLS[key]
    if (url) cell.style.setProperty('--sym', `url('${url}')`)
    cell.setAttribute('aria-label', key)
  },

  /** rows: strip content; visibleRow: strip row aligned to the window top when stopped (win row = visibleRow + 1) */
  _buildStrip(strip: HTMLElement, rows: string[], visibleRow: number) {
    strip.style.setProperty('--rows', String(rows.length))
    strip.innerHTML = rows.map(() => '<div class="slot-cell"></div>').join('')
    Array.from(strip.children).forEach((cell, i) => this._setCell(cell as HTMLElement, rows[i]))
    this._showRow(strip, visibleRow)
  },

  _showRow(strip: HTMLElement, row: number) {
    strip.style.transition = 'none'
    // Row height is fixed to 1/3 of the window height regardless of row count.
    strip.style.transform = 'translateY(' + (-row * 100) / 3 + '%)'
    void strip.offsetHeight
  },

  _fillStatic() {
    const pool = this.symbols.length ? this.symbols : Object.keys(SYMBOL_URLS)
    this._strips.forEach((strip) => this._buildStrip(strip, [pick(pool), pick(pool), pick(pool)], 0))
  },

  start() {
    if (!this._container || !this._container.isConnected) return
    this.cancel()
    this._running = true
    this._container.classList.add('is-spinning')
    const pool = this.symbols.length ? this.symbols : Object.keys(SYMBOL_URLS)
    const n = this.ROLL_ROWS
    const half = n / 2
    this._strips.forEach((strip, col) => {
      const rows = Array.from({ length: n }, (_, i) => (i < half ? pick(pool) : ''))
      for (let i = half; i < n; i++) rows[i] = rows[i - half] // repeat head/tail for a seamless loop
      this._buildStrip(strip, rows, 0)
      strip.style.animation = 'slot-roll ' + (0.5 + col * 0.07) + 's linear infinite'
    })
  },

  /** Stop the reels on `emojis` (3 values, middle row is the payline), then call done(). */
  stop(emojis: string[], done?: () => void, quick?: boolean) {
    if (!this._container || !this._container.isConnected) return
    this._stopTimers.forEach((t) => clearTimeout(t))
    this._stopTimers = []
    const result = Array.isArray(emojis) ? emojis : []
    const pool = this.symbols.length ? this.symbols : Object.keys(SYMBOL_URLS)
    const delay = quick ? 120 : 380
    const slide = quick ? 0.28 : 0.62
    const win =
      result.length >= 3 &&
      symbolKey(result[0]) === symbolKey(result[1]) &&
      symbolKey(result[1]) === symbolKey(result[2])
    let last = 0
    this._strips.forEach((strip, col) => {
      const at = delay * col + 120
      last = Math.max(last, at)
      this._stopTimers.push(
        window.setTimeout(() => {
          const rows = [pick(pool), pick(pool), result[col] || pick(pool), pick(pool), pick(pool), pick(pool)]
          this._buildStrip(strip, rows, 3) // park two rows low, then slide up into place
          if (win) (strip.children[2] as HTMLElement).dataset.win = '1'
          strip.style.animation = 'none'
          this._showRow(strip, 3)
          strip.style.transition = 'transform ' + slide + 's cubic-bezier(.17,.84,.26,1.12)'
          requestAnimationFrame(() => {
            strip.style.transform = 'translateY(' + -1 * (100 / 3) + '%)'
          })
          this._stopTimers.push(
            window.setTimeout(() => {
              strip.style.transition = 'none'
            }, slide * 1000 + 60)
          )
        }, at)
      )
    })
    this._stopTimers.push(
      window.setTimeout(() => {
        this._running = false
        if (this._container) {
          this._container.classList.remove('is-spinning')
          this._container.classList.toggle('has-win', win)
        }
        if (done) done()
      }, last + slide * 1000 + 140)
    )
  },

  cancel() {
    this._stopTimers.forEach((t) => clearTimeout(t))
    this._stopTimers = []
    this._running = false
    if (this._container) {
      this._container.classList.remove('is-spinning')
      this._container.querySelectorAll('.slot-strip').forEach((s) => {
        const strip = s as HTMLElement
        strip.style.animation = 'none'
        strip.style.transition = 'none'
      })
    }
  }
}

/* ==================== Blackjack (playing cards) ==================== */

const SUITS: Record<string, string> = { s: '♠', h: '♥', d: '♦', c: '♣' }

export function handPts(cards: (CasinoBlackjackCard | null | undefined)[] | null | undefined): number {
  let total = 0
  let aces = 0
  ;(cards || []).forEach((c) => {
    const rank = c && (c.r || (c as unknown as { rank?: string }).rank)
    if (rank === 'A') {
      total += 11
      aces++
    } else if (['K', 'Q', 'J', '10'].includes(String(rank))) {
      total += 10
    } else {
      total += parseInt(String(rank), 10) || 0
    }
  })
  while (total > 21 && aces--) total -= 10
  return total
}

export interface BlackjackHandView {
  dealerCards: (CasinoBlackjackCard | null)[]
  dealerHidden: boolean
  playerCards: CasinoBlackjackCard[]
  playerTotal?: number
  playerSoft?: boolean
  labels?: { dealer?: string; you?: string; soft?: string }
}

export function blackjackCardHTML(card: CasinoBlackjackCard | null, hidden?: boolean): string {
  if (hidden || !card) return '<div class="bj-card bj-card-back" aria-label="暗牌"></div>'
  const suit = SUITS[card.s] || (card as unknown as { suit?: string }).suit || ''
  const red = card.s === 'h' || card.s === 'd' || suit === '♥' || suit === '♦'
  return (
    '<div class="bj-card' +
    (red ? ' is-red' : '') +
    '" aria-label="' +
    String(card.r || '') +
    suit +
    '"><span class="bj-rank">' +
    (card.r || '') +
    '</span><span class="bj-suit-small">' +
    suit +
    '</span><span class="bj-suit-large">' +
    suit +
    '</span></div>'
  )
}

export function renderHands(el: HTMLElement | null, st?: BlackjackHandView) {
  if (!el) return
  const state: BlackjackHandView = st || {
    dealerCards: [],
    dealerHidden: false,
    playerCards: []
  }
  const labels = { dealer: '庄家', you: '你', soft: '软', ...(state.labels || {}) }
  const dealerCards = Array.isArray(state.dealerCards) ? state.dealerCards : []
  const playerCards = Array.isArray(state.playerCards) ? state.playerCards : []
  const dealerPoints = state.dealerHidden
    ? dealerCards.length > 1
      ? handPts(dealerCards.slice(0, 1) as CasinoBlackjackCard[])
      : '—'
    : dealerCards.length
      ? handPts(dealerCards as CasinoBlackjackCard[])
      : '—'
  const playerPoints = playerCards.length
    ? String(state.playerTotal ?? handPts(playerCards)) + (state.playerSoft ? ' ' + labels.soft : '')
    : '—'
  const dealerHidden = state.dealerHidden
  const dealerHTML = dealerCards.length
    ? dealerCards.map((c, i) => blackjackCardHTML(c, dealerHidden && i === 1)).join('')
    : '' // 待机态不渲染占位暗牌，避免误读为已发牌
  const playerHTML = playerCards.length
    ? playerCards.map((c) => blackjackCardHTML(c, false)).join('')
    : '' // 同上
  el.innerHTML =
    '<section class="bj-dealer"><div class="bj-hand-label">' +
    labels.dealer +
    ' <b class="bj-points">' +
    dealerPoints +
    '</b></div><div class="bj-cards">' +
    dealerHTML +
    '</div></section><section class="bj-player"><div class="bj-cards">' +
    playerHTML +
    '</div><div class="bj-hand-label">' +
    labels.you +
    ' <b class="bj-points">' +
    playerPoints +
    '</b></div></section>'
}

/** Convenience: build the renderer view-model for a backend blackjack game. */
export function blackjackHandView(
  game: CasinoBlackjackGame | null,
  labels?: BlackjackHandView['labels']
): BlackjackHandView {
  const active = game?.status === 'active'
  const dealerCards: (CasinoBlackjackCard | null)[] = game ? [...game.dealer_cards] : []
  if (active && dealerCards.length === 1) dealerCards.push(null)
  return {
    dealerCards,
    dealerHidden: !!active,
    playerCards: game?.player_cards || [],
    playerTotal: game?.player_total,
    playerSoft: game?.player_soft,
    labels
  }
}

export const BJ = {
  cardHTML: blackjackCardHTML,
  renderHands,
  handPts
}

/** Resolve a wheel spin result to a segment index (accepts index or label). */
export function wheelSegmentIndex(segment: number | string, segments: CasinoWheelSegment[]): number {
  if (typeof segment === 'number' && Number.isFinite(segment)) return segment
  const label = String(segment)
  const byLabel = segments.findIndex((s) => s.label === label)
  return byLabel >= 0 ? byLabel : 0
}
