<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('casino.slots.title') }}
            <span class="ml-1 align-middle text-xl" aria-hidden="true">🎰</span>
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.slots.description') }}</p>
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
            <li>{{ t('casino.slots.spinHint') }}</li>
            <li>{{ t('casino.common.betRange', { min: formatBet(meta?.min_bet ?? 0), max: formatBet(meta?.max_bet ?? 0) }) }}</li>
          </ul>
        </div>
      </transition>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div class="flex min-w-0 flex-col gap-6">
          <!-- Machine scene -->
          <div class="slots-scene">
            <div class="slot-machine">
              <CasinoArt source="slots" :x="266" :y="279" :w="914" :h="470" class="machine-art" />
              <div class="jackpot-readout" aria-hidden="true">
                <span>JACKPOT</span>
                <strong>LUCKY SLOTS</strong>
              </div>
              <SlotMachine ref="reels" :symbols="meta?.slots.symbols || []" class="slots-mount" />
            </div>
            <p class="game-result" :class="resultPos ? 'pos' : ''" role="status" aria-live="polite">
              {{ resultText }}
            </p>
          </div>

          <!-- Controls -->
          <div class="card p-6">
            <div class="grid gap-5 md:grid-cols-[1.5fr_1fr]">
              <div>
                <label for="casino-slots-bet" class="input-label">{{ t('casino.slots.betAmount') }}</label>
                <div class="mt-1 flex items-stretch gap-2">
                  <button type="button" class="btn btn-secondary px-3" aria-label="-" :disabled="spinning" @click="stepBet(-10)">−</button>
                  <input
                    id="casino-slots-bet"
                    v-model.number="bet"
                    type="number"
                    class="input text-center"
                    :min="meta?.min_bet ?? 0"
                    :max="meta?.max_bet ?? 0"
                    step="0.01"
                    :disabled="spinning"
                  />
                  <button type="button" class="btn btn-secondary px-3" aria-label="+" :disabled="spinning" @click="stepBet(10)">+</button>
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
                <label class="input-label">{{ t('casino.slots.lineLabel') }}</label>
                <div class="mt-1 flex items-stretch gap-2">
                  <button type="button" class="btn btn-secondary px-3" disabled aria-label="-">−</button>
                  <span class="input flex flex-1 items-center justify-center text-center font-semibold" title="1">1</span>
                  <button type="button" class="btn btn-secondary px-3" disabled aria-label="+">+</button>
                </div>
              </div>
            </div>
            <div class="mt-5 grid gap-3 md:grid-cols-[1.6fr_1fr_1fr]">
              <button
                type="button"
                class="btn btn-primary h-[58px] flex-col leading-tight"
                :disabled="spinning"
                @click="spin(false)"
              >
                <span class="text-base font-bold tracking-widest">
                  {{ spinning ? t('casino.slots.spinning') : t('casino.slots.spin') }}
                </span>
                <small class="text-xs font-normal opacity-90">{{ t('casino.slots.spinHint') }}</small>
              </button>
              <button type="button" class="btn btn-secondary h-[58px]" :disabled="spinning" @click="bet = meta?.max_bet ?? bet">
                {{ t('casino.slots.maxBet') }}
              </button>
              <button
                type="button"
                class="btn h-[58px]"
                :class="quick ? 'btn-primary' : 'btn-secondary'"
                :aria-pressed="quick"
                :disabled="spinning"
                @click="quick = !quick"
              >
                <Icon name="bolt" size="md" />
                {{ t('casino.slots.quickSpin') }}
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
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.slots.betLabel') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ formatBet(bet) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.slots.winLabel') }}</dt>
                <dd class="font-semibold text-amber-600 dark:text-amber-400">
                  {{ lastPayout === null ? '—' : formatBet(lastPayout) }}
                </dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.common.minBet') }} / {{ t('casino.common.maxBet') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">
                  {{ formatBet(meta?.min_bet ?? 0) }} – {{ formatBet(meta?.max_bet ?? 0) }}
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
                v-for="row in slotsFeed"
                :key="row.id"
                class="flex items-center justify-between gap-3 rounded-xl px-3 py-2 text-sm hover:bg-gray-50 dark:hover:bg-dark-800"
              >
                <div class="min-w-0">
                  <p class="font-medium text-gray-900 dark:text-white">{{ t('casino.games.slots') }}</p>
                  <p class="text-xs text-gray-400 dark:text-dark-500">{{ formatDateTime(row.created_at) }}</p>
                </div>
                <span
                  class="font-semibold"
                  :class="row.delta >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'"
                >
                  {{ row.delta >= 0 ? '+' : '−' }}{{ formatBet(Math.abs(row.delta)) }}
                </span>
              </li>
              <li v-if="!slotsFeed.length" class="px-3 py-6 text-center text-sm text-gray-400 dark:text-dark-500">
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
        :game="t('casino.games.slots')"
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
import CasinoArt from '@/components/casino/CasinoArt.vue'
import SlotMachine from '@/components/casino/SlotMachine.vue'
import { formatBet, useCasinoData } from '@/components/casino/useCasino'
import { casinoAPI, type CasinoBetRow } from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { meta, me, loadAll, refreshMe } = useCasinoData()

const reels = ref<InstanceType<typeof SlotMachine> | null>(null)
const bet = ref(10)
const spinning = ref(false)
const resultModal = ref(false)
const modalData = ref({ win: null as boolean | null, delta: 0, payout: 0, bet: 0, balance: null as number | null })
const quick = ref(false)
const showRules = ref(false)
const resultText = ref('')
const resultPos = ref(false)
const lastPayout = ref<number | null>(null)
const slotsFeed = ref<CasinoBetRow[]>([])

const balance = computed(() => me.value?.user.balance ?? authStore.user?.balance ?? 0)

const presets = computed(() => [10, 50, 100, 500, 1000].filter((v) => v <= (meta.value?.max_bet ?? 5000)))

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
    label: t('casino.slots.betLabel'),
    value: formatBet(bet.value),
    icon: 'dollar',
    iconClass: 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
  },
  {
    label: t('casino.slots.winLabel'),
    value: lastPayout.value === null ? '—' : formatBet(lastPayout.value),
    icon: 'trophy',
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

async function spin(maxBet = false) {
  if (spinning.value) return
  const m = meta.value
  if (!m) return
  if (maxBet) bet.value = m.max_bet
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
  resultText.value = t('casino.slots.spinning')
  reels.value?.start()
  try {
    const result = await casinoAPI.spinSlots(value)
    // reel_ids are the sprite keys on the reference sheet; fall back to emoji reels.
    const reelValues = result.reel_ids && result.reel_ids.length === 3 ? result.reel_ids : result.reels
    await reels.value?.stop(reelValues, quick.value)
    lastPayout.value = result.payout
    if (result.multiplier > 0) {
      resultText.value = t('casino.slots.resultWin', {
        amount: formatBet(result.payout),
        multiplier: result.multiplier
      })
      resultPos.value = true
    } else {
      resultText.value = t('casino.slots.resultLose')
    }
    await Promise.all([refreshMe(), authStore.refreshUser()])
    void loadFeed()
    showResult(result.payout, value, result.balance)
  } catch (error) {
    reels.value?.cancel()
    resultText.value = t('casino.slots.stopFailed')
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
    slotsFeed.value = (items || []).filter((row) => row.game === 'slots').slice(0, 6)
  } catch (error) {
    console.error('Failed to load slots feed', error)
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
.slots-scene {
  border-radius: 20px;
  border: 1px solid #e6e0fa;
  background: linear-gradient(180deg, #f3f0fe, #ece7fd);
  padding: 18px;
}
.dark .slots-scene {
  border-color: rgb(48 55 73);
}
.slot-machine {
  position: relative;
  aspect-ratio: 914 / 470;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 14px 30px rgba(70, 40, 160, 0.22);
}
.machine-art {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  background: transparent;
}
.machine-art :deep(img) {
  object-fit: cover;
}
.slots-mount {
  position: absolute;
  left: 9.4%;
  right: 10.3%;
  top: 21.1%;
  bottom: 7.4%;
  width: auto;
  height: auto;
}
.jackpot-readout {
  position: absolute;
  top: 5.6%;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3;
  display: flex;
  align-items: center;
  gap: 16px;
  background: #0d0b07;
  border: 2px solid #4a4130;
  border-radius: 6px;
  padding: 3px 20px;
  box-shadow: 0 3px 8px rgba(0, 0, 0, 0.45);
}
.jackpot-readout span {
  color: #f2b632;
  font-weight: 800;
  font-size: clamp(11px, 1.5vw, 17px);
  letter-spacing: 1px;
  font-style: italic;
}
.jackpot-readout strong {
  color: #ffd23f;
  font-family: Consolas, 'Courier New', monospace;
  font-size: clamp(13px, 1.9vw, 22px);
  letter-spacing: 2px;
  font-variant-numeric: tabular-nums;
}
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
