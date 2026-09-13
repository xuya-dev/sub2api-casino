<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('casino.sicbo.title') }}
            <span class="ml-1 align-middle text-xl" aria-hidden="true">🎲</span>
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.sicbo.description') }}</p>
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
            <li>{{ t('casino.sicbo.ruleBets') }}</li>
            <li>{{ t('casino.sicbo.ruleTriple') }}</li>
            <li>{{ t('casino.common.betRange', { min: formatBet(meta?.min_bet ?? 0), max: formatBet(meta?.max_bet ?? 0) }) }}</li>
          </ul>
        </div>
      </transition>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div class="flex min-w-0 flex-col gap-6">
          <!-- Dice stage -->
          <div class="sicbo-stage">
            <div class="sicbo-dice" role="img" :aria-label="diceAriaLabel">
              <div
                v-for="(die, i) in dice"
                :key="i"
                class="sicbo-die"
                :class="{ 'is-tumbling': rolling && !frozen[i], 'is-kept': rolling && frozen[i] }"
              >
                <span
                  v-for="cell in 9"
                  :key="cell"
                  class="sicbo-pip"
                  :class="{ on: pipsFor(die).includes(cell - 1) }"
                />
              </div>
            </div>
            <p class="game-result" :class="resultPos ? 'pos' : result ? 'neg' : ''" role="status" aria-live="polite">
              {{ resultText }}
            </p>
          </div>

          <!-- Bet type buttons -->
          <div class="card p-5">
            <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
              <button
                v-for="opt in betOptions"
                :key="opt.value"
                type="button"
                class="sicbo-bet-btn"
                :class="{ 'is-active': betType === opt.value }"
                :disabled="rolling"
                @click="betType = opt.value"
              >
                <strong>{{ opt.label }}</strong>
                <small>{{ opt.range }}</small>
                <span class="sicbo-odds" :class="{ visible: betType === opt.value }">
                  {{ t('casino.sicbo.oddsOneToOne') }}
                </span>
              </button>
            </div>
            <p class="sicbo-note">
              <Icon name="exclamationTriangle" size="sm" />
              {{ t('casino.sicbo.tripleNote') }}
            </p>
          </div>

          <!-- Controls -->
          <div class="card p-6">
            <div class="grid gap-5 md:grid-cols-[1fr_auto] md:items-end">
              <div>
                <label for="casino-sicbo-bet" class="input-label">{{ t('casino.sicbo.betAmount') }}</label>
                <div class="mt-1 flex items-stretch gap-2">
                  <button type="button" class="btn btn-secondary px-3" aria-label="-" :disabled="rolling" @click="stepBet(-10)">
                    −
                  </button>
                  <input
                    id="casino-sicbo-bet"
                    v-model.number="bet"
                    type="number"
                    class="input text-center"
                    :min="meta?.min_bet ?? 0"
                    :max="meta?.max_bet ?? 0"
                    step="0.01"
                    :disabled="rolling"
                  />
                  <button type="button" class="btn btn-secondary px-3" aria-label="+" :disabled="rolling" @click="stepBet(10)">
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
                    :disabled="rolling"
                    @click="bet = preset"
                  >
                    {{ preset }}
                  </button>
                </div>
              </div>
              <button
                type="button"
                class="btn btn-primary h-[58px] min-w-[180px] flex-col leading-tight"
                :disabled="rolling"
                @click="roll"
              >
                <span class="text-base font-bold tracking-widest">
                  {{ rolling ? t('casino.sicbo.rolling') : t('casino.sicbo.roll') }}
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
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.sicbo.currentPick') }}</dt>
                <dd class="font-semibold" :class="pickLabelClass">{{ pickText }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.common.minBet') }} / {{ t('casino.common.maxBet') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">
                  {{ formatBet(meta?.min_bet ?? 0) }} – {{ formatBet(meta?.max_bet ?? 0) }}
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.sicbo.oddsOneToOne') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">1:1</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.stats.gameRecords') }}</dt>
                <dd class="font-semibold" :class="statusClass">{{ statusText }}</dd>
              </div>
            </dl>
          </section>

          <section class="card">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-3.5 dark:border-dark-700">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.sicbo.winRecords') }}</h2>
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
                v-for="row in sicboFeed"
                :key="row.id"
                class="flex items-center justify-between gap-3 rounded-xl px-3 py-2 text-sm hover:bg-gray-50 dark:hover:bg-dark-800"
              >
                <div class="min-w-0">
                  <p class="font-medium text-gray-900 dark:text-white">{{ t('casino.sicbo.title') }}</p>
                  <p class="text-xs text-gray-400 dark:text-dark-500">{{ formatDateTime(row.created_at) }}</p>
                </div>
                <span class="font-semibold" :class="row.delta >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                  {{ row.delta >= 0 ? '+' : '−' }}{{ formatBet(Math.abs(row.delta)) }}
                </span>
              </li>
              <li v-if="!sicboFeed.length" class="px-3 py-6 text-center text-sm text-gray-400 dark:text-dark-500">
                {{ t('casino.sicbo.noWins') }}
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
        :game="t('casino.games.sicbo')"
        :win="modalData.win"
        :delta="modalData.delta"
        :bet="modalData.bet"
        :payout="modalData.payout"
        :balance="modalData.balance"
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
import { formatBet, useCasinoData } from '@/components/casino/useCasino'
import { casinoAPI, type CasinoBetRow, type CasinoSicboBetType, type CasinoSicboRollResult } from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { meta, me, loadAll, refreshMe } = useCasinoData()

const bet = ref(10)
const betType = ref<CasinoSicboBetType | null>(null)
const rolling = ref(false)
const resultModal = ref(false)
const modalData = ref({ win: null as boolean | null, delta: 0, payout: 0, bet: 0, balance: null as number | null })
const showRules = ref(false)
const dice = ref<number[]>([2, 4, 6])
const frozen = ref<boolean[]>([true, true, true])
const result = ref<CasinoSicboRollResult | null>(null)
const lastBet = ref<number | null>(null)
const sicboFeed = ref<CasinoBetRow[]>([])

let tumbleTimer: number | undefined
let freezeTimers: number[] = []

const balance = computed(() => me.value?.user.balance ?? authStore.user?.balance ?? 0)

const presets = computed(() =>
  [10, 20, 50, 100, 200].filter((v) => v <= (meta.value?.max_bet ?? 5000))
)

/* Dice pip layout on a 3x3 grid (cell indexes 0..8). */
const DIE_PIPS: Record<number, number[]> = {
  1: [4],
  2: [0, 8],
  3: [0, 4, 8],
  4: [0, 2, 6, 8],
  5: [0, 2, 4, 6, 8],
  6: [0, 2, 3, 5, 6, 8]
}

function pipsFor(value: number): number[] {
  return DIE_PIPS[value] || []
}

const betOptions = computed(() => [
  { value: 'big' as CasinoSicboBetType, label: t('casino.sicbo.betBig'), range: t('casino.sicbo.betBigRange') },
  { value: 'small' as CasinoSicboBetType, label: t('casino.sicbo.betSmall'), range: t('casino.sicbo.betSmallRange') },
  { value: 'odd' as CasinoSicboBetType, label: t('casino.sicbo.betOdd'), range: t('casino.sicbo.betOddRange') },
  { value: 'even' as CasinoSicboBetType, label: t('casino.sicbo.betEven'), range: t('casino.sicbo.betEvenRange') }
])

const selectedOption = computed(() => betOptions.value.find((opt) => opt.value === betType.value) || null)

const pickText = computed(() => selectedOption.value?.label ?? '—')
const pickLabelClass = computed(() =>
  selectedOption.value ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-dark-500'
)
const pickLabel = computed(() =>
  selectedOption.value
    ? `${selectedOption.value.label} · ${selectedOption.value.range}`
    : t('casino.sicbo.pickBetFirst')
)

const outcomeLabel = computed(() => {
  const r = result.value
  if (!r) return ''
  const map: Record<string, string> = {
    big: t('casino.sicbo.resultBig'),
    small: t('casino.sicbo.resultSmall'),
    odd: t('casino.sicbo.resultOdd'),
    even: t('casino.sicbo.resultEven'),
    triple: t('casino.sicbo.resultTripleLabel')
  }
  return map[r.result] || r.result
})

const resultText = computed(() => {
  if (rolling.value) return t('casino.sicbo.rolling')
  const r = result.value
  if (!r) return ''
  const sumText = t('casino.sicbo.resultSum', { sum: r.sum })
  if (r.result === 'triple') return `${sumText} · ${t('casino.sicbo.resultTriple')}`
  const delta = r.payout - r.bet
  const verdict =
    delta > 0
      ? t('casino.sicbo.resultWin', { amount: formatBet(r.payout) })
      : t('casino.sicbo.resultLose')
  return `${sumText} · ${outcomeLabel.value} · ${verdict}`
})

const resultPos = computed(() => {
  const r = result.value
  return !!r && !rolling.value && r.payout - r.bet > 0
})

const diceAriaLabel = computed(() =>
  rolling.value ? t('casino.sicbo.rolling') : result.value ? t('casino.sicbo.resultSum', { sum: result.value.sum }) : t('casino.sicbo.title')
)

const statItems = computed<CasinoStatItem[]>(() => [
  {
    label: t('casino.stats.balance'),
    value: `$ ${formatBet(balance.value)}`,
    icon: 'creditCard',
    iconClass: 'bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
  },
  {
    label: t('casino.sicbo.roundBet'),
    value: formatBet(lastBet.value ?? bet.value),
    icon: 'dollar',
    iconClass: 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
  },
  {
    label: t('casino.sicbo.diceSumLabel'),
    value: result.value && !rolling.value ? String(result.value.sum) : '—',
    icon: 'bolt',
    iconClass: 'bg-sky-50 text-sky-600 dark:bg-sky-900/30 dark:text-sky-400'
  },
  {
    label: t('casino.sicbo.roundWin'),
    value: result.value && !rolling.value ? formatBet(result.value.payout) : '—',
    icon: 'gift',
    iconClass: 'bg-violet-50 text-violet-600 dark:bg-violet-900/30 dark:text-violet-400',
    valueClass: 'text-amber-600 dark:text-amber-400'
  }
])

const statusText = computed(() => {
  if (rolling.value) return t('casino.status.playing')
  return result.value === null ? t('casino.status.idle') : t('casino.status.settled')
})
const statusClass = computed(() =>
  rolling.value ? 'text-primary-600 dark:text-primary-400' : 'text-gray-900 dark:text-white'
)

function stepBet(delta: number) {
  const next = Number(bet.value) + delta
  if (Number.isFinite(next) && next > 0) bet.value = Number(next.toFixed(2))
}

function errorMessage(error: unknown): string {
  const err = error as { message?: string }
  return err?.message || t('casino.common.loadFailed')
}

function clearDiceTimers() {
  if (tumbleTimer !== undefined) {
    window.clearInterval(tumbleTimer)
    tumbleTimer = undefined
  }
  freezeTimers.forEach((timer) => window.clearTimeout(timer))
  freezeTimers = []
}

/** Tumble all dice for ~1.2s, freezing them onto `final` one by one. */
function animateDice(final: number[]): Promise<void> {
  clearDiceTimers()
  frozen.value = [false, false, false]
  tumbleTimer = window.setInterval(() => {
    dice.value = dice.value.map((value, i) => (frozen.value[i] ? value : 1 + Math.floor(Math.random() * 6)))
  }, 90)
  return new Promise((resolve) => {
    final.slice(0, 3).forEach((value, i) => {
      freezeTimers.push(
        window.setTimeout(() => {
          frozen.value = frozen.value.map((isFrozen, j) => (j === i ? true : isFrozen))
          dice.value = dice.value.map((current, j) => (j === i ? value : current))
        }, 420 * (i + 1))
      )
    })
    freezeTimers.push(
      window.setTimeout(() => {
        clearDiceTimers()
        dice.value = [...final]
        frozen.value = [true, true, true]
        resolve()
      }, 420 * 3 + 60)
    )
  })
}

async function roll() {
  if (rolling.value) return
  const m = meta.value
  if (!m) return
  if (!betType.value) {
    appStore.showWarning(t('casino.sicbo.pickBetFirst'))
    return
  }
  const value = Number(bet.value)
  if (!Number.isFinite(value) || value < m.min_bet || value > m.max_bet) {
    appStore.showWarning(t('casino.common.betRange', { min: formatBet(m.min_bet), max: formatBet(m.max_bet) }))
    return
  }
  if (value > balance.value) {
    appStore.showWarning(t('casino.common.insufficientBalance'))
    return
  }
  rolling.value = true
  result.value = null
  lastBet.value = value
  try {
    const outcome = await casinoAPI.sicboRoll(value, betType.value)
    await animateDice(outcome.dice || [])
    result.value = outcome
    await Promise.all([refreshMe(), authStore.refreshUser()])
    void loadFeed()
    showResult(outcome.payout, value, outcome.balance)
  } catch (error) {
    frozen.value = [true, true, true]
    appStore.showError(errorMessage(error))
  } finally {
    rolling.value = false
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
    balance: balance ?? null
  }
  resultModal.value = true
}

function onPlayAgain() {
  resultModal.value = false
  void roll()
}

async function loadFeed() {
  try {
    const { items } = await casinoAPI.getHistory(50)
    sicboFeed.value = (items || []).filter((row) => row.game === 'sicbo').slice(0, 6)
  } catch (error) {
    console.error('Failed to load sicbo feed', error)
  }
}

onMounted(async () => {
  await loadAll()
  if (meta.value) {
    bet.value = Math.min(Math.max(bet.value, meta.value.min_bet), meta.value.max_bet)
  }
  void loadFeed()
})

onUnmounted(clearDiceTimers)
</script>

<style scoped>
/* ===== Dice stage ===== */
.sicbo-stage {
  position: relative;
  overflow: hidden;
  border-radius: 20px;
  padding: 46px 24px 30px;
  background: radial-gradient(circle at 50% 26%, #ede9fe 0%, #f5f3ff 45%, #ffffff 78%);
  box-shadow: inset 0 0 0 1px rgba(109, 77, 246, 0.08);
}
.dark .sicbo-stage {
  background: radial-gradient(circle at 50% 26%, #2b2554 0%, #1d1a3a 48%, #141628 82%);
  box-shadow: inset 0 0 0 1px rgba(139, 116, 248, 0.14);
}
.sicbo-dice {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: clamp(18px, 4vw, 40px);
  min-height: 128px;
}
.sicbo-die {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  grid-template-rows: repeat(3, 1fr);
  gap: 7px;
  width: clamp(76px, 9vw, 106px);
  aspect-ratio: 1;
  padding: 14px;
  border-radius: 20px;
  background: linear-gradient(150deg, #ffffff 12%, #e9e5fb 88%);
  box-shadow:
    0 12px 24px rgba(63, 43, 168, 0.2),
    inset 0 -4px 8px rgba(63, 43, 168, 0.1),
    inset 0 0 0 1px rgba(63, 43, 168, 0.14);
}
.sicbo-pip {
  border-radius: 50%;
  background: transparent;
}
.sicbo-pip.on {
  background: radial-gradient(circle at 35% 30%, #7c5cf8, #4c2bd4);
  box-shadow: inset 0 -1px 2px rgba(255, 255, 255, 0.4);
}
@keyframes dice-tumble {
  0% { transform: rotate(0deg) translateY(0); }
  25% { transform: rotate(-22deg) translateY(-10px) scale(1.03); }
  50% { transform: rotate(16deg) translateY(3px); }
  75% { transform: rotate(-10deg) translateY(-6px) scale(1.02); }
  100% { transform: rotate(0deg) translateY(0); }
}
.sicbo-die.is-tumbling {
  animation: dice-tumble 0.34s ease-in-out infinite;
}
@keyframes dice-land {
  0% { transform: scale(1.14); }
  100% { transform: scale(1); }
}
.sicbo-die.is-kept {
  animation: dice-land 0.28s ease-out;
}
.game-result {
  margin: 22px auto 0;
  min-height: 24px;
  text-align: center;
  font-size: 15px;
  font-weight: 700;
  color: rgb(107 114 128);
}
.game-result.pos {
  color: rgb(5 150 105);
}
.game-result.neg {
  color: rgb(220 38 38);
}

/* ===== Bet type buttons ===== */
.sicbo-bet-btn {
  position: relative;
  display: flex;
  min-height: 92px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  border-radius: 14px;
  border: 2px solid rgb(229 231 235);
  background: #fff;
  padding: 12px 10px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;
}
.sicbo-bet-btn strong {
  font-size: 19px;
  font-weight: 800;
  color: rgb(17 24 39);
}
.sicbo-bet-btn small {
  font-size: 12px;
  font-weight: 600;
  color: rgb(156 163 175);
}
.sicbo-bet-btn:hover:not(:disabled) {
  border-color: rgb(139 116 248);
  transform: translateY(-1px);
}
.sicbo-bet-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}
.sicbo-bet-btn.is-active {
  border-color: #6d4df6;
  box-shadow: 0 0 0 3px rgba(109, 77, 246, 0.18);
}
.sicbo-bet-btn.is-active strong {
  color: #6d4df6;
}
.sicbo-odds {
  margin-top: 3px;
  padding: 1px 8px;
  border-radius: 999px;
  background: #f1eeff;
  color: #6d4df6;
  font-size: 11px;
  font-weight: 700;
  visibility: hidden;
}
.sicbo-odds.visible {
  visibility: visible;
}
.sicbo-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-top: 12px;
  font-size: 12.5px;
  color: rgb(156 163 175);
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
@media (prefers-reduced-motion: reduce) {
  .sicbo-die.is-tumbling,
  .sicbo-die.is-kept {
    animation: none;
  }
  .sicbo-bet-btn,
  .chip-btn {
    transition: none;
  }
}
</style>
