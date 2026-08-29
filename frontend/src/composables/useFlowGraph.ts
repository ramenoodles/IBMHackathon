import { ref } from 'vue'
import type {
  EnrichPatch,
  FlowEdge,
  FlowGraph,
  FlowNode,
  GraphExpandPayload,
  GraphRootPayload,
} from '@/types/flowGraph'
import { enrichSymbolNodes, fetchGraphRoot } from '@/utils/flowGraphEnrich'
import {
  ENRICHMENT_HORIZON_DEPTH,
  enrichmentHorizon,
  type SymbolFlowState,
} from '@/utils/flowGraphUtils'
import type { useFlowGraphCache } from '@/composables/useFlowGraphCache'

/**
 * Manages scan-first flow graph with cache, progressive reveal, and async LLM labels.
 */
export function useFlowGraph(cache: ReturnType<typeof useFlowGraphCache>) {
  const allNodes = ref<FlowNode[]>([])
  const allEdges = ref<FlowEdge[]>([])
  const nodes = ref<FlowNode[]>([])
  const edges = ref<FlowEdge[]>([])
  const revealedIds = ref<Set<string>>(new Set())
  const enrichedIds = ref<Set<string>>(new Set())
  const inFlightIds = ref<Set<string>>(new Set())
  const rootId = ref('')
  const symbol = ref('')
  const loading = ref(false)
  const enriching = ref(false)
  const expanding = ref(false)
  const error = ref<string | null>(null)
  const isMock = ref(false)
  const parentPath = ref<string[]>([])
  const currentFilePath = ref('')
  const currentWorkspace = ref('')
  const lastPayload = ref<GraphRootPayload | null>(null)

  function snapshotState(): SymbolFlowState {
    return {
      allNodes: allNodes.value.map((n) => ({ ...n })),
      allEdges: [...allEdges.value],
      rootId: rootId.value,
      revealedIds: new Set(revealedIds.value),
      enrichedIds: new Set(enrichedIds.value),
      isMock: isMock.value,
      parentPath: [...parentPath.value],
    }
  }

  function persistToCache(): void {
    if (!currentFilePath.value || !symbol.value) return
    cache.set(currentFilePath.value, symbol.value, snapshotState())
  }

  function hydrateFromState(state: SymbolFlowState, sym: string): void {
    allNodes.value = state.allNodes.map((n) => ({ ...n }))
    allEdges.value = [...state.allEdges]
    rootId.value = state.rootId
    revealedIds.value = new Set(state.revealedIds)
    enrichedIds.value = new Set(state.enrichedIds)
    isMock.value = state.isMock
    parentPath.value = [...state.parentPath]
    symbol.value = sym
    syncVisible()
  }

  function syncVisible(): void {
    const revealed = revealedIds.value
    nodes.value = allNodes.value.filter((n) => revealed.has(n.id))
    edges.value = allEdges.value.filter((e) => revealed.has(e.from) && revealed.has(e.to))
  }

  function applyPatches(patches: EnrichPatch[]): void {
    for (const patch of patches) {
      const node = allNodes.value.find((n) => n.id === patch.id)
      if (!node) continue
      if (patch.title) node.title = patch.title
      if (patch.summary) node.summary = patch.summary
      enrichedIds.value.add(patch.id)
    }
    syncVisible()
    persistToCache()
  }

  async function enrichNodes(nodeIds: string[], opts?: { background?: boolean }): Promise<void> {
    const payload = lastPayload.value
    if (!payload || nodeIds.length === 0) return

    const pending = nodeIds.filter((id) => !enrichedIds.value.has(id) && !inFlightIds.value.has(id))
    if (pending.length === 0) return

    if (!opts?.background) enriching.value = true
    try {
      const state = snapshotState()
      await enrichSymbolNodes(state, payload, pending, inFlightIds.value)
      allNodes.value = state.allNodes
      enrichedIds.value = state.enrichedIds
      syncVisible()
      persistToCache()
    } finally {
      if (!opts?.background) enriching.value = false
    }
  }

  function prefetchAroundNode(nodeId: string): void {
    const payload = lastPayload.value
    if (!payload) return

    const ids: string[] = []
    if (!enrichedIds.value.has(nodeId)) ids.push(nodeId)
    ids.push(
      ...enrichmentHorizon(nodeId, allEdges.value, ENRICHMENT_HORIZON_DEPTH, enrichedIds.value),
    )
    const unique = [...new Set(ids)].filter((id) => !enrichedIds.value.has(id))
    if (unique.length) void enrichNodes(unique, { background: true })
  }

  function hasHiddenChildren(nodeId: string): boolean {
    const revealed = revealedIds.value
    return allEdges.value.some((e) => e.from === nodeId && !revealed.has(e.to))
  }

  function revealFromNode(nodeId: string): string[] {
    const revealed = new Set(revealedIds.value)
    const newly: string[] = []
    for (const edge of allEdges.value) {
      if (edge.from === nodeId && !revealed.has(edge.to)) {
        revealed.add(edge.to)
        newly.push(edge.to)
      }
    }
    if (newly.length === 0) return []
    revealedIds.value = revealed
    syncVisible()
    persistToCache()
    void enrichNodes(newly)
    for (const id of newly) prefetchAroundNode(id)
    prefetchAroundNode(nodeId)
    return newly
  }

  function mergeFragment(fragment: FlowGraph): void {
    const existingIds = new Set(allNodes.value.map((n) => n.id))
    for (const node of fragment.nodes) {
      if (!existingIds.has(node.id)) {
        allNodes.value.push(node)
        existingIds.add(node.id)
      }
    }
    const edgeKeys = new Set(allEdges.value.map((e) => `${e.from}->${e.to}`))
    for (const edge of fragment.edges) {
      const key = `${edge.from}->${edge.to}`
      if (!edgeKeys.has(key)) {
        allEdges.value.push(edge)
        edgeKeys.add(key)
      }
    }
    const expanded = allNodes.value.find((n) => n.id === fragment.rootId)
    if (expanded) {
      expanded.collapsed = false
      expanded.expandable = false
    }
    syncVisible()
    persistToCache()
  }

  function activateSymbol(sym: string, payload: GraphRootPayload): boolean {
    const cached = cache.get(payload.filePath, sym)
    if (!cached) return false
    error.value = null
    currentFilePath.value = payload.filePath
    currentWorkspace.value = payload.workspacePath
    lastPayload.value = { ...payload, symbol: sym }
    hydrateFromState(cached, sym)
    const frontier = [...revealedIds.value].at(-1) ?? rootId.value
    if (frontier) prefetchAroundNode(frontier)
    return true
  }

  async function loadRoot(payload: GraphRootPayload, opts?: { skipCache?: boolean }): Promise<void> {
    if (!opts?.skipCache && cache.has(payload.filePath, payload.symbol)) {
      activateSymbol(payload.symbol, payload)
      return
    }

    loading.value = true
    error.value = null
    parentPath.value = []
    currentFilePath.value = payload.filePath
    currentWorkspace.value = payload.workspacePath
    symbol.value = payload.symbol
    lastPayload.value = payload
    enrichedIds.value = new Set()
    inFlightIds.value = new Set()
    revealedIds.value = new Set()

    try {
      const data = await fetchGraphRoot(payload)
      allNodes.value = data.nodes
      allEdges.value = data.edges
      rootId.value = data.rootId
      isMock.value = Boolean(data.mock) && data.nodes.length === 0

      if (data.rootId) {
        revealedIds.value = new Set([data.rootId])
      }
      syncVisible()
      loading.value = false
      persistToCache()
      if (data.rootId) {
        void enrichNodes([data.rootId])
        prefetchAroundNode(data.rootId)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load graph'
      allNodes.value = []
      allEdges.value = []
      nodes.value = []
      edges.value = []
      loading.value = false
    }
  }

  async function expandNode(
    nodeId: string,
    payload: Omit<GraphExpandPayload, 'nodeId' | 'parentPath' | 'expandLimit'>,
    limit = 3,
  ): Promise<void> {
    expanding.value = true
    error.value = null
    const path = [...parentPath.value, nodeId]

    try {
      const res = await fetch('/api/graph/expand', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...payload,
          nodeId,
          parentPath: path,
          expandLimit: limit,
        }),
      })
      if (!res.ok) throw new Error(`Expand failed (${res.status})`)
      const fragment = (await res.json()) as FlowGraph
      mergeFragment(fragment)
      parentPath.value = path
      const newIds = fragment.nodes.map((n) => n.id).filter((id) => !revealedIds.value.has(id))
      for (const id of newIds) {
        revealedIds.value.add(id)
      }
      revealedIds.value = new Set(revealedIds.value)
      syncVisible()
      persistToCache()
      if (newIds.length) {
        void enrichNodes(newIds)
        for (const id of newIds) prefetchAroundNode(id)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Expand failed'
    } finally {
      expanding.value = false
    }
  }

  function reset(): void {
    allNodes.value = []
    allEdges.value = []
    nodes.value = []
    edges.value = []
    revealedIds.value = new Set()
    enrichedIds.value = new Set()
    inFlightIds.value = new Set()
    rootId.value = ''
    symbol.value = ''
    error.value = null
    isMock.value = false
    parentPath.value = []
    lastPayload.value = null
  }

  return {
    nodes,
    edges,
    rootId,
    symbol,
    loading,
    enriching,
    expanding,
    error,
    isMock,
    currentFilePath,
    currentWorkspace,
    loadRoot,
    activateSymbol,
    expandNode,
    revealFromNode,
    hasHiddenChildren,
    prefetchAroundNode,
    reset,
  }
}
