<template>
  <transition name="crm-zoom">
    <div
      v-if="open"
      class="fixed inset-0 z-[70] flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      :aria-label="game"
    >
      <div class="absolute inset-0 bg-gray-900/60 backdrop-blur-[3px]" @click="close" />

      <div class="crm-card relative w-full max-w-sm" :style="themeVars">
        <button type="button" class="crm-close" aria-label="close" @click="close">✕</button>

        <div class="crm-inner">
          <!-- 顶部金线 + 结果区 -->
          <div class="crm-topline" aria-hidden="true" />
          <p class="crm-game">{{ game }}</p>
          <h3 class="crm-title">{{ title }}</h3>
          <p class="crm-amount">{{ delta >= 0 ? '+' : '−' }}{{ money(Math.abs(delta)) }}</p>
          <p class="crm-msg">{{ message }}</p>

          <!-- 明细 -->
          <div class="crm-detail">
            <div class="crm-grid">
              <div class="crm-cell">
                <span class="crm-k">{{ t('casino.modal.bet') }}</span>
                <span class="crm-v">{{ money(bet) }}</span>
              </div>
              <div class="crm-cell">
                <span class="crm-k">{{ t('casino.modal.payout') }}</span>
                <span class="crm-v">{{ money(payout) }}</span>
              </div>
            </div>
            <div v-if="balance !== null" class="crm-cell crm-balance">
              <span class="crm-k">{{ t('casino.modal.balanceAfter') }}</span>
              <span class="crm-v">{{ money(balance) }}</span>
            </div>
          </div>

          <!-- 弹窗内直接改下注（可选） -->
          <div v-if="betOptions && betOptions.length" class="crm-chips-wrap">
            <p class="crm-k crm-chips-lab">{{ t('casino.modal.betLabel') }}</p>
            <div class="crm-chips">
              <button
                v-for="v in betOptions"
                :key="v"
                type="button"
                class="crm-chip"
                :class="{ 'crm-chip-on': Number(betValue) === v }"
                @click="$emit('update:betValue', v)"
              >
                {{ money(v) }}
                <span v-if="Number(betValue) === v" class="crm-chip-check" aria-hidden="true">✓</span>
              </button>
            </div>
          </div>

          <!-- 操作 -->
          <div class="crm-btns">
            <button
              v-if="canAgain"
              type="button"
              class="crm-btn-main"
              :disabled="againDisabled"
              @click="$emit('again')"
            >
              ↻&nbsp; {{ t('casino.modal.again') }}
            </button>
            <button type="button" class="crm-btn" :class="{ 'crm-btn-solid': !canAgain }" @click="close">
              {{ t('casino.modal.ok') }}
            </button>
          </div>

          <p class="crm-luck">— &nbsp;{{ t('casino.modal.bottomLuck') }}&nbsp; —</p>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
/**
 * CasinoResultModal — 六款游戏统一的回合结果弹窗（黑金会所风，浅/深双主题）。
 * win: true 中奖 / false 未中奖 / null 平局退本。
 * title 可覆盖默认标题（对局类结果语义）；betOptions/betValue 支持弹窗内改下注。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(
  defineProps<{
    open: boolean
    game: string
    /** true 中奖 / false 未中奖 / null 平局 */
    win: boolean | null
    delta: number
    bet: number
    payout: number
    balance?: number | null
    canAgain?: boolean
    againDisabled?: boolean
    /** 覆盖默认标题（对局类结果文案） */
    title?: string
    /** 弹窗内可改下注：可选面值列表 + 当前值（update:betValue） */
    betOptions?: number[]
    betValue?: number
  }>(),
  {
    balance: null,
    canAgain: false,
    againDisabled: false,
    title: undefined,
    betOptions: undefined,
    betValue: undefined
  }
)

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'again'): void
  (e: 'update:betValue', v: number): void
}>()

const { t } = useI18n()

const money = (v: number) =>
  Number(v ?? 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

/** 浅/深双主题：监听主站 html.dark（内联变量注入，不依赖 CSS 管线） */
const isDark = ref(false)
let themeObserver: MutationObserver | null = null
const syncTheme = () => {
  isDark.value = document.documentElement.classList.contains('dark')
}
onMounted(() => {
  syncTheme()
  themeObserver = new MutationObserver(syncTheme)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})
onBeforeUnmount(() => themeObserver?.disconnect())

const themeVars = computed<Record<string, string>>(() =>
  isDark.value
    ? {
        '--crm-bg-color': '#0b0b0f',
        '--crm-bg-image': 'linear-gradient(180deg, #1a1a22 0%, #0b0b0f 100%)',
        '--crm-ink': '#f0e6c8',
        '--crm-sub': '#8f8570',
        '--crm-gold': '#e2c06c',
        '--crm-gold-strong': '#e9d29a',
        '--crm-line': 'rgba(212, 175, 90, 0.42)',
        '--crm-line-soft': 'rgba(212, 175, 90, 0.2)',
        '--crm-cell': '#101015',
        '--crm-detail': '#17171d',
        '--crm-btn-main-bg': 'linear-gradient(180deg, #f2dc9e, #c99b3f)',
        '--crm-btn-main-ink': '#1d1608'
      }
    : {
        '--crm-bg-color': '#faf7f0',
        '--crm-bg-image': 'none',
        '--crm-ink': '#26200f',
        '--crm-sub': '#8f8570',
        '--crm-gold': '#a0793a',
        '--crm-gold-strong': '#8a6a26',
        '--crm-line': 'rgba(176, 141, 66, 0.42)',
        '--crm-line-soft': 'rgba(176, 141, 66, 0.22)',
        '--crm-cell': '#ffffff',
        '--crm-detail': '#f3efe4',
        '--crm-btn-main-bg': 'linear-gradient(180deg, #e8cf94, #c99b3f)',
        '--crm-btn-main-ink': '#221a08'
      }
)

const title = computed(() => {
  if (props.title) return props.title
  if (props.win === true) return t('casino.modal.win')
  if (props.win === false) return t('casino.modal.lose')
  return t('casino.modal.push')
})

const message = computed(() => {
  if (props.win === true) return t('casino.modal.winMsg')
  if (props.win === false) return t('casino.modal.loseMsg')
  return t('casino.modal.pushMsg')
})

function close() {
  emit('close')
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open) close()
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<style scoped>
/* ===== 布局（scoped）===== */

.crm-card {
  position: relative;
  width: 100%;
  border-radius: 20px;
  background-color: var(--crm-bg-color);
  background-image: var(--crm-bg-image);
  box-shadow: 0 30px 70px rgba(0, 0, 0, 0.4);
}

.crm-card::before {
  content: '';
  position: absolute;
  inset: 9px;
  border: 1px solid var(--crm-line);
  border-radius: 14px;
  pointer-events: none;
}

.crm-close {
  position: absolute;
  top: 14px;
  right: 14px;
  z-index: 3;
  width: 30px;
  height: 30px;
  border-radius: 999px;
  color: var(--crm-sub);
  border: 1px solid var(--crm-line-soft);
  font-size: 12px;
  font-weight: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.crm-close:hover {
  color: var(--crm-gold);
  border-color: var(--crm-gold);
}

.crm-inner {
  position: relative;
  padding: 34px 26px 24px;
  text-align: center;
}

.crm-topline {
  width: 64px;
  height: 2px;
  margin: 0 auto 18px;
  background: linear-gradient(90deg, transparent, var(--crm-gold), transparent);
}

.crm-game {
  font-size: 11px;
  letter-spacing: 0.42em;
  text-transform: uppercase;
  color: var(--crm-sub);
}

.crm-title {
  margin-top: 10px;
  font-size: 25px;
  font-weight: 700;
  letter-spacing: 0.14em;
  font-family: Georgia, 'Times New Roman', serif;
  color: var(--crm-gold-strong);
}

.crm-amount {
  margin-top: 6px;
  font-size: 48px;
  font-weight: 800;
  line-height: 1.05;
  font-variant-numeric: tabular-nums;
  color: var(--crm-ink);
}

.crm-msg {
  margin-top: 10px;
  font-size: 13px;
  color: var(--crm-sub);
}

/* ===== 明细 ===== */
.crm-detail {
  margin-top: 20px;
  background: var(--crm-detail);
  border: 1px solid var(--crm-line-soft);
  border-radius: 16px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.crm-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.crm-cell {
  background: var(--crm-cell);
  border: 1px solid var(--crm-line-soft);
  border-radius: 12px;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.crm-k {
  font-size: 12px;
  color: var(--crm-sub);
}

.crm-v {
  font-size: 15px;
  font-weight: 800;
  color: var(--crm-ink);
  font-variant-numeric: tabular-nums;
}

.crm-balance .crm-v {
  font-size: 16px;
}

/* ===== 面值 chips ===== */
.crm-chips-wrap {
  margin-top: 16px;
}

.crm-chips-lab {
  margin-bottom: 8px;
}

.crm-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.crm-chip {
  position: relative;
  min-width: 3.4rem;
  border-radius: 999px;
  border: 1px solid var(--crm-line-soft);
  background: transparent;
  color: var(--crm-sub);
  padding: 7px 16px;
  font-size: 13.5px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  transition: border-color 0.15s ease, color 0.15s ease;
}

.crm-chip:hover {
  border-color: var(--crm-line);
  color: var(--crm-gold);
}

.crm-chip-on {
  border-color: var(--crm-gold);
  color: var(--crm-gold-strong);
  font-weight: 800;
}

:global(.dark) .crm-chip-on {
  color: var(--crm-gold);
}

.crm-chip-check {
  position: absolute;
  top: -7px;
  right: -5px;
  width: 16px;
  height: 16px;
  border-radius: 999px;
  background: var(--crm-gold);
  color: #fff;
  font-size: 10px;
  font-weight: 900;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--crm-cell);
}

/* ===== 按钮 ===== */
.crm-btns {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.crm-btn {
  flex: 1;
  border-radius: 999px;
  padding: 12px 0;
  font-size: 15px;
  font-weight: 800;
  color: var(--crm-gold-strong);
  background: transparent;
  border: 1px solid var(--crm-line);
  transition: filter 0.15s ease, transform 0.1s ease;
}

.crm-btn:hover {
  filter: brightness(1.06);
}

.crm-btn:active {
  transform: scale(0.98);
}

.crm-btn-solid,
.crm-btn-main {
  flex: 1;
  border-radius: 999px;
  padding: 12px 0;
  font-size: 15px;
  font-weight: 800;
  background: var(--crm-btn-main-bg);
  color: var(--crm-btn-main-ink);
  border: none;
  box-shadow: 0 6px 16px rgba(201, 155, 63, 0.3);
  transition: filter 0.15s ease, transform 0.1s ease;
}

.crm-btn-main:hover,
.crm-btn-solid:hover {
  filter: brightness(1.05);
}

.crm-btn-main:active,
.crm-btn-solid:active {
  transform: scale(0.98);
}

.crm-btn-main:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.crm-luck {
  margin-top: 16px;
  font-size: 11.5px;
  color: var(--crm-sub);
  opacity: 0.75;
}

/* ===== 进出场 ===== */
.crm-zoom-enter-active,
.crm-zoom-leave-active {
  transition: opacity 0.18s ease;
}

.crm-zoom-enter-active :deep(.crm-card),
.crm-zoom-leave-active :deep(.crm-card) {
  transition: transform 0.18s ease;
}

.crm-zoom-enter-from,
.crm-zoom-leave-to {
  opacity: 0;
}

.crm-zoom-enter-from :deep(.crm-card),
.crm-zoom-leave-to :deep(.crm-card) {
  transform: scale(0.94) translateY(10px);
}
</style>
