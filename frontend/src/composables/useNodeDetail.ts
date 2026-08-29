import { ref } from 'vue'
import type { FlowConfidence, NodeDetail } from '@/types/flowGraph'

/**
 * Fetches detailed explanation for a flow graph node.
 */
export function useNodeDetail() {
  const detail = ref<NodeDetail | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * Load node detail from the backend.
   */
  async function loadDetail(params: {
    workspace: string
    nodeId: string
    symbol: string
    file?: string
    line?: number
    title?: string
    confidence?: FlowConfidence
    code?: string
    experience?: string
    language?: string
  }): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const q = new URLSearchParams({
        workspace: params.workspace,
        nodeId: params.nodeId,
        symbol: params.symbol,
      })
      if (params.file) q.set('file', params.file)
      if (params.line) q.set('line', String(params.line))
      if (params.title) q.set('title', params.title)
      if (params.code) q.set('code', params.code)
      if (params.confidence) q.set('confidence', params.confidence)
      if (params.experience) q.set('experience', params.experience)
      if (params.language) q.set('language', params.language)

      const res = await fetch(`/api/graph/node?${q}`)
      if (!res.ok) throw new Error(`Detail failed (${res.status})`)
      detail.value = (await res.json()) as NodeDetail
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load detail'
      detail.value = null
    } finally {
      loading.value = false
    }
  }

  function clear(): void {
    detail.value = null
    error.value = null
  }

  return { detail, loading, error, loadDetail, clear }
}
