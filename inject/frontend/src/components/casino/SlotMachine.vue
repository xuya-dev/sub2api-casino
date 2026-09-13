<template>
  <div ref="mountEl" class="casino-slots-mount"></div>
</template>

<script setup lang="ts">
/**
 * SlotMachine — Vue wrapper around the sprite-strip reel renderer.
 * Parent flow: start() when a spin request goes out, stop() with the server
 * reels when the response arrives, cancel() on teardown / failure.
 */
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { Slots } from './renderers'
import type { CasinoSlotSymbol } from '@/api/casino'

const props = defineProps<{
  symbols: CasinoSlotSymbol[]
}>()

const mountEl = ref<HTMLElement | null>(null)

function symbolIds(): string[] {
  return props.symbols.map((s) => s.id)
}

function start() {
  Slots.start()
}

/** Stop the reels on `reels`; resolves once the last reel settles. */
function stop(reels: string[], quick?: boolean): Promise<void> {
  return new Promise((resolve) => {
    Slots.stop(reels, () => resolve(), quick)
  })
}

function cancel() {
  Slots.cancel()
}

onMounted(() => {
  Slots.render(mountEl.value, symbolIds())
})

onUnmounted(() => {
  Slots.cancel()
})

// Re-render the idle reels once the server symbol list arrives (async meta load).
watch(
  () => props.symbols,
  (value) => {
    if (value && value.length) {
      Slots.render(mountEl.value, symbolIds())
    }
  }
)

defineExpose({ start, stop, cancel })
</script>
