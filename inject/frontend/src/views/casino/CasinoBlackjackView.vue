<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('casino.blackjack.title') }}
            <span class="ml-1 align-middle text-xl" aria-hidden="true">🃏</span>
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.blackjack.description') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <router-link to="/casino" class="btn btn-secondary">
            {{ t('casino.backToLobby') }}
          </router-link>
          <button type="button" class="btn btn-secondary" @click="showRules = !showRules">
          <Icon name="questionCircle" size="md" />
          {{ t('casino.common.infoTitle') }}
        </button>        </div>
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
            <li>{{ t('casino.blackjack.rulePays', { pays: paysText }) }}　·　{{ dealerRuleText }}　·　{{ t('casino.blackjack.ruleNoInsurance') }}</li>
            <li>{{ t('casino.common.betRange', { min: formatBet(meta?.min_bet ?? 0), max: formatBet(meta?.max_bet ?? 0) }) }}</li>
          </ul>
        </div>
      </transition>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div class="flex min-w-0 flex-col gap-6">
          <BlackjackTable :game="game" :rules="meta?.blackjack || null" />

          <!-- Result message -->
          <p class="bj-msg" :class="msgClass" role="status" aria-live="polite">{{ msg }}</p>

          <!-- Bet zone -->
          <div class="card p-5">
            <div class="flex flex-wrap items-end justify-between gap-4">
              <div class="flex flex-wrap items-center gap-2">
                <button
                  v-for="chip in chips"
                  :key="chip.value"
                  type="button"
                  class="table-chip"
                  :class="[`table-chip-${chip.tone}`, { 'table-chip-active': Number(bet) === chip.value }]"
                  :disabled="roundLocked"
                  @click="bet = chip.value"
                >
                  {{ chip.label }}
                </button>
              </div>
              <div class="flex items-end gap-3">
                <div>
                  <label for="casino-bj-bet" class="input-label">{{ t('casino.blackjack.betLabel') }}</label>
                  <input
                    id="casino-bj-bet"
                    v-model.number="bet"
                    type="number"
                    class="input mt-1 w-36 text-center"
                    :min="meta?.min_bet ?? 0"
                    :max="meta?.max_bet ?? 0"
                    step="0.01"
                    :disabled="roundLocked"
                  />
                </div>
                <button type="button" class="btn btn-primary h-[42px] px-8" :disabled="roundLocked" @click="deal">
                  {{ game?.status === 'settled' ? t('casino.blackjack.dealAgain') : t('casino.blackjack.deal') }}
                </button>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="grid grid-cols-2 gap-3 md:grid-cols-4">
            <button type="button" class="btn btn-secondary h-12" :disabled="!roundActive || busy" @click="hit">
              <Icon name="bolt" size="md" />
              {{ t('casino.blackjack.hit') }}
            </button>
            <button type="button" class="btn btn-secondary h-12" :disabled="!roundActive || busy" @click="stand">
              <Icon name="ban" size="md" />
              {{ t('casino.blackjack.stand') }}
            </button>
            <button type="button" class="btn btn-secondary h-12" :disabled="!roundActive || !canDouble || busy" @click="double">
              <strong class="mr-1">2×</strong>
              {{ t('casino.blackjack.double') }}
            </button>
            <button type="button" class="btn btn-secondary h-12 opacity-60" disabled :title="t('casino.blackjack.splitDisabled')">
              <Icon name="copy" size="md" />
              {{ t('casino.blackjack.split') }}
            </button>
          </div>
        </div>

        <!-- Aside -->
        <div class="flex min-w-0 flex-col gap-6 self-start">
          <section class="card p-5">
            <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.common.infoTitle') }}</h2>
            <dl class="space-y-2.5 text-sm">
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.blackjack.currentBet') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ formatBet(roundBet) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.common.minBet') }} / {{ t('casino.common.maxBet') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">
                  {{ formatBet(meta?.min_bet ?? 0) }} – {{ formatBet(meta?.max_bet ?? 0) }}
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.blackjack.double') }}</dt>
                <dd class="font-semibold" :class="doubleAllowed ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                  {{ doubleAllowed ? '✓' : '×' }}
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.stats.gameRecords') }}</dt>
                <dd class="font-semibold" :class="statusClass">{{ statusText }}</dd>
              </div>
            </dl>
          </section>

          <section class="card">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-3.5 dark:border-dark-700">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.lobby.recentPlays') }}</h2>
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
                v-for="row in bjFeed"
                :key="row.id"
                class="flex items-center justify-between gap-3 rounded-xl px-3 py-2 text-sm hover:bg-gray-50 dark:hover:bg-dark-800"
              >
                <div class="min-w-0">
                  <p class="font-medium text-gray-900 dark:text-white">{{ t('casino.games.blackjack') }}</p>
                  <p class="text-xs text-gray-400 dark:text-dark-500">{{ formatDateTime(row.created_at) }}</p>
                </div>
                <span
                  class="font-semibold"
                  :class="row.delta >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'"
                >
                  {{ row.delta >= 0 ? '+' : '−' }}{{ formatBet(Math.abs(row.delta)) }}
                </span>
              </li>
              <li v-if="!bjFeed.length" class="px-3 py-6 text-center text-sm text-gray-400 dark:text-dark-500">
                {{ t('casino.lobby.noRecords') }}
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
        :game="t('casino.games.blackjack')"
        :win="modalData.win"
        :delta="modalData.delta"
        :bet="modalData.bet"
        :payout="modalData.payout"
        :balance="modalData.balance"
        :title="modalData.title"
        :emoji="modalData.emoji"
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
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import CasinoStatsBar, { type CasinoStatItem } from '@/components/casino/CasinoStatsBar.vue'
import CasinoResultModal from '@/components/casino/CasinoResultModal.vue'
import BlackjackTable from '@/components/casino/BlackjackTable.vue'
import { formatBet, useCasinoData } from '@/components/casino/useCasino'
import {
  casinoAPI,
  type CasinoBetRow,
  type CasinoBlackjackGame
} from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { meta, me, loadAll, refreshMe } = useCasinoData()

const game = ref<CasinoBlackjackGame | null>(null)
const busy = ref(false)
const resultModal = ref(false)
const modalData = ref({
  win: null as boolean | null,
  delta: 0,
  payout: 0,
  bet: 0,
  balance: null as number | null,
  title: undefined as string | undefined,
  emoji: undefined as string | undefined
})
const bet = ref(10)
const showRules = ref(false)
const bjFeed = ref<CasinoBetRow[]>([])

const balance = computed(() => game.value?.balance ?? me.value?.user.balance ?? authStore.user?.balance ?? 0)

const roundActive = computed(() => game.value?.status === 'active')
const roundLocked = computed(() => roundActive.value || busy.value)
const roundBet = computed(() =>
  game.value ? game.value.bet * (game.value.doubled ? 2 : 1) : Number(bet.value)
)
const doubleAllowed = computed(() => meta.value?.blackjack.double_allowed ?? true)
const canDouble = computed(() => !!game.value?.can_double && doubleAllowed.value)

const paysText = computed(() => {
  const pays = meta.value?.blackjack.blackjack_pays ?? 2.5
  return pays === 2.5 ? '3:2' : `${pays - 1}:1`
})
const dealerRuleText = computed(() =>
  meta.value?.blackjack.dealer_stands_soft17
    ? t('casino.blackjack.ruleDealerStands')
    : t('casino.blackjack.ruleDealerHits')
)

const chips = computed(() =>
  [
    { value: 10, tone: 'gray', label: '10' },
    { value: 50, tone: 'red', label: '50' },
    { value: 100, tone: 'green', label: '100' },
    { value: 500, tone: 'black', label: '500' },
    { value: 1000, tone: 'purple', label: '1K' },
    { value: 5000, tone: 'gold', label: '5K' }
  ].filter((chip) => chip.value <= (meta.value?.max_bet ?? 5000))
)

const statItems = computed<CasinoStatItem[]>(() => [
  {
    label: t('casino.stats.balance'),
    value: `$ ${formatBet(balance.value)}`,
    icon: 'creditCard',
    iconClass: 'bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
  },
  {
    label: t('casino.blackjack.currentBet'),
    value: formatBet(roundBet.value),
    icon: 'dollar',
    iconClass: 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
  },
  {
    label: t('casino.blackjack.roundWin'),
    value: game.value ? formatBet(game.value.payout) : '—',
    icon: 'trophy',
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
  if (!game.value) return t('casino.status.idle')
  return game.value.status === 'active' ? t('casino.status.playing') : t('casino.status.settled')
})
const statusClass = computed(() =>
  roundActive.value || busy.value ? 'text-primary-600 dark:text-primary-400' : 'text-gray-900 dark:text-white'
)

const resultText = computed(() => {
  const g = game.value
  if (!g || g.status !== 'settled') return ''
  const keyByResult: Record<string, string> = {
    win: 'casino.blackjack.resultWin',
    blackjack: 'casino.blackjack.resultBlackjack',
    push: 'casino.blackjack.resultPush',
    bust: 'casino.blackjack.resultBust',
    lose: 'casino.blackjack.resultLose'
  }
  return keyByResult[g.result || ''] ? t(keyByResult[g.result || '']) : t('casino.status.settled')
})

const msg = computed(() => {
  const g = game.value
  if (g?.status === 'settled') {
    const delta = g.payout - g.bet * (g.doubled ? 2 : 1)
    return delta !== 0 ? `${resultText.value} · ${delta > 0 ? '+' : '−'}${formatBet(Math.abs(delta))}` : resultText.value
  }
  if (!g) return t('casino.blackjack.idleMsg')
  return t('casino.blackjack.thinking')
})
const msgClass = computed(() => {
  const g = game.value
  if (g?.status !== 'settled') return ''
  const delta = g.payout - g.bet * (g.doubled ? 2 : 1)
  if (delta > 0) return 'msg-pos'
  if (delta < 0) return 'msg-neg'
  return ''
})

function errorMessage(error: unknown): string {
  const err = error as { message?: string }
  return err?.message || t('casino.common.loadFailed')
}

function syncAfterSettle() {
  return Promise.all([refreshMe(), authStore.refreshUser()])
}

async function deal() {
  if (roundLocked.value) return
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
  try {
    const { game: dealt } = await casinoAPI.blackjackDeal(value)
    game.value = dealt
    await syncAfterSettle()
    void loadFeed()
    if (dealt.status === 'settled') showResult(dealt.payout, dealt.bet, dealt.balance ?? null)
  } catch (error) {
    appStore.showError(errorMessage(error))
  } finally {
    busy.value = false
  }
}

async function action(kind: 'hit' | 'stand' | 'double') {
  if (!roundActive.value || busy.value || !game.value) return
  busy.value = true
  try {
    const caller = {
      hit: casinoAPI.blackjackHit,
      stand: casinoAPI.blackjackStand,
      double: casinoAPI.blackjackDouble
    }[kind]
    const { game: updated } = await caller(game.value.game_id)
    game.value = updated
    await syncAfterSettle()
    void loadFeed()
    if (updated.status === 'settled') showResult(updated.payout, updated.bet, updated.balance ?? null)
  } catch (error) {
    appStore.showError(errorMessage(error))
    try {
      const current = await casinoAPI.blackjackCurrent()
      game.value = current.game
    } catch {
      // keep the last confirmed hand
    }
  } finally {
    busy.value = false
  }
}

const hit = () => action('hit')
const stand = () => action('stand')
const double = () => action('double')

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
    title: resultTitle(),
    emoji: undefined
  }
  resultModal.value = true
}

function resultTitle(): string | undefined {
  if (!game.value) return undefined
  const { payout, bet, player_cards } = game.value
  // 自然 blackjack：首两张即 21 点且 payout > bet × 2（3:2 派彩）
  const natural = player_cards?.length === 2 && payout > bet * 2
  if (natural) return t('casino.modal.blackjackTitle')
  if (payout > bet) return t('casino.modal.winTitle')
  if (payout === bet) return t('casino.modal.push')
  if (payout < bet) return t('casino.modal.bankerWin') // 21 点输了即庄家胜
  return undefined
}

function onPlayAgain() {
  resultModal.value = false
  void deal()
}

async function loadFeed() {
  try {
    const { items } = await casinoAPI.getHistory(50)
    bjFeed.value = (items || []).filter((row) => row.game === 'blackjack').slice(0, 6)
  } catch (error) {
    console.error('Failed to load blackjack feed', error)
  }
}

onMounted(async () => {
  await loadAll()
  if (meta.value) {
    bet.value = Math.min(Math.max(bet.value, meta.value.min_bet), meta.value.max_bet)
  }
  try {
    const current = await casinoAPI.blackjackCurrent()
    game.value = current.game
  } catch (error) {
    console.error('Failed to restore blackjack game', error)
  }
  void loadFeed()
})
</script>

<style scoped>
.bj-msg {
  min-height: 24px;
  text-align: center;
  font-size: 15px;
  font-weight: 700;
  color: rgb(107 114 128);
}
.msg-pos {
  color: rgb(5 150 105);
}
.msg-neg {
  color: rgb(220 38 38);
}
.table-chip {
  display: inline-flex;
  min-width: 52px;
  height: 52px;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  border: 4px dashed rgba(255, 255, 255, 0.85);
  box-shadow: 0 4px 10px rgba(30, 30, 60, 0.25), inset 0 0 0 3px rgba(0, 0, 0, 0.08);
  color: #fff;
  font-size: 14px;
  font-weight: 800;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}
.table-chip:hover:not(:disabled) {
  transform: translateY(-2px);
}
.table-chip:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}
.table-chip-active {
  box-shadow: 0 0 0 3px #6d4df6, 0 6px 14px rgba(109, 77, 246, 0.4);
  transform: translateY(-2px);
}
.table-chip-gray {
  background: radial-gradient(circle at 50% 35%, #9ca3af, #4b5563);
}
.table-chip-red {
  background: radial-gradient(circle at 50% 35%, #f87171, #b91c1c);
}
.table-chip-green {
  background: radial-gradient(circle at 50% 35%, #34d399, #047857);
}
.table-chip-black {
  background: radial-gradient(circle at 50% 35%, #4b5563, #111827);
}
.table-chip-purple {
  background: radial-gradient(circle at 50% 35%, #a78bfa, #5b21b6);
}
.table-chip-gold {
  background: radial-gradient(circle at 50% 35%, #fbbf24, #b45309);
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
@media (prefers-reduced-motion: reduce) {
  .table-chip {
    transition: none;
  }
}
</style>
