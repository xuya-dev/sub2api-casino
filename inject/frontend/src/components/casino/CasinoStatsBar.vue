<template>
  <div class="card overflow-hidden">
    <div class="flex flex-wrap items-stretch divide-y divide-gray-100 dark:divide-dark-700 sm:divide-y-0">
      <div
        v-for="item in items"
        :key="item.label"
        class="flex min-w-[160px] flex-1 items-center gap-3 px-5 py-4 sm:border-l sm:border-gray-100 sm:first:border-l-0 dark:sm:border-dark-700"
      >
        <span
          :class="[
            'flex h-10 w-10 flex-none items-center justify-center rounded-xl',
            item.iconClass || 'bg-primary-50 text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
          ]"
        >
          <Icon :name="item.icon" size="md" />
        </span>
        <span class="min-w-0">
          <span class="block truncate text-xs font-medium text-gray-500 dark:text-dark-400">{{ item.label }}</span>
          <strong
            :class="['block truncate text-lg font-bold', item.valueClass || 'text-gray-900 dark:text-white']"
            :title="item.value"
          >
            {{ item.value }}
          </strong>
        </span>
      </div>
      <div v-if="$slots.default" class="flex items-center px-5 py-4 sm:ml-auto">
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * CasinoStatsBar — the metric strip shown on every game page
 * (balance / today's P/L / round bet / round win + a right-side slot).
 */
import Icon from '@/components/icons/Icon.vue'

export interface CasinoStatItem {
  label: string
  value: string
  icon: InstanceType<typeof Icon>['$props']['name']
  iconClass?: string
  valueClass?: string
}

defineProps<{
  items: CasinoStatItem[]
}>()
</script>
