<template>
  <div class="space-y-6">
    <!-- 娱乐模式总开关 -->
    <div class="card flex items-center justify-between gap-4 p-6">
      <div>
        <h3 class="text-base font-semibold">{{ t('casino.admin.modeTitle') }}</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.admin.modeDesc') }}</p>
      </div>
      <Toggle v-model="form.enabled" />
    </div>

    <!-- 概率与奖励配置 -->
    <div class="card p-6">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-base font-semibold">{{ t('casino.admin.tabConfig') }}</h3>
        <span class="text-sm text-gray-500">RTP {{ rtpWheel }} / {{ rtpSlots }}</span>
      </div>

      <h4 class="mb-3 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.limits') }}</h4>
      <div class="grid gap-4 sm:grid-cols-3">
        <label class="block">
          <span class="input-label">{{ t('casino.admin.minBet') }}</span>
          <input v-model.number="form.min_bet" type="number" step="0.01" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.maxBet') }}</span>
          <input v-model.number="form.max_bet" type="number" step="0.01" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.dailyLoss') }}</span>
          <input v-model.number="form.daily_loss_limit" type="number" step="1" class="input" />
        </label>
      </div>

      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.wheelTitle') }}</h4>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 text-left text-gray-500 dark:border-dark-700">
              <th class="px-2 py-2">{{ t('casino.admin.label') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.multiplier') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.weight') }}</th>
              <th class="px-2 py-2"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(seg, i) in form.wheel.segments" :key="i" class="border-b border-gray-50 dark:border-dark-800">
              <td class="px-2 py-2"><input v-model="seg.label" type="text" class="input" /></td>
              <td class="px-2 py-2"><input v-model.number="seg.multiplier" type="number" step="0.01" class="input w-28" /></td>
              <td class="px-2 py-2"><input v-model.number="seg.weight" type="number" step="1" class="input w-24" /></td>
              <td class="px-2 py-2 text-right">
                <button type="button" class="text-red-500 hover:text-red-600" @click="form.wheel.segments.splice(i, 1)">✕</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <button type="button" class="btn btn-ghost btn-sm mt-3" @click="form.wheel.segments.push({ label: '1x', multiplier: 1, weight: 5 })">
        + {{ t('casino.admin.addRow') }}
      </button>

      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.slotsTitle') }}</h4>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 text-left text-gray-500 dark:border-dark-700">
              <th class="px-2 py-2">ID</th>
              <th class="px-2 py-2">{{ t('casino.admin.emoji') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.weight') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.triplePay') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.pairPay') }}</th>
              <th class="px-2 py-2"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(sym, i) in form.slots.symbols" :key="i" class="border-b border-gray-50 dark:border-dark-800">
              <td class="px-2 py-2"><input v-model="sym.id" type="text" class="input w-28" /></td>
              <td class="px-2 py-2"><input v-model="sym.emoji" type="text" class="input w-16 text-center" /></td>
              <td class="px-2 py-2"><input v-model.number="sym.weight" type="number" step="1" class="input w-24" /></td>
              <td class="px-2 py-2"><input :value="slotPays[sym.id]?.triple ?? 0" @input="setPay(sym.id, 'triple', $event)" type="number" step="0.1" class="input w-24" /></td>
              <td class="px-2 py-2"><input :value="slotPays[sym.id]?.pair ?? 0" @input="setPay(sym.id, 'pair', $event)" type="number" step="0.1" class="input w-24" /></td>
              <td class="px-2 py-2 text-right">
                <button type="button" class="text-red-500 hover:text-red-600" @click="form.slots.symbols.splice(i, 1)">✕</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <button type="button" class="btn btn-ghost btn-sm mt-3" @click="addSymbol">+ {{ t('casino.admin.addRow') }}</button>

      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.bjTitle') }}</h4>
      <div class="grid gap-4 sm:grid-cols-3">
        <label class="block">
          <span class="input-label">{{ t('casino.admin.bjPays') }}</span>
          <input v-model.number="form.blackjack.blackjack_pays" type="number" step="0.1" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.soft17') }}</span>
          <select v-model="form.blackjack.dealer_stands_soft17" class="input">
            <option :value="true">{{ t('casino.admin.s17stand') }}</option>
            <option :value="false">{{ t('casino.admin.s17hit') }}</option>
          </select>
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.double') }}</span>
          <select v-model="form.blackjack.double_allowed" class="input">
            <option :value="true">{{ t('casino.admin.allowed') }}</option>
            <option :value="false">{{ t('casino.admin.forbidden') }}</option>
          </select>
        </label>
      </div>

      <!-- 骰宝 -->
      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.sicboTitle') }}</h4>
      <div class="grid gap-4 sm:grid-cols-2">
        <label class="block">
          <span class="input-label">{{ t('casino.admin.sicboBig') }}</span>
          <input v-model.number="form.sicbo.big" type="number" step="0.05" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.sicboSmall') }}</span>
          <input v-model.number="form.sicbo.small" type="number" step="0.05" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.sicboOdd') }}</span>
          <input v-model.number="form.sicbo.odd" type="number" step="0.05" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.sicboEven') }}</span>
          <input v-model.number="form.sicbo.even" type="number" step="0.05" class="input" />
        </label>
      </div>

      <!-- 百家乐 -->
      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.baccaratTitle') }}</h4>
      <div class="grid gap-4 sm:grid-cols-3">
        <label class="block">
          <span class="input-label">{{ t('casino.admin.baccaratPlayer') }}</span>
          <input v-model.number="form.baccarat.player" type="number" step="0.05" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.baccaratBanker') }}</span>
          <input v-model.number="form.baccarat.banker" type="number" step="0.05" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.baccaratTie') }}</span>
          <input v-model.number="form.baccarat.tie" type="number" step="0.05" class="input" />
        </label>
      </div>

      <!-- 刮刮乐 · 经典刮奖 -->
      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.scratchTitle') }}</h4>
      <div class="grid gap-4 sm:grid-cols-3">
        <label class="block">
          <span class="input-label">{{ t('casino.admin.scratchWinRate') }}</span>
          <input v-model.number="form.scratch.win_rate" type="number" step="0.01" min="0.01" max="1" class="input" />
        </label>
      </div>
      <div class="mt-3 overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 text-left text-gray-500 dark:border-dark-700">
              <th class="px-2 py-2">{{ t('casino.admin.tierMultiplier') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.tierProb') }}</th>
              <th class="px-2 py-2"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(tier, i) in form.scratch.tiers" :key="i" class="border-b border-gray-50 dark:border-dark-800">
              <td class="px-2 py-2"><input v-model.number="tier.multiplier" type="number" step="0.001" class="input w-28" /></td>
              <td class="px-2 py-2"><input v-model.number="tier.probability" type="number" step="0.001" class="input w-28" /></td>
              <td class="px-2 py-2 text-right">
                <button type="button" class="text-red-500 hover:text-red-600" @click="form.scratch.tiers.splice(i, 1)">✕</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <button type="button" class="btn btn-ghost btn-sm mt-3" @click="form.scratch.tiers.push({ multiplier: 1, probability: 0.05 })">
        + {{ t('casino.admin.addRow') }}
      </button>

      <!-- 刮刮乐 · 幸运7 -->
      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.scratchLucky7Title') }}</h4>
      <div class="grid gap-4 sm:grid-cols-2">
        <label class="block">
          <span class="input-label">{{ t('casino.admin.lucky7Cells') }}</span>
          <input v-model.number="form.scratch.lucky7.cells" type="number" step="1" min="1" max="12" class="input" />
        </label>
        <label class="block">
          <span class="input-label">{{ t('casino.admin.lucky7HitRate') }}</span>
          <input v-model.number="form.scratch.lucky7.hit_rate" type="number" step="0.005" min="0.001" max="1" class="input" />
        </label>
      </div>

      <!-- 刮刮乐 · 幸运连线 -->
      <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.scratchLinesTitle') }}</h4>
      <div class="grid gap-4 sm:grid-cols-3">
        <label class="block">
          <span class="input-label">{{ t('casino.admin.linesWinRate') }}</span>
          <input v-model.number="form.scratch.lines.win_rate" type="number" step="0.01" min="0.01" max="1" class="input" />
        </label>
      </div>
      <div class="mt-3 overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 text-left text-gray-500 dark:border-dark-700">
              <th class="px-2 py-2">{{ t('casino.admin.lineSymbols') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.lineMult') }}</th>
              <th class="px-2 py-2">{{ t('casino.admin.lineProb') }}</th>
              <th class="px-2 py-2"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(sym, i) in form.scratch.lines.symbols" :key="i" class="border-b border-gray-50 dark:border-dark-800">
              <td class="px-2 py-2"><input v-model="sym.id" type="text" class="input w-28" /></td>
              <td class="px-2 py-2"><input v-model.number="sym.multiplier" type="number" step="0.001" class="input w-24" /></td>
              <td class="px-2 py-2"><input v-model.number="sym.probability" type="number" step="0.001" class="input w-24" /></td>
              <td class="px-2 py-2 text-right">
                <button type="button" class="text-red-500 hover:text-red-600" @click="form.scratch.lines.symbols.splice(i, 1)">✕</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <button type="button" class="btn btn-ghost btn-sm mt-3" @click="form.scratch.lines.symbols.push({ id: 'sym' + (form.scratch.lines.symbols.length + 1), multiplier: 1, probability: 0.1 })">
        + {{ t('casino.admin.addRow') }}
      </button>

      <div class="mt-6 flex items-center gap-3">
        <button type="button" class="btn btn-primary" :disabled="saving" @click="save">
          {{ saving ? t('common.saving') : t('casino.admin.save') }}
        </button>
        <span v-if="saveMsg" class="text-sm" :class="saveOk ? 'text-emerald-600' : 'text-red-600'">{{ saveMsg }}</span>
      </div>
    </div>

    <!-- 盈亏统计 -->
    <div class="card p-6">
      <h3 class="mb-4 text-base font-semibold">{{ t('casino.admin.tabStats') }}</h3>
      <div v-if="!stats" class="py-6 text-center text-gray-400">{{ t('common.loading') }}</div>
      <template v-else>
        <div class="grid gap-4 sm:grid-cols-3 xl:grid-cols-6">
          <div v-for="card in statCards" :key="card.label" class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
            <div class="text-xs text-gray-500 dark:text-dark-400">{{ card.label }}</div>
            <div class="mt-1 text-lg font-bold" :class="card.cls">{{ card.value }}</div>
          </div>
        </div>
        <h4 class="mb-3 mt-6 text-sm font-semibold text-gray-700 dark:text-dark-300">{{ t('casino.admin.byGame') }}</h4>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 text-left text-gray-500 dark:border-dark-700">
                <th class="px-3 py-2">{{ t('casino.history.game') }}</th>
                <th class="px-3 py-2">{{ t('casino.admin.rounds') }}</th>
                <th class="px-3 py-2">{{ t('casino.history.bet') }}</th>
                <th class="px-3 py-2">{{ t('casino.history.payout') }}</th>
                <th class="px-3 py-2">{{ t('casino.admin.margin') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(g, i) in stats.games" :key="i" class="border-b border-gray-50 dark:border-dark-800">
                <td class="px-3 py-2 font-medium">{{ gameName(String(g.game ?? '')) }}</td>
                <td class="px-3 py-2">{{ g.rounds }}</td>
                <td class="px-3 py-2">{{ num(g.bet) }}</td>
                <td class="px-3 py-2">{{ num(g.payout) }}</td>
                <td class="px-3 py-2" :class="Number(g.house_margin) >= 0 ? 'text-emerald-600' : 'text-red-600'">
                  {{ (Number(g.house_margin) * 100).toFixed(1) }}%
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import { casinoAPI } from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const loadError = ref('')
const saving = ref(false)
const saveMsg = ref('')
const saveOk = ref(false)
const rtp = ref<Record<string, unknown>>({})
const stats = ref<{ summary: Record<string, unknown>; games: Record<string, unknown>[] } | null>(null)

const form = reactive({
  enabled: true,
  min_bet: 0,
  max_bet: 0,
  daily_loss_limit: 0,
  wheel: { segments: [] as { label: string; multiplier: number; weight: number }[] },
  slots: { symbols: [] as { id: string; emoji: string; weight: number }[] },
  blackjack: { blackjack_pays: 2.5, dealer_stands_soft17: true, double_allowed: true },
  sicbo: { big: 2, small: 2, odd: 2, even: 2 },
  baccarat: { player: 2, banker: 1.95, tie: 9 },
  scratch: {
    win_rate: 0.29,
    tiers: [] as { multiplier: number; probability: number }[],
    lucky7: { cells: 7, hit_rate: 0.085, tiers: [] as { multiplier: number; probability: number }[] },
    lines: { win_rate: 0.27, symbols: [] as { id: string; multiplier: number; probability: number }[] }
  }
})
const slotPays = reactive<Record<string, { triple: number; pair: number }>>({})

const rtpWheel = computed(() => (Number(rtp.value.wheel ?? 0) * 100).toFixed(1) + '%')
const rtpSlots = computed(() => (Number(rtp.value.slots ?? 0) * 100).toFixed(1) + '%')

const num = (v: unknown) => Number(v ?? 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const gameName = (g: string) =>
  ({ wheel: t('casino.admin.gameWheel'), slots: t('casino.admin.gameSlots'), blackjack: t('casino.admin.gameBJ'), sicbo: t('casino.admin.gameSicbo'), baccarat: t('casino.admin.gameBaccarat') })[g] || g

const statCards = computed(() => {
  const s = stats.value?.summary ?? {}
  const profit = Number(s.profit ?? 0)
  const today = Number(s.today_profit ?? 0)
  return [
    { label: t('casino.admin.totalBets'), value: String(s.total_bets ?? 0), cls: '' },
    { label: t('casino.admin.players'), value: String(s.players ?? 0), cls: '' },
    { label: t('casino.admin.totalBet'), value: num(s.total_bet), cls: '' },
    { label: t('casino.admin.totalPayout'), value: num(s.total_payout), cls: '' },
    { label: t('casino.admin.profit'), value: num(profit), cls: profit >= 0 ? 'text-emerald-600' : 'text-red-600' },
    { label: t('casino.admin.todayProfit'), value: num(today), cls: today >= 0 ? 'text-emerald-600' : 'text-red-600' }
  ]
})

function setPay(id: string, key: 'triple' | 'pair', e: Event) {
  if (!slotPays[id]) slotPays[id] = { triple: 0, pair: 0 }
  slotPays[id][key] = Number((e.target as HTMLInputElement).value) || 0
}

function addSymbol() {
  const id = 'sym' + (form.slots.symbols.length + 1)
  form.slots.symbols.push({ id, emoji: '🍀', weight: 5 })
  slotPays[id] = { triple: 0, pair: 0 }
}

async function load() {
  loadError.value = ''
  try {
    const r = await casinoAPI.getAdminConfig()
    const cfg = r.config as unknown as {
      slots: { triple_pays?: Record<string, number>; pair_pays?: Record<string, number> }
    }
    form.enabled = r.config.enabled !== false
    form.min_bet = r.config.min_bet
    form.max_bet = r.config.max_bet
    form.daily_loss_limit = r.config.daily_loss_limit
    form.wheel.segments = r.config.wheel.segments.map(x => ({
      label: x.label, multiplier: Number(x.multiplier), weight: Number(x.weight ?? 1)
    }))
    form.slots.symbols = r.config.slots.symbols.map(x => ({
      id: x.id, emoji: x.emoji, weight: Number(x.weight ?? 1)
    }))
    form.blackjack = { ...r.config.blackjack }
    const cfg = r.config as unknown as Record<string, unknown>
    form.sicbo = { ...(cfg.sicbo as typeof form.sicbo) }
    form.baccarat = { ...(cfg.baccarat as typeof form.baccarat) }
    const sc = cfg.scratch as typeof form.scratch | undefined
    if (sc) {
      form.scratch.win_rate = sc.win_rate
      form.scratch.tiers = sc.tiers?.map(x => ({ multiplier: Number(x.multiplier), probability: Number(x.probability) })) ?? []
      if (sc.lucky7) form.scratch.lucky7 = { ...sc.lucky7 }
      if (sc.lines) form.scratch.lines = { ...sc.lines, symbols: sc.lines.symbols?.map(x => ({ id: x.id, multiplier: Number(x.multiplier), probability: Number(x.probability) })) ?? [] }
    }
    Object.keys(slotPays).forEach(k => delete slotPays[k])
    form.slots.symbols.forEach(x => {
      slotPays[x.id] = { triple: Number(cfg.slots.triple_pays?.[x.id] ?? 0), pair: Number(cfg.slots.pair_pays?.[x.id] ?? 0) }
    })
    rtp.value = r.rtp
    stats.value = await casinoAPI.getAdminStats()
  } catch (e) {
    loadError.value = (e as { message?: string })?.message || String(e)
  }
}

async function save() {
  saving.value = true
  saveMsg.value = ''
  try {
    const triple: Record<string, number> = {}
    const pair: Record<string, number> = {}
    form.slots.symbols.forEach(x => {
      triple[x.id] = slotPays[x.id]?.triple ?? 0
      pair[x.id] = slotPays[x.id]?.pair ?? 0
    })
    await casinoAPI.updateAdminConfig({
      enabled: form.enabled,
      min_bet: form.min_bet,
      max_bet: form.max_bet,
      daily_loss_limit: form.daily_loss_limit,
      wheel: { segments: form.wheel.segments },
      slots: { symbols: form.slots.symbols, triple_pays: triple, pair_pays: pair },
      blackjack: { ...form.blackjack },
      sicbo: { ...form.sicbo },
      baccarat: { ...form.baccarat },
      scratch: { ...form.scratch }
    })
    saveOk.value = true
    saveMsg.value = t('casino.admin.saved')
    appStore.showSuccess(t('casino.admin.saved'))
    await load()
  } catch (e) {
    saveOk.value = false
    saveMsg.value = (e as { message?: string })?.message || String(e)
    appStore.showError(saveMsg.value)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  if (authStore.user?.role !== 'admin') {
    loadError.value = t('casino.admin.needAdmin')
    return
  }
  load()
})
</script>
