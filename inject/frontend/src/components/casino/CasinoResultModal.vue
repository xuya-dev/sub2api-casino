<template>
  <transition name="crm-zoom">
    <div
      v-if="open"
      class="fixed inset-0 z-[70] flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      :aria-label="game"
    >
      <div class="absolute inset-0 bg-gray-900/55 backdrop-blur-[2px]" @click="close" />
      <div class="relative w-full max-w-sm overflow-hidden rounded-3xl bg-white shadow-[0_24px_70px_rgba(15,23,42,0.35)] dark:bg-dark-900">
        <!-- 结果横幅：渐变底 + 光斑 + 大金额，底部预留徽章嵌入空间 -->
        <div class="relative overflow-hidden px-6 pb-12 pt-8 text-center" :class="bannerClass">
          <div class="crm-glow crm-glow-a" aria-hidden="true" />
          <div class="crm-glow crm-glow-b" aria-hidden="true" />
          <div class="crm-sparkle absolute inset-0" aria-hidden="true" />
          <p class="relative text-xs font-semibold uppercase tracking-[0.3em] text-white/75">{{ game }}</p>
          <p class="relative mt-2 text-2xl font-black text-white drop-shadow-sm">{{ title }}</p>
          <p
            class="relative mt-1 text-[2.6rem] font-black leading-none tabular-nums tracking-tight text-white drop-shadow-md"
          >
            {{ delta >= 0 ? '+' : '−' }}{{ money(Math.abs(delta)) }}
          </p>
        </div>

        <div class="px-6 pb-6">
          <!-- 结果徽章：负边距压入横幅下缘 -->
          <div class="-mt-9 mb-4 flex justify-center">
            <div
              class="flex h-16 w-16 items-center justify-center rounded-full bg-white text-3xl shadow-lg ring-4 ring-white dark:bg-dark-800 dark:ring-dark-900"
              aria-hidden="true"
            >
              {{ emoji }}
            </div>
          </div>

          <!-- 明细：投注/派奖 双列 + 余额 -->
          <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-800/60">
            <div class="grid grid-cols-2 gap-3">
              <div class="rounded-xl bg-white px-3 py-2.5 text-center shadow-sm dark:bg-dark-900">
                <p class="text-xs text-gray-400 dark:text-dark-500">{{ t('casino.modal.bet') }}</p>
                <p class="mt-0.5 text-base font-bold tabular-nums text-gray-900 dark:text-white">{{ money(bet) }}</p>
              </div>
              <div class="rounded-xl bg-white px-3 py-2.5 text-center shadow-sm dark:bg-dark-900">
                <p class="text-xs text-gray-400 dark:text-dark-500">{{ t('casino.modal.payout') }}</p>
                <p class="mt-0.5 text-base font-bold tabular-nums text-gray-900 dark:text-white">{{ money(payout) }}</p>
              </div>
            </div>
            <div v-if="balance !== null" class="mt-2 flex items-center justify-between rounded-xl bg-white px-3 py-2 shadow-sm dark:bg-dark-900">
              <span class="text-xs text-gray-400 dark:text-dark-500">{{ t('casino.modal.balanceAfter') }}</span>
              <span class="text-sm font-bold tabular-nums text-gray-900 dark:text-white">{{ money(balance) }}</span>
            </div>
          </div>

          <!-- 弹窗内直接改下注（可选） -->
          <div v-if="betOptions && betOptions.length" class="mt-4">
            <p class="mb-1.5 text-xs font-medium text-gray-400 dark:text-dark-500">{{ t('casino.modal.betLabel') }}</p>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="v in betOptions"
                :key="v"
                type="button"
                class="min-w-[3.25rem] rounded-full border px-3 py-1.5 text-sm font-semibold tabular-nums transition"
                :class="Number(betValue) === v
                  ? 'border-primary-500 bg-primary-50 text-primary-600 dark:bg-primary-500/10'
                  : 'border-gray-200 text-gray-500 hover:border-primary-300 dark:border-dark-700 dark:text-dark-300'"
                @click="$emit('update:betValue', v)"
              >
                {{ money(v) }}
              </button>
            </div>
          </div>

          <div class="mt-5 flex gap-3">
            <button
              v-if="canAgain"
              type="button"
              class="btn btn-primary flex-1"
              :disabled="againDisabled"
              @click="$emit('again')"
            >
              {{ t('casino.modal.again') }}
            </button>
            <button type="button" class="btn flex-1" :class="canAgain ? 'btn-secondary' : 'btn-primary'" @click="close">
              {{ t('casino.modal.ok') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
/**
 * CasinoResultModal — 六款游戏统一的回合结果弹窗。
 * win: true 中奖 / false 未中奖 / null 平局退本（和局、1 倍返还）。
 * 可选 betOptions/betValue：弹窗内直接改下注额，配合 again 事件重开一局。
 */
import { computed, onBeforeUnmount, onMounted } from 'vue'
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
    /** 覆盖默认标题（win→中奖啦 / lose→未中奖 / push→平局），用于对局类结果文案 */
    title?: string
    /** 覆盖默认 emoji */
    emoji?: string
    /** 弹窗内可改下注：可选面值列表 + 当前值（update:betValue） */
    betOptions?: number[]
    betValue?: number
  }>(),
  { balance: null, canAgain: false, againDisabled: false, betOptions: undefined, betValue: undefined }
)

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'again'): void
  (e: 'update:betValue', v: number): void
}>()

const { t } = useI18n()

const money = (v: number) =>
  Number(v ?? 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

const bannerClass = computed(() => {
  if (props.win === true) return 'bg-gradient-to-br from-emerald-400 via-emerald-600 to-teal-800'
  if (props.win === false) return 'bg-gradient-to-br from-rose-400 via-red-600 to-rose-900'
  return 'bg-gradient-to-br from-amber-300 via-amber-500 to-orange-700'
})

const emoji = computed(() => props.emoji ?? (props.win === true ? '🎉' : props.win === false ? '🎲' : '🤝'))

const title = computed(() => {
  if (props.title) return props.title
  if (props.win === true) return t('casino.modal.win')
  if (props.win === false) return t('casino.modal.lose')
  return t('casino.modal.push')
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
.crm-glow {
  position: absolute;
  border-radius: 9999px;
  filter: blur(2px);
  background: radial-gradient(circle, rgba(255, 255, 255, 0.32) 0%, rgba(255, 255, 255, 0) 70%);
}

.crm-glow-a {
  width: 220px;
  height: 220px;
  left: -70px;
  top: -90px;
}

.crm-glow-b {
  width: 260px;
  height: 260px;
  right: -90px;
  bottom: -140px;
}

.crm-sparkle {
  background-image: radial-gradient(rgba(255, 255, 255, 0.55) 1px, transparent 1.6px);
  background-size: 20px 20px;
  animation: crm-twinkle 2.4s ease-in-out infinite alternate;
}

@keyframes crm-twinkle {
  from {
    opacity: 0.3;
  }
  to {
    opacity: 0.75;
  }
}

.crm-zoom-enter-active,
.crm-zoom-leave-active {
  transition: opacity 0.18s ease;
}

.crm-zoom-enter-active :deep(.relative),
.crm-zoom-leave-active :deep(.relative) {
  transition: transform 0.18s ease;
}

.crm-zoom-enter-from,
.crm-zoom-leave-to {
  opacity: 0;
}

.crm-zoom-enter-from :deep(.relative),
.crm-zoom-leave-to :deep(.relative) {
  transform: scale(0.92) translateY(10px);
}
</style>
