<template>
  <span
    class="ref-art"
    :class="[{ 'ref-art-avatar': avatar }, `ref-art-${source}`]"
    :style="{ aspectRatio: `${w} / ${h}` }"
    aria-hidden="true"
  >
    <img
      :src="src"
      alt=""
      draggable="false"
      :style="{
        width: `${(153600 / w).toFixed(4)}%`,
        height: `${(102400 / h).toFixed(4)}%`,
        left: `${(-100 * x / w).toFixed(4)}%`,
        top: `${(-100 * y / h).toFixed(4)}%`
      }"
    />
  </span>
</template>

<script setup lang="ts">
/**
 * CasinoArt — crops an illustration region out of one of the 1536x1024
 * reference JPEGs (same technique as the legacy ui.js art() helper).
 * The image is imported through Vite so it gets a hashed build URL.
 */
import { computed } from 'vue'
import lobbyArt from '@/assets/casino/reference-lobby.jpg'
import wheelArt from '@/assets/casino/reference-wheel.jpg'
import slotsArt from '@/assets/casino/reference-slots.jpg'
import blackjackArt from '@/assets/casino/reference-blackjack.jpg'

const props = withDefaults(
  defineProps<{
    source: 'lobby' | 'wheel' | 'slots' | 'blackjack'
    /** Crop origin / size on the 1536x1024 reference sheet */
    x: number
    y: number
    w: number
    h: number
    avatar?: boolean
  }>(),
  { avatar: false }
)

const SOURCES = {
  lobby: lobbyArt,
  wheel: wheelArt,
  slots: slotsArt,
  blackjack: blackjackArt
} as const

const src = computed(() => SOURCES[props.source])
</script>

<style scoped>
.ref-art {
  position: relative;
  display: block;
  overflow: hidden;
  background: #edebf7;
}
.ref-art img {
  position: absolute;
  display: block;
  max-width: none;
  pointer-events: none;
  user-select: none;
}
.ref-art-avatar {
  width: 34px;
  height: 34px;
  aspect-ratio: auto !important;
  border-radius: 50%;
  flex: none;
}
</style>
