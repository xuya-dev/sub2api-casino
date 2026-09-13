<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <!-- Page head -->
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('casino.history.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('casino.history.description') }}</p>
        </div>
        <router-link to="/casino" class="btn btn-secondary">{{ t('casino.backToLobby') }}</router-link>
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="load">
          <svg v-if="loading" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          <Icon v-else name="refresh" size="sm" />
          {{ t('casino.history.refresh') }}
        </button>
      </div>

      <div class="card">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:border-dark-700 dark:text-dark-400">
                <th class="px-6 py-3.5">{{ t('casino.history.time') }}</th>
                <th class="px-6 py-3.5">{{ t('casino.history.game') }}</th>
                <th class="px-6 py-3.5 text-right">{{ t('casino.history.bet') }}</th>
                <th class="px-6 py-3.5 text-right">{{ t('casino.history.payout') }}</th>
                <th class="px-6 py-3.5 text-right">{{ t('casino.history.delta') }}</th>
                <th class="px-6 py-3.5 text-right">{{ t('casino.history.balance') }}</th>
                <th class="px-6 py-3.5">{{ t('casino.history.status') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-if="loading && !items.length">
                <td colspan="7" class="px-6 py-10 text-center">
                  <svg class="mx-auto h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                </td>
              </tr>
              <tr
                v-for="row in items"
                :key="row.id"
                class="transition-colors hover:bg-gray-50 dark:hover:bg-dark-800"
              >
                <td class="whitespace-nowrap px-6 py-3.5 text-gray-500 dark:text-dark-400">
                  {{ formatDateTime(row.created_at) }}
                </td>
                <td class="px-6 py-3.5">
                  <span class="font-medium text-gray-900 dark:text-white">{{ gameName(row.game) }}</span>
                </td>
                <td class="whitespace-nowrap px-6 py-3.5 text-right text-gray-900 dark:text-white">
                  {{ formatBet(row.bet) }}
                </td>
                <td class="whitespace-nowrap px-6 py-3.5 text-right text-gray-900 dark:text-white">
                  {{ formatBet(row.payout) }}
                </td>
                <td
                  class="whitespace-nowrap px-6 py-3.5 text-right font-semibold"
                  :class="row.delta >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'"
                >
                  {{ row.delta >= 0 ? '+' : '−' }}{{ formatBet(Math.abs(row.delta)) }}
                </td>
                <td class="whitespace-nowrap px-6 py-3.5 text-right text-gray-900 dark:text-white">
                  {{ formatBet(row.balance_after) }}
                </td>
                <td class="px-6 py-3.5">
                  <span
                    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
                    :class="
                      row.status === 'settled'
                        ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                        : 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                    "
                  >
                    {{ row.status === 'settled' ? t('casino.status.settled') : t('casino.status.playing') }}
                  </span>
                </td>
              </tr>
              <tr v-if="!loading && !items.length">
                <td colspan="7" class="px-6 py-12 text-center text-sm text-gray-400 dark:text-dark-500">
                  {{ t('casino.history.empty') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <p class="flex items-center justify-center gap-2 pb-2 text-xs text-gray-400 dark:text-dark-500">
        <Icon name="shield" size="sm" />
        {{ t('casino.disclaimer') }}
      </p>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatBet } from '@/components/casino/useCasino'
import { casinoAPI, type CasinoBetRow } from '@/api/casino'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const items = ref<CasinoBetRow[]>([])
const loading = ref(false)

const GAME_NAME_KEYS = ['blackjack', 'slots', 'wheel', 'sicbo', 'baccarat']

function gameName(game: string): string {
  return (GAME_NAME_KEYS as string[]).includes(game) ? t(`casino.games.${game}`) : game
}

async function load() {
  loading.value = true
  try {
    const { items: rows } = await casinoAPI.getHistory(50)
    items.value = rows || []
  } catch (error) {
    console.error('Failed to load casino history', error)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
