<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('casino.lobby.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.lobby.description') }}</p>
        </div>
          <div class="card flex items-center gap-4 px-5 py-3">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('casino.lobby.balance') }}</p>
            <p class="text-xl font-bold text-primary-600 dark:text-primary-400">$ {{ balanceText }}</p>
          </div>
          <router-link to="/purchase" class="btn btn-primary">{{ t('casino.lobby.recharge') }}</router-link>
        </div>
      </div>

      <!-- Hero banner -->
      <section class="casino-hero">
        <div class="casino-hero-art">
          <CasinoArt source="lobby" :x="600" :y="171" :w="913" :h="212" class="h-full w-full" />
        </div>
        <div class="casino-hero-copy">
          <h2>{{ t('casino.lobby.heroTitle') }}</h2>
          <p>{{ t('casino.lobby.heroSubtitle') }}</p>
          <router-link to="/casino/slots" class="casino-hero-cta">{{ t('casino.lobby.heroCta') }}</router-link>
        </div>
      </section>

      <!-- Feature games -->
      <section class="grid gap-4 md:grid-cols-3">
        <router-link
          v-for="feature in features"
          :key="feature.key"
          :to="feature.to"
          class="feature-card"
          :class="`feature-${feature.key}`"
        >
          <div class="feature-copy">
            <h2>{{ t(`casino.games.${feature.key}`) }}</h2>
            <h3>{{ t(`casino.lobby.feature.${feature.key}.sub`) }}</h3>
            <p>{{ t(`casino.lobby.feature.${feature.key}.desc`) }}</p>
            <span class="feature-cta">
              {{ t('casino.lobby.playNow') }}
              <Icon name="arrowRight" size="xs" />
            </span>
          </div>
          <CasinoArt
            source="lobby"
            :x="feature.x"
            :y="feature.y"
            :w="feature.w"
            :h="feature.h"
            class="feature-art"
          />
        </router-link>
      </section>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_330px]">
        <div class="flex min-w-0 flex-col gap-6">
          <!-- Hot games -->
          <section class="card">
            <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('casino.lobby.hotGames') }}</h2>
              <router-link to="/casino/history" class="text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
                {{ t('casino.lobby.viewAll') }}
              </router-link>
            </div>
            <div class="grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 md:grid-cols-6">
              <router-link v-for="hot in hotGames" :key="hot.key" :to="hot.to" class="hot-card">
                <div
                  v-if="hot.key === 'scratch'"
                  class="hot-art hot-art-scratch"
                  :style="{ backgroundImage: `url(${cardBgUrl})` }"
                >
                  <img
                    v-for="sym in scratchArtSymbols"
                    :key="sym.alt"
                    :src="sym.src"
                    :alt="sym.alt"
                    class="hot-art-symbol"
                    draggable="false"
                  />
                </div>
                <CasinoArt v-else source="lobby" :x="hot.x" :y="658" :w="160" :h="103" class="hot-art" />
                <div class="hot-body">
                  <h3>{{ hot.name }}</h3>
                  <span v-if="hot.badge === 'hot'" class="hot-badge">
                    <Icon name="fire" size="xs" />
                    {{ t('casino.lobby.hotBadge') }}
                  </span>
                  <span v-else class="new-badge">{{ t('casino.lobby.newBadge') }}</span>
                </div>
                <small class="hot-foot">
                  <Icon name="users" size="xs" />
                  {{ t('casino.lobby.tryIt') }}
                </small>
              </router-link>
            </div>
          </section>

          <!-- Recent plays -->
          <section class="card">
            <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('casino.lobby.recentPlays') }}</h2>
              <router-link
                to="/casino/history"
                class="inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400"
              >
                {{ t('casino.lobby.allRecords') }}
                <Icon name="chevronRight" size="xs" />
              </router-link>
            </div>
            <div class="p-4">
              <div v-if="loadingHistory" class="flex items-center justify-center py-8">
                <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
              </div>
              <div v-else-if="recentItems.length" class="grid gap-3 sm:grid-cols-3">
                <router-link
                  v-for="row in recentItems"
                  :key="row.id"
                  :to="gameRoute(row.game)"
                  class="recent-game"
                >
                  <div
                    v-if="row.game === 'scratch'"
                    class="recent-art recent-art-scratch"
                    :style="{ backgroundImage: `url(${cardBgUrl})` }"
                    aria-hidden="true"
                  />
                  <CasinoArt
                    v-else
                    :source="recentArt(row.game).source"
                    :x="recentArt(row.game).x"
                    :y="recentArt(row.game).y"
                    :w="recentArt(row.game).w"
                    :h="recentArt(row.game).h"
                    class="recent-art"
                  />
                  <div class="min-w-0">
                    <h3>{{ gameName(row.game) }}</h3>
                    <small class="block truncate text-xs text-gray-400 dark:text-dark-500">
                      {{ formatDateTime(row.created_at) }}
                    </small>
                    <p class="text-sm text-gray-500 dark:text-dark-400">
                      <template v-if="row.status === 'pending'">{{ t('casino.lobby.inProgress') }}</template>
                      <template v-else>
                        <span :class="row.delta >= 0 ? 'pos' : 'neg'">
                          {{ row.delta >= 0 ? t('casino.lobby.win') : t('casino.lobby.lose') }}
                          {{ formatBet(Math.abs(row.delta)) }}
                        </span>
                      </template>
                    </p>
                  </div>
                </router-link>
              </div>
              <div v-else class="empty-state py-8">
                <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800">
                  <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
                </div>
                <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('casino.lobby.noRecords') }}</p>
              </div>
            </div>
          </section>
        </div>

        <!-- Leaderboard -->
        <section class="card self-start">
          <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('casino.lobby.leaderboard') }}</h2>
            <Icon name="trophy" size="md" class="text-amber-500" />
          </div>
          <ol class="space-y-1 p-3">
            <li
              v-for="row in leaderboard"
              :key="`${row.rank}-${row.player}`"
              class="flex items-center gap-3 rounded-xl px-3 py-2 transition-colors hover:bg-gray-50 dark:hover:bg-dark-800"
            >
              <span class="rank-medal" :class="`rank-${row.rank <= 3 ? row.rank : 'rest'}`">{{ row.rank }}</span>
              <span class="min-w-0 flex-1 truncate text-sm font-medium text-gray-900 dark:text-white">{{ row.player }}</span>
              <span class="text-sm font-semibold pos">+ {{ formatBet(row.profit) }}</span>
            </li>
            <li v-if="!leaderboard.length && !loadingHistory" class="px-3 py-6 text-center text-sm text-gray-400 dark:text-dark-500">
              {{ t('casino.lobby.noRecords') }}
            </li>
          </ol>
        </section>
      </div>

      <!-- Disclaimer -->
      <p class="flex items-center justify-center gap-2 pb-2 text-xs text-gray-400 dark:text-dark-500">
        <Icon name="shield" size="sm" />
        {{ t('casino.disclaimer') }}
      </p>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import CasinoArt from '@/components/casino/CasinoArt.vue'
import cardBgUrl from '@/assets/casino/scratch/cardbg.png'
import sevenSymbolUrl from '@/assets/casino/scratch/sym_seven.png'
import diamondSymbolUrl from '@/assets/casino/scratch/sym_diamond.png'
import cherrySymbolUrl from '@/assets/casino/scratch/sym_cherry.png'
import { formatBet, useCasinoData } from '@/components/casino/useCasino'
import { casinoAPI, type CasinoBetRow, type CasinoLeaderboardRow } from '@/api/casino'
import { useAuthStore } from '@/stores/auth'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const { me, loadAll } = useCasinoData()

const recentItems = ref<CasinoBetRow[]>([])
const leaderboard = ref<CasinoLeaderboardRow[]>([])
const loadingHistory = ref(true)

const balanceText = computed(
  () => formatBet(me.value?.user.balance ?? authStore.user?.balance ?? 0)
)

const features = [
  { key: 'blackjack', to: '/casino/blackjack', x: 434, y: 405, w: 177, h: 178 },
  { key: 'slots', to: '/casino/slots', x: 793, y: 405, w: 179, h: 178 },
  { key: 'wheel', to: '/casino/wheel', x: 1163, y: 405, w: 164, h: 178 }
] as const

const hotGames = [
  { key: 'blackjack', name: t('casino.games.blackjack'), to: '/casino/blackjack' as const, x: 284, badge: 'hot' as const },
  { key: 'slots', name: t('casino.games.slots'), to: '/casino/slots' as const, x: 460, badge: 'hot' as const },
  { key: 'wheel', name: t('casino.games.wheel'), to: '/casino/wheel' as const, x: 637, badge: 'hot' as const },
  { key: 'sicbo', name: t('casino.games.sicbo'), to: '/casino/sicbo' as const, x: 813, badge: 'new' as const },
  { key: 'baccarat', name: t('casino.games.baccarat'), to: '/casino/baccarat' as const, x: 990, badge: 'new' as const },
  { key: 'scratch', name: t('casino.games.scratch'), to: '/casino/scratch' as const, x: 0, badge: 'new' as const }
]

/** Flavour symbols on the scratch hot-card illustration（与游戏内连线符号同套素材）. */
const scratchArtSymbols = [
  { src: sevenSymbolUrl, alt: 'seven' },
  { src: diamondSymbolUrl, alt: 'diamond' },
  { src: cherrySymbolUrl, alt: 'cherry' }
]

const PLAYABLE_GAMES = ['blackjack', 'slots', 'wheel', 'sicbo', 'baccarat', 'scratch']

function gameRoute(game: string): string {
  return (PLAYABLE_GAMES as string[]).includes(game) ? `/casino/${game}` : '/casino'
}

function gameName(game: string): string {
  return (PLAYABLE_GAMES as string[]).includes(game) ? t(`casino.games.${game}`) : game
}

function recentArt(game: string): { source: 'lobby'; x: number; y: number; w: number; h: number } {
  // sicbo / baccarat reuse their hot-card illustration crops
  if (game === 'sicbo') return { source: 'lobby', x: 813, y: 658, w: 160, h: 103 }
  if (game === 'baccarat') return { source: 'lobby', x: 990, y: 658, w: 160, h: 103 }
  const index = ['blackjack', 'slots', 'wheel'].indexOf(game)
  return { source: 'lobby', x: [284, 558, 850][Math.max(0, index)], y: 890, w: 64, h: 61 }
}

onMounted(async () => {
  await loadAll()
  try {
    const [history, board] = await Promise.all([casinoAPI.getHistory(3), casinoAPI.getLeaderboard()])
    recentItems.value = history.items || []
    leaderboard.value = board.items || []
  } catch (error) {
    console.error('Failed to load lobby records', error)
  } finally {
    loadingHistory.value = false
  }
})
</script>

<style scoped>
/* ===== Hero banner (purple illustration keeps the reference art direction) ===== */
.casino-hero {
  position: relative;
  overflow: hidden;
  border-radius: 20px;
  min-height: 212px;
  background: linear-gradient(120deg, #6d4df6 0%, #7c5cf8 45%, #8a6cf9 100%);
  box-shadow: 0 14px 30px rgba(93, 62, 220, 0.25);
}
.casino-hero-art {
  position: absolute;
  right: 0;
  top: 0;
  height: 100%;
  width: 74%;
}
.casino-hero-art :deep(.ref-art) {
  height: 100%;
  width: 100%;
  background: transparent;
}
.casino-hero-copy {
  position: relative;
  z-index: 1;
  max-width: 46%;
  padding: 40px 36px;
}
.casino-hero-copy h2 {
  color: #fff;
  font-size: clamp(20px, 2.4vw, 30px);
  font-weight: 800;
  letter-spacing: 0.02em;
}
.casino-hero-copy p {
  margin-top: 10px;
  color: rgba(255, 255, 255, 0.85);
  font-size: 14px;
}
.casino-hero-cta {
  display: inline-block;
  margin-top: 18px;
  padding: 9px 26px;
  border-radius: 12px;
  background: #fff;
  color: #5b3df0;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 6px 14px rgba(30, 12, 90, 0.28);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}
.casino-hero-cta:hover {
  transform: translateY(-1px);
  box-shadow: 0 9px 18px rgba(30, 12, 90, 0.32);
}
@media (max-width: 760px) {
  .casino-hero-copy {
    max-width: 100%;
    padding: 24px;
  }
  .casino-hero-art {
    width: 100%;
    opacity: 0.35;
  }
}

/* ===== Feature cards ===== */
.feature-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-radius: 18px;
  padding: 18px;
  transition: transform 0.18s ease, box-shadow 0.18s ease;
}
.feature-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(63, 43, 168, 0.16);
}
.feature-blackjack {
  background: linear-gradient(135deg, #f2effe, #e7e1fd);
}
.feature-slots {
  background: linear-gradient(135deg, #fdf6e3, #fbe9c8);
}
.feature-wheel {
  background: linear-gradient(135deg, #eef1fe, #dde5fd);
}
.feature-copy {
  min-width: 0;
}
.feature-copy h2 {
  font-size: 20px;
  font-weight: 800;
  color: #312264;
}
.dark .feature-copy h2 {
  color: #cbc4f1;
}
.feature-copy h3 {
  margin-top: 6px;
  font-size: 13px;
  font-weight: 700;
  color: #574a86;
}
.dark .feature-copy h3 {
  color: #a79dd6;
}
.feature-copy p {
  margin: 5px 0 14px;
  font-size: 12.5px;
  color: #7a709e;
}
.dark .feature-copy p {
  color: #8f86b8;
}
.feature-cta {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border-radius: 10px;
  background: #6d4df6;
  color: #fff;
  font-size: 13px;
  font-weight: 700;
}
.feature-art {
  width: clamp(110px, 11vw, 150px);
  border-radius: 14px;
  flex: none;
}

/* ===== Hot games ===== */
.hot-card {
  display: block;
  overflow: hidden;
  border-radius: 14px;
  border: 1px solid rgb(229 231 235);
  background: #fff;
  padding-bottom: 12px;
  transition: 0.18s;
}
.dark .hot-card {
  border-color: rgb(48 55 73);
  background: rgb(23 28 43);
}
.hot-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(40, 30, 90, 0.12);
}
.hot-art {
  width: 100%;
  border-radius: 0;
  background: #f0eef9;
}
/* Scratch card illustration: purple-gold card face + three lucky symbols */
.hot-art-scratch {
  aspect-ratio: 160 / 103;
  background-size: cover;
  background-position: center;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4%;
}
.hot-art-symbol {
  width: 22%;
  max-width: 44px;
  object-fit: contain;
  filter: drop-shadow(0 2px 3px rgba(46, 16, 101, 0.55));
}
.hot-body {
  padding: 10px 12px 4px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}
.hot-body h3 {
  font-size: 14px;
  font-weight: 700;
  color: rgb(17 24 39);
}
.dark .hot-body h3 {
  color: #fff;
}
.hot-badge,
.new-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  flex: none;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 999px;
}
.hot-badge {
  color: #e2482f;
  background: #fdeceb;
}
.new-badge {
  color: #6d4df6;
  background: #f1eeff;
}
.dark .new-badge {
  color: #a78bfa;
  background: rgba(109, 77, 246, 0.18);
}
.hot-foot {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 12px 0;
  font-size: 11.5px;
  color: #9ca3af;
}

/* ===== Recent plays ===== */
.recent-game {
  display: flex;
  gap: 10px;
  align-items: center;
  border-radius: 12px;
  padding: 8px;
  transition: background 0.15s ease;
}
.recent-game:hover {
  background: rgb(249 250 251);
}
.dark .recent-game:hover {
  background: rgb(30 36 54);
}
.recent-art {
  width: 64px;
  border-radius: 10px;
  flex: none;
}
.recent-art-scratch {
  aspect-ratio: 160 / 103;
  background-size: cover;
  background-position: center;
  box-shadow: inset 0 0 0 1px rgba(202, 138, 26, 0.4);
}
.recent-game h3 {
  font-size: 14px;
  font-weight: 700;
  color: rgb(17 24 39);
}
.dark .recent-game h3 {
  color: #fff;
}
.recent-game .pos {
  color: #059669;
  font-weight: 700;
}
.recent-game .neg {
  color: #dc2626;
  font-weight: 700;
}
.pos {
  color: rgb(5 150 105);
}
.neg {
  color: rgb(220 38 38);
}

/* ===== Leaderboard ===== */
.rank-medal {
  display: inline-flex;
  width: 24px;
  height: 24px;
  flex: none;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 800;
  color: #fff;
}
.rank-1 {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
}
.rank-2 {
  background: linear-gradient(135deg, #cbd5e1, #94a3b8);
}
.rank-3 {
  background: linear-gradient(135deg, #fdba74, #ea580c);
}
.rank-rest {
  background: rgb(243 244 246);
  color: #6b7280;
}
.dark .rank-rest {
  background: rgb(48 55 73);
  color: #9ca3af;
}

@media (prefers-reduced-motion: reduce) {
  .feature-card,
  .hot-card,
  .casino-hero-cta {
    transition: none;
  }
}
</style>
