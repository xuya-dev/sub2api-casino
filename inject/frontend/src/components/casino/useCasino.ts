/**
 * Shared casino data composable.
 * Module-level refs so all casino views share one copy of meta/me state;
 * balance changes flow through refreshMe() plus authStore.refreshUser() to
 * keep the main-site top bar in sync.
 */
import { ref } from 'vue'
import { casinoAPI, type CasinoMeta, type CasinoMe } from '@/api/casino'

const meta = ref<CasinoMeta | null>(null)
const me = ref<CasinoMe | null>(null)
const loading = ref(false)
const loadError = ref('')

let metaPromise: Promise<void> | null = null

export function useCasinoData() {
  async function ensureMeta(): Promise<void> {
    if (meta.value) return
    if (!metaPromise) {
      metaPromise = casinoAPI
        .getMeta()
        .then((value) => {
          meta.value = value
        })
        .finally(() => {
          metaPromise = null
        })
    }
    await metaPromise
  }

  async function refreshMe(): Promise<void> {
    try {
      me.value = await casinoAPI.getMe()
    } catch (error) {
      console.error('Failed to load casino profile', error)
    }
  }

  async function loadAll(): Promise<void> {
    loading.value = true
    loadError.value = ''
    try {
      await ensureMeta()
      await refreshMe()
    } catch (error) {
      console.error('Failed to load casino data', error)
      loadError.value = error instanceof Error ? error.message : String(error)
    } finally {
      loading.value = false
    }
  }

  return { meta, me, loading, loadError, loadAll, ensureMeta, refreshMe }
}

/** Shared bet formatting helper. */
export function formatBet(value: number | null | undefined): string {
  return Number(value || 0).toFixed(2)
}
