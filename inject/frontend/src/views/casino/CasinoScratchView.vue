<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('casino.scratch.title') }}
            <span class="ml-1 align-middle text-xl" aria-hidden="true">🎫</span>
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.scratch.description') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <router-link to="/casino" class="btn btn-secondary">
            {{ t('casino.backToLobby') }}
          </router-link>
          <button type="button" class="btn btn-secondary" @click="showRules = !showRules">
            <Icon name="questionCircle" size="md" />
            {{ t('casino.scratch.howToPlay') }}
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
            <li>{{ t('casino.scratch.ruleBuy') }}</li>
            <li>{{ t('casino.scratch.ruleClassic') }}</li>
            <li>{{ t('casino.scratch.ruleLucky7') }}</li>
            <li>{{ t('casino.scratch.ruleLines') }}</li>
            <li>{{ t('casino.scratch.ruleScratch') }}</li>
            <li>{{ t('casino.scratch.ruleInstant') }}</li>
          </ul>
        </div>
      </transition>

      <!-- Mode tabs -->
      <div class="grid gap-3 sm:grid-cols-3">
        <button
          v-for="m in MODES"
          :key="m"
          type="button"
          class="rounded-xl border-2 p-4 text-left transition disabled:cursor-not-allowed disabled:opacity-60"
          :class="mode === m ? 'border-primary-500 bg-primary-50 shadow-sm' : 'border-gray-200 bg-white hover:border-primary-300 dark:border-dark-700 dark:bg-dark-900'"
          :disabled="playing || buying"
          @click="switchMode(m)"
        >
          <div class="flex items-center justify-between gap-2">
            <span class="text-base font-bold" :class="mode === m ? 'text-primary-600' : 'text-gray-900 dark:text-white'">
              {{ t(modeLabelKey[m]) }}
            </span>
            <span v-if="mode === m" class="text-lg font-black text-primary-600" aria-hidden="true">✓</span>
          </div>
          <div class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">{{ t(modeDescKey[m]) }}</div>
        </button>
      </div>

      <!-- Main grid: stage+controls | aside -->
      <div class="grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div class="flex min-w-0 flex-col gap-6">
          <!-- Main stage -->
          <div
            class="scratch-stage relative h-[460px] overflow-hidden rounded-2xl border-2 border-amber-300/40 shadow-[0_18px_44px_rgba(120,15,30,0.4)]"
            :style="{ backgroundImage: `url(${stageBgUrl})` }"
          >
            <img :src="coinsUrl" alt="" class="pointer-events-none absolute -bottom-4 -left-3 z-20 w-36 select-none drop-shadow-lg" aria-hidden="true" />

            <!-- Promo panel (left) -->
            <div class="promo-panel absolute inset-y-0 left-0 z-10 flex w-[38%] flex-col items-center justify-center gap-4 px-4 text-center">
              <div class="promo-title leading-[1.05]">
                <div class="text-4xl font-black tracking-widest">{{ t('casino.scratch.promoTitle1') }}</div>
                <div class="text-5xl font-black tracking-widest">{{ t('casino.scratch.promoTitle2') }}</div>
              </div>
              <div class="promo-banner -rotate-2 rounded-full px-5 py-1.5 text-sm font-bold text-violet-900 shadow-md">
                {{ t('casino.scratch.promoBanner', { amount: maxPrizeText }) }}
              </div>
              <button type="button" class="mt-1 inline-flex items-center gap-1 text-xs text-white/85 hover:text-white" @click="showRules = !showRules">
                <Icon name="questionCircle" size="xs" />
                {{ t('casino.scratch.howToPlay') }}
              </button>
            </div>

            <!-- Coating area (right) -->
            <div class="absolute inset-y-0 right-0 z-10 w-[62%] p-4 pb-10">
              <div class="coat-frame relative h-full overflow-hidden rounded-xl bg-gradient-to-br from-amber-50 to-orange-50">
                <!-- Idle -->
                <div v-if="!session" class="absolute inset-0 flex flex-col items-center justify-center gap-2">
                  <div class="text-2xl font-semibold text-gray-400">{{ t('casino.scratch.coatingHint1') }}</div>
                  <div class="text-lg text-gray-400">{{ t('casino.scratch.coatingHint2') }}</div>
                </div>
                <!-- Session summary -->
                <div v-else-if="allDone" class="absolute inset-0 flex flex-col items-center justify-center gap-2">
                  <div class="text-lg font-bold text-amber-600">{{ sessionSummary }}</div>
                  <div class="text-sm text-gray-500">{{ t('casino.scratch.againHint') }}</div>
                </div>
                <!-- Cells -->
                <div v-else class="h-full w-full p-3">
                  <div :class="gridClass">
                    <div
                      v-for="(cell, i) in currentCells"
                      :key="`${sessionId}-${cardIdx}-${i}`"
                      class="relative overflow-hidden rounded-xl bg-white shadow-sm transition-shadow"
                      :class="[cellSizeClass, cellPrizeClass(cell, i)]"
                    >
                      <div class="absolute inset-0 flex flex-col items-center justify-center gap-0.5 p-1">
                        <!-- classic: 单格倍数 -->
                        <template v-if="mode === 'classic'">
                          <div class="text-4xl font-black" :class="cell.multiplier ? 'text-amber-500' : 'text-gray-300'">
                            ×{{ fmtMult(cell.multiplier ?? 0) }}
                          </div>
                          <div class="text-xs text-gray-400">
                            {{ cell.multiplier ? '¥ ' + money((cell.multiplier ?? 0) * face) : t('casino.scratch.prizeNone') }}
                          </div>
                        </template>
                        <!-- lucky7: 幸运7 / 水果填充 -->
                        <template v-else-if="mode === 'lucky7'">
                          <template v-if="cell.kind === 'seven'">
                            <img :src="SYMBOL_URLS.seven" alt="" class="w-[52%] select-none" draggable="false" />
                            <span class="rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-bold text-amber-700">
                              ×{{ fmtMult(cell.multiplier ?? 0) }}
                            </span>
                          </template>
                          <img v-else :src="SYMBOL_URLS[dudFor(i)]" alt="" class="w-[46%] select-none opacity-60 grayscale-[40%]" draggable="false" />
                        </template>
                        <!-- lines: 连线符号 -->
                        <template v-else>
                          <img :src="SYMBOL_URLS[cell.symbol ?? 'cherry']" alt="" class="w-[58%] select-none" draggable="false" />
                          <span class="text-[11px] font-semibold" :class="winCellSet.has(i) ? 'text-amber-600' : 'text-gray-400'">
                            ×{{ fmtMult(cell.multiplier ?? 0) }}
                          </span>
                        </template>
                      </div>
                      <ScratchFoil v-show="!cardDone" :ref="(el) => setFoilRef(i, el)" :disabled="!interactive" @cleared="onCellCleared" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Session progress strip -->
            <div class="absolute inset-x-0 bottom-0 z-20 flex items-center justify-between gap-3 bg-[#2b0812]/75 px-4 py-1.5 text-xs text-amber-100 backdrop-blur-sm">
              <span>{{ progressLabel }}</span>
              <span v-if="session && !allDone && cardDone" class="font-semibold" :class="(currentCard?.multiplier ?? 0) > 0 ? 'text-amber-300' : ''">
                {{ cardOutcomeText }}
              </span>
              <span v-else-if="session">{{ t('casino.scratch.sessionWin', { amount: money(session.total_payout), wins: winCount, total: session.count }) }}</span>
            </div>
          </div>

          <!-- Controls -->
          <div class="card p-6">
            <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
              <div>
                <div class="input-label">{{ t('casino.scratch.faceValue') }}</div>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="v in faceOptions"
                    :key="v"
                    type="button"
                    class="min-w-[4rem] rounded-xl border-2 px-4 py-2 text-base font-bold tabular-nums transition"
                    :class="Number(face) === v
                      ? 'border-primary-500 bg-primary-50 text-primary-600 shadow-sm dark:bg-primary-500/10 dark:text-primary-400'
                      : 'border-gray-200 bg-white text-gray-500 hover:border-primary-300 hover:text-gray-700 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300 dark:hover:border-dark-500'"
                    :disabled="playing"
                    @click="face = v"
                  >
                    {{ v }}<span class="ml-0.5 text-xs font-medium opacity-70">{{ t('casino.scratch.yuan') }}</span>
                  </button>
                </div>
              </div>
              <div>
                <div class="input-label">{{ t('casino.scratch.buyCount') }}</div>
                <div class="flex items-center gap-3">
                  <button type="button" class="btn btn-secondary h-9 w-9 p-0 text-lg leading-none" :disabled="playing" @click="count = Math.max(1, count - 1)">−</button>
                  <input v-model.number="count" type="number" min="1" max="10" class="input w-20 text-center" :disabled="playing" />
                  <button type="button" class="btn btn-secondary h-9 w-9 p-0 text-lg leading-none" :disabled="playing" @click="count = Math.min(10, count + 1)">+</button>
                </div>
              </div>
              <div>
                <div class="input-label">{{ t('casino.scratch.autoScratch') }}</div>
                <Toggle v-model="auto" />
              </div>
            </div>
            <div class="mt-5 flex flex-wrap items-stretch gap-3">
              <button type="button" class="btn btn-primary h-[58px] min-w-[220px] flex-col leading-tight" :disabled="playing || buying" @click="startSession()">
                <span class="text-base font-bold">{{ playing ? t('casino.scratch.scratching') : t('casino.scratch.startBtn') }}</span>
                <small class="text-xs font-normal opacity-90">
                  {{ t('casino.scratch.cost', { amount: money(face * count) }) }}
                </small>
              </button>
              <button type="button" class="btn btn-secondary h-[58px] min-w-[132px] flex-col leading-tight" :disabled="playing || buying" @click="startSession(1)">
                <span class="inline-flex items-center gap-1.5 text-base font-bold">
                  <Icon name="refresh" size="sm" />
                  {{ t('casino.scratch.changeOne') }}
                </span>
                <small class="text-xs font-normal opacity-90">{{ t('casino.scratch.cost', { amount: money(face) }) }}</small>
              </button>
              <button
                type="button"
                class="btn h-[58px] min-w-[132px] flex-col leading-tight"
                :class="interactive ? 'btn-primary' : 'btn-secondary'"
                :disabled="!interactive"
                @click="revealAllNow()"
              >
                <span class="inline-flex items-center gap-1.5 text-base font-bold">
                  <Icon name="eye" size="sm" />
                  {{ t('casino.scratch.revealAll') }}
                </span>
                <small class="text-xs font-normal opacity-90">
                  {{ revealing ? t('casino.scratch.scratching') : t('casino.scratch.revealAllHint') }}
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
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.scratch.cardFace') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ money(face) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ modeWinRateLabel }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ modeWinRatePct }}</dd>
              </div>
              <div v-if="mode !== 'lucky7'" class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.scratch.maxMultLabel') }}</dt>
                <dd class="font-semibold text-amber-600">×{{ fmtMult(modeMaxMult) }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.scratch.statusLabel') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ progressLabel }}</dd>
              </div>
              <div class="flex items-center justify-between gap-3">
                <dt class="text-gray-500 dark:text-dark-400">{{ t('casino.scratch.remainingLabel') }}</dt>
                <dd class="font-semibold text-gray-900 dark:text-white">{{ remaining }}</dd>
              </div>
            </dl>
          </section>

          <!-- 连线玩法符号赔率表 -->
          <section v-if="mode === 'lines'" class="card p-5">
            <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.scratch.linesPayTable') }}</h2>
            <div class="grid grid-cols-2 gap-2">
              <div
                v-for="sym in lineSymbols"
                :key="sym.id"
                class="flex items-center gap-2 rounded-lg border border-gray-100 bg-gray-50/60 px-2.5 py-1.5 dark:border-dark-700 dark:bg-dark-800/60"
              >
                <img :src="SYMBOL_URLS[sym.id] ?? SYMBOL_URLS.cherry" alt="" class="h-8 w-8 select-none object-contain" draggable="false" />
                <span class="ml-auto text-sm font-bold text-gray-900 dark:text-white">×{{ fmtMult(sym.multiplier) }}</span>
              </div>
            </div>
          </section>

          <section class="card p-5">
            <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">{{ t('casino.scratch.feedTitle') }}</h2>
            <ul class="divide-y divide-gray-100 dark:divide-dark-800">
              <li v-for="row in scratchFeed" :key="row.id" class="flex items-center justify-between gap-3 py-2.5 text-sm">
                <span class="text-gray-500 dark:text-dark-400">{{ row.created_at }}</span>
                <span class="font-semibold" :class="row.delta >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                  {{ row.delta >= 0 ? '+' : '−' }}{{ money(Math.abs(row.delta)) }}
                </span>
              </li>
              <li v-if="!scratchFeed.length" class="px-3 py-6 text-center text-sm text-gray-400 dark:text-dark-500">
                {{ t('casino.scratch.feedEmpty') }}
              </li>
            </ul>
          </section>
        </div>
      </div>

      <CasinoResultModal
        :open="resultModal"
        :game="t('casino.games.scratch')"
        :win="modalData.win"
        :delta="modalData.delta"
        :bet="modalData.bet"
        :payout="modalData.payout"
        :balance="modalData.balance"
        can-again
        :bet-options="faceOptions"
        :bet-value="Number(face)"
        @update:bet-value="face = $event"
        @close="resultModal = false"
        @again="onPlayAgain"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Toggle from '@/components/common/Toggle.vue'
import ScratchFoil from '@/components/casino/ScratchFoil.vue'
import CasinoStatsBar, { type CasinoStatItem } from '@/components/casino/CasinoStatsBar.vue'
import CasinoResultModal from '@/components/casino/CasinoResultModal.vue'
import { useCasinoData } from '@/components/casino/useCasino'
import {
  casinoAPI,
  type CasinoBetRow,
  type CasinoScratchCell,
  type CasinoScratchMode,
  type CasinoScratchRevealResult
} from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import coinsUrl from '@/assets/casino/scratch/coins.png'
import stageBgUrl from '@/assets/casino/scratch/stage-bg.png'
import symBell from '@/assets/casino/scratch/sym_bell.png'
import symCherry from '@/assets/casino/scratch/sym_cherry.png'
import symCrown from '@/assets/casino/scratch/sym_crown.png'
import symDiamond from '@/assets/casino/scratch/sym_diamond.png'
import symGrape from '@/assets/casino/scratch/sym_grape.png'
import symLemon from '@/assets/casino/scratch/sym_lemon.png'
import symMelon from '@/assets/casino/scratch/sym_melon.png'
import symSeven from '@/assets/casino/scratch/sym_seven.png'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const { meta, loadAll, refreshMe } = useCasinoData()

/* ==================== Constants ==================== */

const FACE_PRESETS = [2, 5, 10, 20, 50]
const MODES: CasinoScratchMode[] = ['classic', 'lucky7', 'lines']
const SYMBOL_URLS: Record<string, string> = {
  bell: symBell,
  cherry: symCherry,
  crown: symCrown,
  diamond: symDiamond,
  grape: symGrape,
  lemon: symLemon,
  melon: symMelon,
  seven: symSeven
}
const DUD_SYMBOLS = ['cherry', 'lemon', 'grape', 'melon']
const DEFAULT_LINE_SYMBOLS = [
  { id: 'cherry', multiplier: 1 },
  { id: 'lemon', multiplier: 1.5 },
  { id: 'grape', multiplier: 2 },
  { id: 'melon', multiplier: 3 },
  { id: 'bell', multiplier: 5 },
  { id: 'diamond', multiplier: 10 },
  { id: 'crown', multiplier: 20 },
  { id: 'seven', multiplier: 88.888 }
]
const modeLabelKey: Record<CasinoScratchMode, string> = {
  classic: 'casino.scratch.modeClassic',
  lucky7: 'casino.scratch.modeLucky7',
  lines: 'casino.scratch.modeLines'
}
const modeDescKey: Record<CasinoScratchMode, string> = {
  classic: 'casino.scratch.modeDescClassic',
  lucky7: 'casino.scratch.modeDescLucky7',
  lines: 'casino.scratch.modeDescLines'
}

interface ScratchMetaInfo {
  win_rate?: number
  max_multiplier?: number
  lucky7?: { cells?: number; hit_rate?: number; max_cells_mult?: number }
  lines?: { win_rate?: number; symbols?: { id: string; multiplier: number }[] }
}

/* ==================== State ==================== */

const face = ref(10)
const count = ref(1)
const auto = ref(false)
const playing = ref(false)
const buying = ref(false)
const showRules = ref(false)
const mode = ref<CasinoScratchMode>('classic')
const session = ref<CasinoScratchRevealResult | null>(null)
const sessionId = ref(0)
const cardIdx = ref(0)
const cardDone = ref(false)
const allDone = ref(false)
const clearedCount = ref(0)
const resultModal = ref(false)
const modalData = ref({ win: null as boolean | null, delta: 0, payout: 0, bet: 0, balance: null as number | null })
const revealing = ref(false)
const scratchFeed = ref<CasinoBetRow[]>([])

const foils = ref<Array<InstanceType<typeof ScratchFoil> | null>>([])
let disposed = false
let cardResolve: (() => void) | null = null

/* ==================== Helpers ==================== */

const money = (v: number) => Number(v ?? 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const fmtMult = (v: number) => (Number.isInteger(v) ? String(v) : String(Math.round(v * 1000) / 1000))
const pct = (v: number) => (Math.round(v * 1000) / 10).toFixed(1) + '%'
const tx = (key: string, params?: Record<string, unknown>) =>
  (t as (k: string, p?: Record<string, unknown>) => string)(key, params)

/* ==================== Derived ==================== */

const scratchInfo = computed<ScratchMetaInfo>(
  () => (meta.value as unknown as { scratch?: ScratchMetaInfo } | null)?.scratch ?? {}
)

const lucky7Cells = computed(() => scratchInfo.value.lucky7?.cells ?? 7)
const lucky7Hit = computed(() => scratchInfo.value.lucky7?.hit_rate ?? 0.085)

const modeWinRate = computed(() => {
  if (mode.value === 'lucky7') {
    // 至少找到一个幸运7的概率 = 1 − (1−hit)^cells
    return 1 - Math.pow(1 - lucky7Hit.value, lucky7Cells.value)
  }
  if (mode.value === 'lines') return scratchInfo.value.lines?.win_rate ?? 0.27
  return scratchInfo.value.win_rate ?? 0.29
})

const modeWinRatePct = computed(() => pct(modeWinRate.value))

const modeWinRateLabel = computed(() => {
  if (mode.value === 'lucky7') return t('casino.scratch.lucky7Hit')
  if (mode.value === 'lines') return t('casino.scratch.linesWinRate')
  return t('casino.scratch.winRateLabel')
})

const modeMaxMult = computed(() => {
  if (mode.value === 'lucky7') return scratchInfo.value.lucky7?.max_cells_mult ?? lucky7Cells.value * 88.888
  if (mode.value === 'lines') {
    const syms = lineSymbols.value
    return syms.reduce((max, s) => Math.max(max, s.multiplier), 0)
  }
  return scratchInfo.value.max_multiplier ?? 88.888
})

const lineSymbols = computed(() => {
  const syms = scratchInfo.value.lines?.symbols
  return syms && syms.length ? syms : DEFAULT_LINE_SYMBOLS
})

const faceOptions = computed(() =>
  FACE_PRESETS.filter((v) => v >= (meta.value?.min_bet ?? 0) && v <= (meta.value?.max_bet ?? Infinity))
)

const maxPrizeText = computed(() => money(modeMaxMult.value * (face.value || 0)))

const remaining = computed(() =>
  session.value ? session.value.count - cardIdx.value - (cardDone.value ? 1 : 0) : count.value
)

const winCount = computed(() => session.value?.win_count ?? 0)
const currentCard = computed(() => session.value?.cards[cardIdx.value] ?? null)
const currentCells = computed<CasinoScratchCell[]>(() => currentCard.value?.cells ?? [])

const interactive = computed(() => playing.value && !buying.value && !cardDone.value && !allDone.value && !revealing.value)

const winCellSet = computed(() => new Set<number>((currentCard.value?.win_lines ?? []).flat()))

const progressLabel = computed(() => {
  if (!session.value) return t('casino.scratch.statusIdle')
  return allDone.value
    ? t('casino.scratch.statusDone')
    : t('casino.scratch.cardProgress', { current: cardIdx.value + 1, total: session.value.count })
})

const sessionSummary = computed(() => {
  if (!session.value) return ''
  return t('casino.scratch.sessionSummary', {
    amount: money(session.value.total_payout),
    wins: session.value.win_count,
    total: session.value.count
  })
})

const cardOutcomeText = computed(() => {
  const mult = currentCard.value?.multiplier ?? 0
  return mult > 0
    ? tx('casino.scratch.prizeWin', { amount: money(mult * face.value) })
    : String(t('casino.scratch.prizeNone'))
})

const gridClass = computed(() => {
  if (mode.value === 'classic') return 'grid h-full w-full grid-cols-1'
  if (mode.value === 'lucky7') return 'flex h-full w-full flex-wrap content-center justify-center gap-2.5'
  return 'grid h-full w-full grid-cols-3 grid-rows-3 gap-2.5'
})

const cellSizeClass = computed(() => (mode.value === 'lucky7' ? 'aspect-square w-[23%]' : ''))

function cellPrizeClass(cell: CasinoScratchCell, i: number): string {
  if (mode.value === 'lines' && cardDone.value && winCellSet.value.has(i)) {
    return 'cell-win ring-2 ring-amber-400'
  }
  if (mode.value === 'lucky7' && cell.kind === 'seven') {
    return 'ring-2 ring-amber-300'
  }
  return 'ring-1 ring-amber-200/70'
}

function dudFor(i: number): string {
  return DUD_SYMBOLS[(cardIdx.value * 2 + i) % DUD_SYMBOLS.length]
}

const statItems = computed<CasinoStatItem[]>(() => [
  {
    label: t('casino.scratch.statsInvested'),
    value: money(session.value ? session.value.total_bet : face.value * count.value),
    icon: 'creditCard'
  },
  {
    label: t('casino.scratch.statsWinnable'),
    value: money(session.value ? session.value.total_payout : 0),
    icon: 'trendingUp'
  },
  {
    label: t('casino.scratch.statsMaxPrize'),
    value: money(modeMaxMult.value * face.value),
    icon: 'gift'
  },
  {
    label: t('casino.scratch.statsWinRate'),
    value: modeWinRatePct.value,
    icon: 'trendingUp'
  }
])

/* ==================== Reveal flow ==================== */

function setFoilRef(i: number, el: unknown) {
  foils.value[i] = (el as InstanceType<typeof ScratchFoil>) || null
}

function onCellCleared() {
  if (!session.value || cardDone.value || revealing.value) return
  clearedCount.value++
  if (clearedCount.value >= currentCells.value.length) completeCard()
}

function completeCard() {
  if (cardDone.value || !session.value) return
  cardDone.value = true
  // 兜底翻开尚未刮完的格子
  foils.value.forEach((f) => f?.reveal(false))
  if (cardResolve) {
    cardResolve()
    cardResolve = null
  }
}

function waitCard(): Promise<void> {
  return new Promise((resolve) => {
    cardResolve = resolve
  })
}

/** 一键刮开（自动模式同用）：按格子逐个快速擦开涂层，全部翻开后结算。 */
async function revealAllNow(): Promise<void> {
  if (!interactive.value || revealing.value) return
  revealing.value = true
  try {
    const list = foils.value.filter((f): f is InstanceType<typeof ScratchFoil> => !!f)
    await Promise.all(
      list.map(
        (f, j) =>
          new Promise<void>((res) => {
            setTimeout(() => {
              f.reveal(true).then(() => res())
            }, j * 110)
          })
      )
    )
  } finally {
    revealing.value = false
  }
  completeCard()
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms))

async function playSession() {
  const cards = session.value?.cards ?? []
  for (let i = 0; i < cards.length; i++) {
    if (disposed) return
    // 新卡：key 带 sessionId/cardIdx 会重挂载全新涂层，此处仅同步状态。
    // 注意不要清空 foils：函数 ref 只在挂载/补丁时回调，清空后不会自动重注册。
    cardIdx.value = i
    cardDone.value = false
    clearedCount.value = 0
    revealing.value = false
    await nextTick()
    if (auto.value) {
      await revealAllNow()
    } else {
      await waitCard()
      completeCard()
    }
    if (disposed) return
    await sleep(auto.value ? 900 : 1400)
  }
  finishSession()
}

function finishSession() {
  allDone.value = true
  playing.value = false
  if (session.value) showResult(session.value.total_payout, session.value.total_bet, session.value.balance)
  void refreshMe()
  void authStore.refreshUser()
  void loadFeed()
}

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
  void startSession()
}

async function startSession(forcedCount?: number) {
  if (playing.value || buying.value) return
  const cnt = Math.max(1, Math.min(10, forcedCount ?? count.value))
  buying.value = true
  try {
    session.value = await casinoAPI.scratchReveal(face.value, cnt, mode.value)
    sessionId.value++
    allDone.value = false
    playing.value = true
    cardDone.value = false
    cardIdx.value = 0
    clearedCount.value = 0
    foils.value = []
    void authStore.refreshUser() // 不阻塞刮奖流程
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || String(error))
    return
  } finally {
    // 出票完成即解除 buying，否则手动刮奖会被 interactive 守卫拦住
    buying.value = false
  }
  await playSession()
}

function switchMode(m: CasinoScratchMode) {
  if (playing.value || buying.value || mode.value === m) return
  mode.value = m
  session.value = null
  allDone.value = false
  cardIdx.value = 0
  cardDone.value = false
}

/* ==================== Feed ==================== */

async function loadFeed() {
  try {
    const { items } = await casinoAPI.getHistory(50)
    if (disposed) return
    scratchFeed.value = (items || []).filter((row) => row.game === 'scratch').slice(0, 6)
  } catch (error) {
    console.error('Failed to load scratch feed', error)
  }
}

/* ==================== Lifecycle ==================== */

onMounted(async () => {
  await loadAll()
  if (meta.value) {
    face.value = Math.min(Math.max(face.value, meta.value.min_bet), meta.value.max_bet)
  }
  void loadFeed()
})

onBeforeUnmount(() => {
  disposed = true
  completeCard()
})
</script>

<style scoped>
/* 红丝绒颁奖夜舞台：幕布追光背景图 + 左侧暗色渐变保证金字可读 */
.scratch-stage {
  background-color: #3d0a14;
  background-size: cover;
  background-position: center;
}

.promo-panel {
  background: linear-gradient(90deg, rgba(43, 7, 16, 0.8) 0%, rgba(43, 7, 16, 0.45) 62%, rgba(43, 7, 16, 0) 100%);
}

.promo-title {
  background: linear-gradient(180deg, #fef3c7 0%, #fcd34d 45%, #f59e0b 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  text-shadow: 0 2px 14px rgba(245, 158, 11, 0.35);
  filter: drop-shadow(0 3px 6px rgba(0, 0, 0, 0.35));
}

.promo-banner {
  background: linear-gradient(180deg, #fde68a 0%, #fbbf24 100%);
}

.coat-frame {
  box-shadow: inset 0 2px 12px rgba(88, 15, 25, 0.28), 0 1px 0 rgba(255, 255, 255, 0.6);
}

.cell-win {
  animation: cell-win-glow 1.2s ease-in-out infinite alternate;
}

@keyframes cell-win-glow {
  from {
    box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.35);
  }
  to {
    box-shadow: 0 0 14px 2px rgba(245, 158, 11, 0.55);
  }
}
</style>
