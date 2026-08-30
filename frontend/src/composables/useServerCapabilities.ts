import { onMounted, ref } from 'vue'
import { api } from '@/api'

/**
 * Server feature flags from GET /api/health (e.g. local workspace source on self-hosted).
 */
export function useServerCapabilities() {
  const allowLocalSource = ref(false)
  const loaded = ref(false)

  onMounted(async () => {
    try {
      const health = await api.health()
      allowLocalSource.value = health.allowLocalSource
    } catch {
      allowLocalSource.value = false
    } finally {
      loaded.value = true
    }
  })

  return { allowLocalSource, loaded }
}
