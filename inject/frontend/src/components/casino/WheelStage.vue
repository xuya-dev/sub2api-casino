<template>
  <div ref="stageEl" class="casino-wheel-mount" :aria-busy="spinningInternal"></div>
</template>

<script setup lang="ts">
/**
 * WheelStage — Vue wrapper around the canvas wheel renderer.
 * The hub button lives inside the renderer DOM; its click is forwarded to the
 * parent as a `spin` event (the parent validates the bet, calls the API and
 * then invokes spin() with the winning segment index).
 */
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Wheel, wheelSegmentIndex } from './renderers'
import type { CasinoWheelSegment } from '@/api/casino'

const props = withDefaults(
  defineProps<{
    segments: CasinoWheelSegment[]
    bet: number
    disabled?: boolean
  }>(),
  { disabled: false }
)

const emit = defineEmits<{ spin: [] }>()

const { t } = useI18n()

const stageEl = ref<HTMLElement | null>(null)
const spinningInternal = ref(false)

function hubLabels() {
  return {
    label: t('casino.wheel.hubLabel'),
    costPrefix: t('casino.wheel.hubCost')
  }
}

function handleHubClick() {
  if (props.disabled || spinningInternal.value) return
  emit('spin')
}

/** Animate the wheel to `segment` (index or label); resolves when the wheel stops. */
function spin(segment: number | string, quick?: boolean): Promise<void> {
  return new Promise((resolve) => {
    if (!stageEl.value) {
      resolve()
      return
    }
    const index = wheelSegmentIndex(segment, props.segments.length ? props.segments : Wheel.segments)
    spinningInternal.value = true
    Wheel.spinTo(
      index,
      () => {
        spinningInternal.value = false
        resolve()
      },
      quick
    )
  })
}

onMounted(() => {
  Wheel.setBet(props.bet)
  Wheel.render(stageEl.value, props.segments, hubLabels())
  Wheel.setBet(props.bet)
  Wheel.hub?.addEventListener('click', handleHubClick)
  syncHubDisabled()
})

onUnmounted(() => {
  Wheel.hub?.removeEventListener('click', handleHubClick)
  Wheel.dispose()
})

function syncHubDisabled() {
  if (Wheel.hub) {
    Wheel.hub.disabled = props.disabled || spinningInternal.value
  }
}

watch([() => props.disabled, spinningInternal], syncHubDisabled)

watch(
  () => props.bet,
  (value) => Wheel.setBet(value)
)

watch(
  () => props.segments,
  (value) => {
    if (spinningInternal.value) return
    Wheel.render(stageEl.value, value, hubLabels())
    Wheel.hub?.addEventListener('click', handleHubClick)
    syncHubDisabled()
  }
)

defineExpose({ spin })
</script>
