import { ref } from 'vue'
import type { GraphRootPayload } from '@/types/flowGraph'
import { enrichSymbolNodes, fetchGraphRoot } from '@/utils/flowGraphEnrich'
import {
  ENRICHMENT_HORIZON_DEPTH,
  WARM_CONCURRENCY,
  WARM_REVEAL_COUNT,
  createSymbolFlowState,
  enrichmentHorizon,
} from '@/utils/flowGraphUtils'
import type { useFlowGraphCache } from '@/composables/useFlowGraphCache'

export interface WarmProgress {
  done: number
  total: number
  currentSymbol: string
}

/**
 * Warms flow graphs for multiple symbols in parallel with progress tracking.
 */
export function useFileFlowWarm(cache: ReturnType<typeof useFlowGraphCache>) {
  const warming = ref(false)
  const progress = ref<WarmProgress>({ done: 0, total: 0, currentSymbol: '' })

  async function warmFile(
    payload: Omit<GraphRootPayload, 'symbol'>,
    symbols: string[],
  ): Promise<void> {
    if (!symbols.length) return

    warming.value = true
    progress.value = { done: 0, total: symbols.length, currentSymbol: '' }

    const queue = [...symbols]
    const inFlight = new Set<string>()

    async function warmOne(sym: string): Promise<void> {
      progress.value = { ...progress.value, currentSymbol: sym }
      const fullPayload: GraphRootPayload = { ...payload, symbol: sym }
      const graph = await fetchGraphRoot(fullPayload)
      const state = createSymbolFlowState(graph, WARM_REVEAL_COUNT)
      const revealed = [...state.revealedIds]
      const frontier = revealed[revealed.length - 1] ?? state.rootId
      const horizon = enrichmentHorizon(frontier, state.allEdges, ENRICHMENT_HORIZON_DEPTH, state.enrichedIds)
      const toEnrich = [...new Set([...revealed, ...horizon])]
      await enrichSymbolNodes(state, fullPayload, toEnrich, inFlight)
      cache.set(payload.filePath, sym, state)
      progress.value = {
        ...progress.value,
        done: progress.value.done + 1,
      }
    }

    async function worker(): Promise<void> {
      while (queue.length > 0) {
        const sym = queue.shift()
        if (!sym) break
        await warmOne(sym)
      }
    }

    try {
      await Promise.all(Array.from({ length: Math.min(WARM_CONCURRENCY, symbols.length) }, () => worker()))
      cache.markFileWarmed(payload.filePath)
    } finally {
      warming.value = false
      progress.value = { ...progress.value, currentSymbol: '' }
    }
  }

  return { warming, progress, warmFile }
}
