<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('casino.wheel.title') }}
            <span class="ml-1 align-middle text-xl" aria-hidden="true">🎡</span>
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.wheel.description') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <router-link to="/casino" class="btn btn-secondary">
            {{ t('casino.backToLobby') }}
          </router-link>
          <button type="button" class="btn btn-secondary" @click="showRules = !showRules">
          <Icon name="questionCircle" size="md" />
          {{ t('casino.wheel.segments') }}
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

      <!-- Prize table -->
      <transition name="fade">
        <div v-if="showRules && meta" class="card p-6">
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(segment, i) in meta.wheel.segments"
              :key="i"
              class="inline-flex items-center gap-2 rounded-lg bg-primary-50 px-3 py-1.5 text-sm dark:bg-primary-900/20"
            >
              <span class="font-medium text-gray-900 dark:text-white">{{ segment.label }}</span>
              <b class="text-primary-600 dark:text-primary-400">{{ segment.multiplier }}x</b>
            </span>
          </div>
        </div>
      </transition>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div class="flex min-w-0 flex-col gap-6">
          <!-- Wheel stage -->
          <div class="relative">
            <WheelStage ref="stage" :segments="meta?.wheel.segments || []" :bet="bet" :disabled="spinning" />
            <p
              class="game-result"
              :class="resultPos ? 'pos' : ''"
              role="status"
              aria-live="polite"
            >
              {{ resultText }}
            </p>
          </div>

          <!-- Controls -->
          <div class="card p-6">
            <div class="grid gap-5 md:grid-cols-[1fr_1fr_auto] md:items-end">
              <div>
                <label for="casino-wheel-bet" class="input-label">{{ t('casino.wheel.betAmount') }}</label>
                <div class="mt-1 flex items-stretch gap-2">
                  <button type="button" class="btn btn-secondary px-3" aria-label="-" :disabled="spinning" @click="stepBet(-10)">
                    −
                  </button>
                  <input
                    id="casino-wheel-bet"
                    v-model.number="bet"
                    type="number"
                    class="input text-center"
                    :min="meta?.min_bet ?? 0"
                    :max="meta?.max_bet ?? 0"
                    step="0.01"
                    :disabled="spinning"
                  />
                  <button type="button" class="btn btn-secondary px-3" aria-label="+" :disabled="spinning" @click="stepBet(10)">
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
                    :disabled="spinning"
                    @click="bet = preset"
                  >
                    {{ preset }}
                  </button>
                </div>
              </div>
              <div>
                <label for="casino-wheel-anim" class="input-label">{{ t('casino.wheel.animation') }}</label>
                <select id="casino-wheel-anim" v-model="quick" class="input mt-1">
                  <option :value="false">{{ t('casino.wheel.animNormal') }}</option>
                  <option :value="true">{{ t('casino.wheel.animQuick') }}</option>
                </select>
              </div>
              <button type="button" class="btn btn-primary h-[58px] min-w-[180px] flex-col leading-tight" :disabled="spinning" @click="spin">
                <span class="text-base font-bold">
                  {{ spinning ? t('casino.wheel.spinning') : t('casino.wheel.startSpin') }}
                </span>
                <small class="text-xs font-normal opacity-90">
                  {{ t('casino.wheel.hubCost') }} {{ formatBet(bet) }}
                </small>
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
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.wheel.perCost') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ formatBet(bet) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.common.minBet') }} / {{ t('casino.common.maxBet') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">
                  {{ formatBet(meta?.min_bet ?? 0) }} – {{ formatBet(meta?.max_bet ?? 0) }}
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.wheel.segments') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ meta?.wheel.segments.length ?? '—' }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.stats.gameRecords') }}</dt>
                <dd class="font-semibold" :class="statusClass">{{ statusText }}</dd>
              </div>
            </dl>
          </section>

          <section class="card">
            <div class="flex items-center justify-between border-b border-gray-100 px-5 py-3.5 dark:border-dark-700">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.wheel.winRecords') }}</h2>
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
                v-for="row in wheelFeed"
                :key="row.id"
                class="flex items-center justify-between gap-3 rounded-xl px-3 py-2 text-sm hover:bg-gray-50 dark:hover:bg-dark-800"
              >
                <div class="min-w-0">
                  <p class="font-medium text-gray-900 dark:text-white">{{ t('casino.games.wheel') }}</p>
                  <p class="text-xs text-gray-400 dark:text-dark-500">{{ formatDateTime(row.created_at) }}</p>
                </div>
                <span class="font-semibold" :class="row.delta >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                  {{ row.delta >= 0 ? '+' : '−' }}{{ formatBet(Math.abs(row.delta)) }}
                </span>
              </li>
              <li v-if="!wheelFeed.length" class="px-3 py-6 text-center text-sm text-gray-400 dark:text-dark-500">
                {{ t('casino.wheel.noWins') }}
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
        :game="t('casino.games.wheel')"
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
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import CasinoStatsBar, { type CasinoStatItem } from '@/components/casino/CasinoStatsBar.vue'
import CasinoResultModal from '@/components/casino/CasinoResultModal.vue'
import WheelStage from '@/components/casino/WheelStage.vue'
import { formatBet, useCasinoData } from '@/components/casino/useCasino'
import { casinoAPI, type CasinoBetRow } from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { meta, me, loadAll, refreshMe } = useCasinoData()

const stage = ref<InstanceType<typeof WheelStage> | null>(null)
const bet = ref(10)
const spinning = ref(false)
const resultModal = ref(false)
const modalData = ref({ win: null as boolean | null, delta: 0, payout: 0, bet: 0, balance: null as number | null })
const quick = ref(false)
const showRules = ref(false)
const resultText = ref('')
const resultPos = ref(false)
const lastCost = ref<number | null>(null)
const lastPayout = ref<number | null>(null)
const wheelFeed = ref<CasinoBetRow[]>([])

const balance = computed(() => me.value?.user.balance ?? authStore.user?.balance ?? 0)

const presets = computed(() =>
  [10, 20, 50, 100, 200].filter((v) => v <= (meta.value?.max_bet ?? 5000))
)

const statItems = computed<CasinoStatItem[]>(() => [
  {
    label: t('casino.stats.balance'),
    value: `$ ${formatBet(balance.value)}`,
    icon: 'creditCard',
    iconClass: 'bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
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
  },
  {
    label: t('casino.wheel.perCost'),
    value: formatBet(lastCost.value ?? bet.value),
    icon: 'dollar',
    iconClass: 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
  },
  {
    label: t('casino.wheel.reward'),
    value: lastPayout.value === null ? '—' : formatBet(lastPayout.value),
    icon: 'gift',
    iconClass: 'bg-violet-50 text-violet-600 dark:bg-violet-900/30 dark:text-violet-400',
    valueClass: 'text-amber-600 dark:text-amber-400'
  }
])

const statusText = computed(() => {
  if (spinning.value) return t('casino.status.playing')
  return lastPayout.value === null ? t('casino.status.idle') : t('casino.status.settled')
})
const statusClass = computed(() =>
  spinning.value ? 'text-primary-600 dark:text-primary-400' : 'text-gray-900 dark:text-white'
)

function stepBet(delta: number) {
  const next = Number(bet.value) + delta
  if (Number.isFinite(next) && next > 0) bet.value = Number(next.toFixed(2))
}

function errorMessage(error: unknown): string {
  const err = error as { message?: string }
  return err?.message || t('casino.common.loadFailed')
}

async function spin() {
  if (spinning.value) return
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
  spinning.value = true
  resultPos.value = false
  resultText.value = t('casino.wheel.spinning')
  lastCost.value = value
  try {
    const result = await casinoAPI.spinWheel(value)
    await stage.value?.spin(result.segment, quick.value)
    lastPayout.value = result.payout
    if (result.multiplier > 0) {
      resultText.value = t('casino.wheel.resultWin', {
        amount: formatBet(result.payout),
        multiplier: result.multiplier
      })
      resultPos.value = true
    } else {
      resultText.value = t('casino.wheel.resultLose')
    }
    await Promise.all([refreshMe(), authStore.refreshUser()])
    void loadFeed()
    showResult(result.payout, value, result.balance)
  } catch (error) {
    resultText.value = t('casino.wheel.stopFailed')
    appStore.showError(errorMessage(error))
  } finally {
    spinning.value = false
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
  void spin()
}

async function loadFeed() {
  try {
    const { items } = await casinoAPI.getHistory(50)
    wheelFeed.value = (items || []).filter((row) => row.game === 'wheel').slice(0, 6)
  } catch (error) {
    console.error('Failed to load wheel feed', error)
  }
}

onMounted(async () => {
  await loadAll()
  if (meta.value) {
    bet.value = Math.min(Math.max(bet.value, meta.value.min_bet), meta.value.max_bet)
  }
  void loadFeed()
})
</script>

<style scoped>
.game-result {
  margin: 12px auto 0;
  min-height: 24px;
  text-align: center;
  font-size: 15px;
  font-weight: 700;
  color: rgb(107 114 128);
}
.game-result.pos {
  color: rgb(5 150 105);
}
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
</style>
