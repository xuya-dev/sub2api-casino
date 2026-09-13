<template>
  <canvas
    ref="cv"
    class="foil-canvas absolute inset-0 z-10 h-full w-full touch-none select-none"
    :class="disabled ? 'cursor-default' : 'cursor-crosshair'"
    @pointerdown="onDown"
    @pointermove="onMove"
    @pointerup="onUp"
    @pointercancel="onUp"
    @pointerleave="onUp"
  />
</template>

<script setup lang="ts">
/**
 * ScratchFoil 可刮银箔涂层：铺在奖品层之上的单格 canvas。
 *  - 指针拖刮以 destination-out 擦除，刮开面积 ≥ 55% 判定翻开并 emit cleared；
 *  - reveal() 自动擦开（自动刮奖用），reset() 重铺涂层；
 *  - 银箔图案全站共享一份 Image，pattern 按各实例 ctx 独立创建。
 */
import { onBeforeUnmount, onMounted, ref } from 'vue'
import foilUrl from '@/assets/casino/scratch/foil.png'

const props = defineProps<{ disabled?: boolean; hint?: string }>()
const emit = defineEmits<{ (e: 'cleared'): void }>()

const CLEARED_THRESHOLD = 0.55

// 共享银箔图（模块级加载一次）
let sharedFoil: HTMLImageElement | null = null
let sharedReady = false
const foilWaiters: (() => void)[] = []
function loadFoil(): Promise<HTMLImageElement | null> {
  if (sharedReady) return Promise.resolve(sharedFoil)
  return new Promise((resolve) => foilWaiters.push(() => resolve(sharedFoil)))
}
if (typeof window !== 'undefined') {
  const img = new Image()
  img.onload = () => {
    sharedFoil = img
    sharedReady = true
    while (foilWaiters.length) foilWaiters.shift()!()
  }
  img.src = foilUrl
}

const cv = ref<HTMLCanvasElement | null>(null)
let ctx: CanvasRenderingContext2D | null = null
let pattern: CanvasPattern | null = null
let dpr = 1
let painted = false
let done = false
let drawing = false
let lastX = 0
let lastY = 0
let moves = 0
let observer: ResizeObserver | null = null

let resolveReady: (() => void) | null = null
const ready = new Promise<void>((r) => (resolveReady = r))

function paint() {
  const canvas = cv.value
  if (!canvas || !canvas.parentElement) return
  dpr = Math.min(2, window.devicePixelRatio || 1)
  const w = Math.round(canvas.parentElement.clientWidth * dpr)
  const h = Math.round(canvas.parentElement.clientHeight * dpr)
  if (w < 4 || h < 4) return
  // 已翻开的格子只同步尺寸（缩放会把 canvas 清成透明，正合预期），绝不能重铺箔层
  if (done) {
    if (canvas.width !== w) canvas.width = w
    if (canvas.height !== h) canvas.height = h
    return
  }
  if (canvas.width !== w || canvas.height !== h) {
    canvas.width = w
    canvas.height = h
  }
  ctx = canvas.getContext('2d')
  if (!ctx) return
  if (sharedFoil && !pattern) pattern = ctx.createPattern(sharedFoil, 'repeat')
  ctx.globalCompositeOperation = 'source-over'
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  ctx.fillStyle = pattern || '#c3c9d4'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
  const sheen = ctx.createLinearGradient(0, 0, canvas.width, canvas.height)
  sheen.addColorStop(0, 'rgba(255,255,255,0.25)')
  sheen.addColorStop(0.5, 'rgba(255,255,255,0)')
  sheen.addColorStop(1, 'rgba(255,255,255,0.18)')
  ctx.fillStyle = sheen
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  // 大格中央绘制提示文字（小格子空间不足则跳过）
  if (props.hint && canvas.width > 220 * dpr) {
    ctx.fillStyle = 'rgba(100, 88, 130, 0.85)'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.font = `600 ${20 * dpr}px system-ui, sans-serif`
    ctx.fillText(props.hint, canvas.width / 2, canvas.height * 0.5)
  }
  if (!painted) {
    painted = true
    resolveReady?.()
  }
}

function clearedRatio(): number {
  const canvas = cv.value
  if (!canvas || !ctx) return 0
  const step = Math.max(4, Math.round(8 * dpr))
  const data = ctx.getImageData(0, 0, canvas.width, canvas.height).data
  let clear = 0
  let total = 0
  for (let y = 0; y < canvas.height; y += step) {
    for (let x = 0; x < canvas.width; x += step) {
      total++
      if (data[(y * canvas.width + x) * 4 + 3] < 128) clear++
    }
  }
  return total ? clear / total : 0
}

function finish() {
  if (done) return
  done = true
  drawing = false
  if (ctx && cv.value) {
    ctx.globalCompositeOperation = 'destination-out'
    ctx.fillRect(0, 0, cv.value.width, cv.value.height)
  }
  emit('cleared')
}

function onDown(e: PointerEvent) {
  if (done || props.disabled) return
  const canvas = cv.value
  if (!canvas) return
  drawing = true
  const rect = canvas.getBoundingClientRect()
  lastX = ((e.clientX - rect.left) / rect.width) * canvas.width
  lastY = ((e.clientY - rect.top) / rect.height) * canvas.height
  canvas.setPointerCapture?.(e.pointerId)
}

function onMove(e: PointerEvent) {
  if (!drawing || done || props.disabled) return
  const canvas = cv.value
  if (!canvas || !ctx) return
  const rect = canvas.getBoundingClientRect()
  const x = ((e.clientX - rect.left) / rect.width) * canvas.width
  const y = ((e.clientY - rect.top) / rect.height) * canvas.height
  ctx.globalCompositeOperation = 'destination-out'
  ctx.lineWidth = Math.max(canvas.width, canvas.height) * 0.13
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.beginPath()
  ctx.moveTo(lastX, lastY)
  ctx.lineTo(x, y)
  ctx.stroke()
  lastX = x
  lastY = y
  if (++moves % 8 === 0 && clearedRatio() >= CLEARED_THRESHOLD) finish()
}

function onUp() {
  drawing = false
  if (!done && !props.disabled && clearedRatio() >= CLEARED_THRESHOLD) finish()
}

/** 自动擦开涂层（animated=false 时瞬间翻开），完成后 emit cleared。
 *  动画为三道波浪形刮痕依次划过（模拟手指刮涂）；
 *  后台/遮挡页签的 rAF 会被无限节流：首帧 250ms 内未到达则直接翻开。 */
function reveal(animated = true): Promise<void> {
  return ready.then(
    () =>
      new Promise<void>((resolve) => {
        const canvas = cv.value
        if (done || !canvas || !ctx || !animated) {
          finish()
          resolve()
          return
        }
        const duration = 620
        const start = performance.now()
        let started = false
        const fallback = window.setTimeout(() => {
          if (!started && !done) {
            finish()
            resolve()
          }
        }, 250)
        // 三道横向波浪刮痕，按动画进度依次划过
        const strokes = [0.26, 0.5, 0.74].map((fy, idx) => ({
          y0: canvas.height * fy,
          amp: canvas.height * 0.055,
          width: canvas.height * 0.22,
          phase: idx * 1.9
        }))
        const drawStrokes = (p: number) => {
          ctx!.globalCompositeOperation = 'destination-out'
          ctx!.lineCap = 'round'
          ctx!.lineJoin = 'round'
          strokes.forEach((s, k) => {
            const local = Math.min(1, Math.max(0, p * strokes.length - k))
            if (local <= 0) return
            const xEnd = canvas.width * local
            ctx!.lineWidth = s.width
            ctx!.beginPath()
            let startedPath = false
            const steps = 28
            for (let i = 0; i <= steps; i++) {
              const tt = i / steps
              const x = tt * canvas.width
              if (x > xEnd) break
              const y = s.y0 + Math.sin(tt * Math.PI * 3 + s.phase) * s.amp
              if (startedPath) {
                ctx!.lineTo(x, y)
              } else {
                ctx!.moveTo(x, y)
                startedPath = true
              }
            }
            ctx!.stroke()
          })
        }
        const step = (now: number) => {
          if (!started) {
            started = true
            clearTimeout(fallback)
          }
          if (done) {
            resolve()
            return
          }
          const p = Math.min(1, (now - start) / duration)
          drawStrokes(p)
          if (p < 1) {
            requestAnimationFrame(step)
          } else {
            finish()
            resolve()
          }
        }
        requestAnimationFrame(step)
      })
  )
}

/** 重铺涂层（下一张卡复用同组组件时使用）。 */
function reset() {
  done = false
  drawing = false
  moves = 0
  paint()
}

defineExpose({ reveal, reset })

onMounted(async () => {
  await loadFoil()
  paint()
  const parent = cv.value?.parentElement
  if (parent && typeof ResizeObserver !== 'undefined') {
    observer = new ResizeObserver(() => paint())
    observer.observe(parent)
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})
</script>
