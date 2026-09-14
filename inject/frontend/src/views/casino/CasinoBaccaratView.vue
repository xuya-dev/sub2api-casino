<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('casino.baccarat.title') }}
            <span class="ml-1 align-middle text-xl" aria-hidden="true">🎴</span>
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.baccarat.description') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <router-link to="/casino" class="btn btn-secondary">
            {{ t('casino.backToLobby') }}
          </router-link>
          <button type="button" class="btn btn-secondary" @click="showRules = !showRules">
            <Icon name="questionCircle" size="md" />
            {{ t('casino.common.infoTitle') }}
          </button>
        </div>
      </div>

      <!-- Stats bar -->
      <CasinoStatsBar :items="statItems">
        <router-link to="/casino/history" class="btn btn-secondary">
          <Icon name="document" size="sm" />
          {{ t('casino.stats.gameRecords') }}
          <Icon name="chevronDown" size="xs" />
        </router-link>
      </CasinoStatsBar>

      <!-- Rules -->
      <transition name="fade">
        <div v-if="showRules" class="card p-6">
          <ul class="list-inside list-disc space-y-1 text-sm text-gray-600 dark:text-dark-300">
            <li>{{ t('casino.baccarat.ruleOdds') }}</li>
            <li>{{ t('casino.baccarat.rulePoints') }}</li>
            <li>{{ t('casino.baccarat.ruleDeal') }}</li>
            <li>{{ t('casino.common.betRange', { min: formatBet(meta?.min_bet ?? 0), max: formatBet(meta?.max_bet ?? 0) }) }}</li>
          </ul>
        </div>
      </transition>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div class="flex min-w-0 flex-col gap-6">
          <!-- Table -->
          <div class="bac-table">
            <!-- Player hand -->
            <section class="bac-hand">
              <h2 class="bac-hand-title player">{{ t('casino.baccarat.player') }}</h2>
              <div class="bac-cards">
                <template v-if="!result">
                  <div v-for="i in 2" :key="'p-back-' + i" class="bac-card-slot" v-html="BJ.cardHTML(null, true)" />
                </template>
                <template v-else>
                  <div
                    v-for="(card, i) in visiblePlayerCards"
                    :key="'p-' + i"
                    class="bac-card-slot"
                    v-html="BJ.cardHTML(card)"
                  />
                </template>
              </div>
            </section>

            <!-- Center points badges -->
            <div class="bac-center">
              <div class="bac-points-badge" :class="{ 'is-win': result?.outcome === 'player' }">
                <span class="bac-points-label">{{ t('casino.baccarat.player') }}</span>
                <strong>{{ shownPlayerPoints ?? '—' }}</strong>
              </div>
              <span class="bac-vs" aria-hidden="true">VS</span>
              <div class="bac-points-badge" :class="{ 'is-win': result?.outcome === 'banker' }">
                <span class="bac-points-label">{{ t('casino.baccarat.banker') }}</span>
                <strong>{{ shownBankerPoints ?? '—' }}</strong>
              </div>
              <p v-if="!result && !busy" class="bac-idle">{{ t('casino.baccarat.idleMsg') }}</p>
            </div>

            <!-- Banker hand -->
            <section class="bac-hand">
              <h2 class="bac-hand-title banker">{{ t('casino.baccarat.banker') }}</h2>
              <div class="bac-cards">
                <template v-if="!result">
                  <div v-for="i in 2" :key="'b-back-' + i" class="bac-card-slot" v-html="BJ.cardHTML(null, true)" />
                </template>
                <template v-else>
                  <div
                    v-for="(card, i) in visibleBankerCards"
                    :key="'b-' + i"
                    class="bac-card-slot"
                    v-html="BJ.cardHTML(card)"
                  />
                </template>
              </div>
            </section>
          </div>

          <!-- Result banner -->
          <transition name="fade">
            <div v-if="bannerVisible" class="bac-banner" :class="`bac-banner-${bannerTone}`" role="status" aria-live="polite">
              <strong>{{ bannerOutcome }}</strong>
              <span class="bac-banner-score">{{ bannerScore }}</span>
              <span v-if="bannerPayout" class="bac-banner-payout">
                {{ t('casino.baccarat.payout', { amount: formatBet(bannerPayout) }) }}
              </span>
            </div>
          </transition>

          <!-- Bet zone -->
          <div class="card p-5">
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <button
                v-for="opt in sideOptions"
                :key="opt.value"
                type="button"
                class="bac-side-btn"
                :class="{ 'is-active': side === opt.value }"
                :disabled="busy"
                @click="side = opt.value"
              >
                <strong>{{ opt.label }}</strong>
                <span class="bac-side-odds">{{ opt.odds }}</span>
              </button>
            </div>
          </div>

          <!-- Controls -->
          <div class="card p-6">
            <div class="grid gap-5 md:grid-cols-[1fr_auto] md:items-end">
              <div>
                <label for="casino-baccarat-bet" class="input-label">{{ t('casino.baccarat.betAmount') }}</label>
                <div class="mt-1 flex items-stretch gap-2">
                  <button type="button" class="btn btn-secondary px-3" aria-label="-" :disabled="busy" @click="stepBet(-10)">
                    −
                  </button>
                  <input
                    id="casino-baccarat-bet"
                    v-model.number="bet"
                    type="number"
                    class="input text-center"
                    :min="meta?.min_bet ?? 0"
                    :max="meta?.max_bet ?? 0"
                    step="0.01"
                    :disabled="busy"
                  />
                  <button type="button" class="btn btn-secondary px-3" aria-label="+" :disabled="busy" @click="stepBet(10)">
                    +
                  </button>
                </div>
                <div class="mt-2 flex flex-wrap gap-2">
                  <button
                    v-for="preset in presets"
                    :key="preset"
                    type="button"
                    class="chip-btn"
                    :class="{ 'chip-btn-active': Number(bet) === preset }"
                    :disabled="busy"
                    @click="bet = preset"
                  >
                    {{ preset }}
                  </button>
                </div>
              </div>
              <button
                type="button"
                class="btn btn-primary h-[58px] min-w-[180px] flex-col leading-tight"
                :disabled="busy"
                @click="deal"
              >
                <span class="text-base font-bold tracking-widest">
                  {{ busy ? t('casino.baccarat.dealing') : t('casino.baccarat.deal') }}
                </span>
                <small class="text-xs font-normal opacity-90">{{ pickLabel }}</small>
              </button>
            </div>
          </div>
        </div>

        <!-- Aside -->
        <div class="flex min-w-0 flex-col gap-6 self-start">
          <section class="card p-5">
            <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.common.infoTitle') }}</h2>
            <dl class="space-y-2.5 text-sm">
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.baccarat.currentPick') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ pickText }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.common.minBet') }} / {{ t('casino.common.maxBet') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">
                  {{ formatBet(meta?.min_bet ?? 0) }} – {{ formatBet(meta?.max_bet ?? 0) }}
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.baccarat.player') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ t('casino.baccarat.playerOdds') }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.baccarat.banker') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ t('casino.baccarat.bankerOdds') }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.baccarat.tie') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ t('casino.baccarat.tieOdds') }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.stats.gameRecords') }}</dt>
                <dd class="font-semibold" :class="statusClass">{{ statusText }}</dd>
              </div>
            </dl>
          </section>

          <section class="card">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-3.5 dark:border-dark-700">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.baccarat.winRecords') }}</h2>
              <router-link
                to="/casino/history"
                class="inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400"
              >
                {{ t('casino.common.more') }}
                <Icon name="chevronRight" size="xs" />
              </router-link>
            </div>
            <ul class="space-y-1 p-3">
              <li
                v-for="row in baccaratFeed"
                :key="row.id"
                class="flex items-center justify-between gap-3 rounded-xl px-3 py-2 text-sm hover:bg-gray-50 dark:hover:bg-dark-800"
              >
                <div class="min-w-0">
                  <p class="font-medium text-gray-900 dark:text-white">{{ t('casino.baccarat.title') }}</p>
                  <p class="text-xs text-gray-400 dark:text-dark-500">{{ formatDateTime(row.created_at) }}</p>
                </div>
                <span class="font-semibold" :class="row.delta >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                  {{ row.delta >= 0 ? '+' : '−' }}{{ formatBet(Math.abs(row.delta)) }}
                </span>
              </li>
              <li v-if="!baccaratFeed.length" class="px-3 py-6 text-center text-sm text-gray-400 dark:text-dark-500">
                {{ t('casino.baccarat.noWins') }}
              </li>
            </ul>
          </section>
        </div>
      </div>

      <p class="flex items-center justify-center gap-2 pb-2 text-xs text-gray-400 dark:text-dark-500">
        <Icon name="shield" size="sm" />
        {{ t('casino.disclaimer') }}
      </p>

      <CasinoResultModal
        :open="resultModal"
        :game="t('casino.games.baccarat')"
        :win="modalData.win"
        :delta="modalData.delta"
        :bet="modalData.bet"
        :payout="modalData.payout"
        :balance="modalData.balance"
        :title="modalData.title"
        can-again
        :bet-options="betPresets"
        :bet-value="Number(bet)"
        @update:bet-value="bet = $event"
        @close="resultModal = false"
        @again="onPlayAgain"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import CasinoStatsBar, { type CasinoStatItem } from '@/components/casino/CasinoStatsBar.vue'
import CasinoResultModal from '@/components/casino/CasinoResultModal.vue'
import { BJ } from '@/components/casino/renderers'
import { formatBet, useCasinoData } from '@/components/casino/useCasino'
import {
  casinoAPI,
  type CasinoBaccaratDealResult,
  type CasinoBaccaratSide,
  type CasinoBetRow,
  type CasinoBlackjackCard
} from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { meta, me, loadAll, refreshMe } = useCasinoData()

const bet = ref(10)
const side = ref<CasinoBaccaratSide>('player')
const busy = ref(false)
const resultModal = ref(false)
const modalData = ref({
  win: null as boolean | null,
  delta: 0,
  payout: 0,
  bet: 0,
  balance: null as number | null,
  title: undefined as string | undefined
})
const showRules = ref(false)
const result = ref<CasinoBaccaratDealResult | null>(null)
const revealed = ref({ player: 0, banker: 0 })
const lastBet = ref<number | null>(null)
const baccaratFeed = ref<CasinoBetRow[]>([])

let revealTimers: number[] = []

const REVEAL_STEP = 380

const balance = computed(() => me.value?.user.balance ?? authStore.user?.balance ?? 0)

const presets = computed(() =>
  [10, 20, 50, 100, 200].filter((v) => v <= (meta.value?.max_bet ?? 5000))
)

const sideOptions = computed(() => [
  { value: 'player' as CasinoBaccaratSide, label: t('casino.baccarat.player'), odds: t('casino.baccarat.playerOdds') },
  { value: 'banker' as CasinoBaccaratSide, label: t('casino.baccarat.banker'), odds: t('casino.baccarat.bankerOdds') },
  { value: 'tie' as CasinoBaccaratSide, label: t('casino.baccarat.tie'), odds: t('casino.baccarat.tieOdds') }
])

const selectedSide = computed(() => sideOptions.value.find((opt) => opt.value === side.value) || null)
const pickText = computed(() =>
  selectedSide.value ? `${selectedSide.value.label} ${selectedSide.value.odds}` : '—'
)
const pickLabel = computed(() =>
  selectedSide.value ? `${selectedSide.value.label} · ${selectedSide.value.odds}` : t('casino.baccarat.idleMsg')
)

/* Cards revealed so far (dealt one by one for the flip-in animation). */
const visiblePlayerCards = computed<CasinoBlackjackCard[]>(() =>
  (result.value?.player_cards || []).slice(0, revealed.value.player)
)
const visibleBankerCards = computed<CasinoBlackjackCard[]>(() =>
  (result.value?.banker_cards || []).slice(0, revealed.value.banker)
)

/** Baccarat points: A=1, faces=0, last digit of the total. */
function baccaratPts(cards: CasinoBlackjackCard[]): number {
  return cards.reduce((total, card) => {
    const rank = String(card?.r || '')
    if (rank === 'A') return total + 1
    if (['K', 'Q', 'J', '10'].includes(rank)) return total
    return total + (parseInt(rank, 10) || 0)
  }, 0) % 10
}

const shownPlayerPoints = computed(() =>
  result.value && revealed.value.player > 0 ? baccaratPts(visiblePlayerCards.value) : null
)
const shownBankerPoints = computed(() =>
  result.value && revealed.value.banker > 0 ? baccaratPts(visibleBankerCards.value) : null
)

/* Result banner */
const bannerVisible = computed(() => !!result.value && !busy.value)
const bannerOutcome = computed(() => {
  const outcome = result.value?.outcome
  if (outcome === 'player') return t('casino.baccarat.resultPlayer')
  if (outcome === 'banker') return t('casino.baccarat.resultBanker')
  return t('casino.baccarat.resultTie')
})
const bannerTone = computed(() => {
  const outcome = result.value?.outcome
  return outcome === 'player' ? 'player' : outcome === 'banker' ? 'banker' : 'tie'
})
const bannerScore = computed(() =>
  result.value ? `${result.value.player_points} : ${result.value.banker_points}` : ''
)
const bannerPayout = computed(() => {
  const r = result.value
  return r && r.payout > 0 ? r.payout : null
})

const statItems = computed<CasinoStatItem[]>(() => [
  {
    label: t('casino.stats.balance'),
    value: `$ ${formatBet(balance.value)}`,
    icon: 'creditCard',
    iconClass: 'bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
  },
  {
    label: t('casino.baccarat.roundBet'),
    value: formatBet(lastBet.value ?? bet.value),
    icon: 'dollar',
    iconClass: 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
  },
  {
    label: t('casino.baccarat.roundWin'),
    value: bannerPayout.value === null ? '—' : formatBet(bannerPayout.value),
    icon: 'gift',
    iconClass: 'bg-violet-50 text-violet-600 dark:bg-violet-900/30 dark:text-violet-400',
    valueClass: 'text-amber-600 dark:text-amber-400'
  },
  {
    label: t('casino.stats.todayProfit'),
    value: `${(me.value?.today_profit ?? 0) >= 0 ? '+' : '−'} ${formatBet(Math.abs(me.value?.today_profit ?? 0))}`,
    icon: 'trendingUp',
    iconClass: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400',
    valueClass:
      (me.value?.today_profit ?? 0) >= 0
        ? 'text-emerald-600 dark:text-emerald-400'
        : 'text-red-600 dark:text-red-400'
  }
])

const statusText = computed(() => {
  if (busy.value) return t('casino.status.playing')
  return result.value === null ? t('casino.status.idle') : t('casino.status.settled')
})
const statusClass = computed(() =>
  busy.value ? 'text-primary-600 dark:text-primary-400' : 'text-gray-900 dark:text-white'
)

function stepBet(delta: number) {
  const next = Number(bet.value) + delta
  if (Number.isFinite(next) && next > 0) bet.value = Number(next.toFixed(2))
}

function errorMessage(error: unknown): string {
  const err = error as { message?: string }
  return err?.message || t('casino.common.loadFailed')
}

function clearRevealTimers() {
  revealTimers.forEach((timer) => window.clearTimeout(timer))
  revealTimers = []
}

/** Reveal cards one by one, alternating player/banker (p1, b1, p2, b2, p3, b3). */
function animateReveal(outcome: CasinoBaccaratDealResult): Promise<void> {
  clearRevealTimers()
  revealed.value = { player: 0, banker: 0 }
  const playerCount = outcome.player_cards?.length ?? 0
  const bankerCount = outcome.banker_cards?.length ?? 0
  const order: ('player' | 'banker')[] = []
  for (let i = 0; i < Math.max(playerCount, bankerCount); i++) {
    if (i < playerCount) order.push('player')
    if (i < bankerCount) order.push('banker')
  }
  return new Promise((resolve) => {
    order.forEach((who, i) => {
      revealTimers.push(
        window.setTimeout(() => {
          revealed.value = { ...revealed.value, [who]: revealed.value[who] + 1 }
        }, REVEAL_STEP * (i + 1))
      )
    })
    revealTimers.push(window.setTimeout(resolve, REVEAL_STEP * (order.length + 1) + 120))
  })
}

async function deal() {
  if (busy.value) return
  const m = meta.value
  if (!m) return
  const value = Number(bet.value)
  if (!Number.isFinite(value) || value < m.min_bet || value > m.max_bet) {
    appStore.showWarning(t('casino.common.betRange', { min: formatBet(m.min_bet), max: formatBet(m.max_bet) }))
    return
  }
  if (value > balance.value) {
    appStore.showWarning(t('casino.common.insufficientBalance'))
    return
  }
  busy.value = true
  lastBet.value = value
  try {
    const outcome = await casinoAPI.baccaratDeal(value, side.value)
    result.value = outcome
    await animateReveal(outcome)
    await Promise.all([refreshMe(), authStore.refreshUser()])
    void loadFeed()
    showResult(outcome.payout, value, outcome.balance)
  } catch (error) {
    appStore.showError(errorMessage(error))
  } finally {
    busy.value = false
  }
}

const betPresets = computed(() =>
  [1, 2, 5, 10, 20, 50].filter((v) => v >= (meta.value?.min_bet ?? 0) && v <= (meta.value?.max_bet ?? Infinity))
)

function showResult(payout: number, bet: number, balance?: number | null) {
  modalData.value = {
    win: payout > bet ? true : payout === bet ? null : false,
    delta: payout - bet,
    payout,
    bet,
    balance: balance ?? null,
    title: resultTitle()
  }
  resultModal.value = true
}

function resultTitle(): string | undefined {
  const outcome = result.value?.outcome
  if (!outcome) return undefined
  if (outcome === 'player') return t('casino.modal.playerWin')
  if (outcome === 'banker') return t('casino.modal.bankerWin')
  if (outcome === 'tie') return t('casino.modal.push')
  return undefined
}

function onPlayAgain() {
  resultModal.value = false
  void deal()
}

async function loadFeed() {
  try {
    const { items } = await casinoAPI.getHistory(50)
    baccaratFeed.value = (items || []).filter((row) => row.game === 'baccarat').slice(0, 6)
  } catch (error) {
    console.error('Failed to load baccarat feed', error)
  }
}

onMounted(async () => {
  await loadAll()
  if (meta.value) {
    bet.value = Math.min(Math.max(bet.value, meta.value.min_bet), meta.value.max_bet)
  }
  void loadFeed()
})

onUnmounted(clearRevealTimers)
</script>

<style scoped>
/* ===== Table ===== */
.bac-table {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 18px;
  border-radius: 20px;
  padding: 30px 26px;
  background: radial-gradient(circle at 50% 24%, #ede9fe 0%, #f5f3ff 48%, #ffffff 80%);
  box-shadow: inset 0 0 0 1px rgba(109, 77, 246, 0.08);
}
.dark .bac-table {
  background: radial-gradient(circle at 50% 24%, #2b2554 0%, #1d1a3a 50%, #141628 84%);
  box-shadow: inset 0 0 0 1px rgba(139, 116, 248, 0.14);
}
.bac-hand {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  text-align: center;
}
.bac-hand-title {
  font-size: 15px;
  font-weight: 800;
  letter-spacing: 0.06em;
}
.bac-hand-title.player {
  color: #3d6ee8;
}
.bac-hand-title.banker {
  color: #d63742;
}
.bac-cards {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 9px;
  min-height: 112px;
}
.bac-card-slot {
  display: inline-flex;
  flex: none;
}
@keyframes bac-flip-in {
  from {
    opacity: 0;
    transform: perspective(700px) rotateY(88deg) scale(0.94);
  }
  to {
    opacity: 1;
    transform: perspective(700px) rotateY(0deg) scale(1);
  }
}
.bac-card-slot :deep(.bj-card) {
  animation: bac-flip-in 0.45s ease-out both;
}

/* ===== Center points badges ===== */
.bac-center {
  display: flex;
  min-width: 96px;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.bac-points-badge {
  display: flex;
  min-width: 84px;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  border-radius: 12px;
  border: 1px solid rgba(109, 77, 246, 0.18);
  background: #fff;
  padding: 7px 14px;
  box-shadow: 0 4px 10px rgba(63, 43, 168, 0.08);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}
.dark .bac-points-badge {
  border-color: rgb(48 55 73);
  background: rgb(23 28 43);
}
.bac-points-badge.is-win {
  border-color: #6d4df6;
  box-shadow: 0 0 0 3px rgba(109, 77, 246, 0.16);
}
.bac-points-label {
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: rgb(156 163 175);
}
.bac-points-badge strong {
  font-size: 22px;
  font-weight: 800;
  color: rgb(17 24 39);
}
.dark .bac-points-badge strong {
  color: #fff;
}
.bac-vs {
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.14em;
  color: #a5a0c8;
}
.bac-idle {
  max-width: 130px;
  text-align: center;
  font-size: 11.5px;
  line-height: 1.5;
  color: rgb(156 163 175);
}

/* ===== Result banner ===== */
.bac-banner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-radius: 16px;
  border: 1px solid transparent;
  padding: 14px 20px;
  font-size: 15px;
}
.bac-banner strong {
  font-size: 16px;
  font-weight: 800;
}
.bac-banner-player {
  border-color: rgba(61, 110, 232, 0.25);
  background: #eef4ff;
  color: #2c5cc5;
}
.bac-banner-banker {
  border-color: rgba(214, 55, 66, 0.22);
  background: #fdeeee;
  color: #b02a35;
}
.bac-banner-tie {
  border-color: rgba(217, 119, 6, 0.25);
  background: #fef6e7;
  color: #b45309;
}
.dark .bac-banner-player {
  background: rgba(61, 110, 232, 0.14);
  color: #93b4ff;
}
.dark .bac-banner-banker {
  background: rgba(214, 55, 66, 0.14);
  color: #fca5a5;
}
.dark .bac-banner-tie {
  background: rgba(217, 119, 6, 0.14);
  color: #fcd34d;
}
.bac-banner-score {
  padding: 2px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.75);
  font-weight: 800;
  letter-spacing: 0.08em;
}
.dark .bac-banner-score {
  background: rgba(20, 22, 40, 0.55);
}
.bac-banner-payout {
  font-weight: 700;
}

/* ===== Side (bet) buttons ===== */
.bac-side-btn {
  display: flex;
  min-height: 68px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border-radius: 14px;
  border: 2px solid rgb(229 231 235);
  background: #fff;
  padding: 10px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;
}
.bac-side-btn strong {
  font-size: 16px;
  font-weight: 800;
  color: rgb(17 24 39);
}
.bac-side-btn:hover:not(:disabled) {
  border-color: rgb(139 116 248);
  transform: translateY(-1px);
}
.bac-side-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}
.bac-side-btn.is-active {
  border-color: #6d4df6;
  box-shadow: 0 0 0 3px rgba(109, 77, 246, 0.18);
}
.bac-side-btn.is-active strong {
  color: #6d4df6;
}
.bac-side-odds {
  padding: 1px 10px;
  border-radius: 999px;
  background: #f1eeff;
  color: #6d4df6;
  font-size: 12px;
  font-weight: 700;
}
.dark .bac-side-btn {
  border-color: rgb(48 55 73);
  background: rgb(23 28 43);
}
.dark .bac-side-btn strong {
  color: #f3f4f6;
}
.dark .bac-side-btn.is-active strong {
  color: #a78bfa;
}
.dark .bac-side-odds {
  background: rgba(109, 77, 246, 0.18);
  color: #a78bfa;
}

/* ===== Chips (bet presets) ===== */
.chip-btn {
  border-radius: 10px;
  border: 1px solid rgb(229 231 235);
  background: #fff;
  padding: 4px 14px;
  font-size: 13px;
  font-weight: 600;
  color: rgb(75 85 99);
  transition: all 0.15s ease;
}
.dark .chip-btn {
  border-color: rgb(48 55 73);
  background: rgb(23 28 43);
  color: rgb(209 213 219);
}
.chip-btn:hover:not(:disabled) {
  border-color: rgb(129 140 248);
  color: rgb(99 102 241);
}
.chip-btn-active {
  border-color: #6d4df6;
  background: #6d4df6;
  color: #fff;
}
.chip-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
.fade-enter-active,
.fade-leave-active {
  transition: all 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
@media (max-width: 760px) {
  .bac-table {
    grid-template-columns: 1fr;
  }
  .bac-center {
    flex-direction: row;
    justify-content: center;
    min-width: 0;
  }
  .bac-vs {
    margin: 0 4px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .bac-card-slot :deep(.bj-card) {
    animation: none;
  }
  .bac-side-btn,
  .chip-btn {
    transition: none;
  }
}
</style>
