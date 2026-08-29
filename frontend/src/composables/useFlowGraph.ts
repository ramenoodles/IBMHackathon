import { ref } from 'vue'
import type {
  FlowEdge,
  FlowGraph,
  FlowNode,
  GraphExpandPayload,
  GraphRootPayload,
} from '@/types/flowGraph'
import { enrichSymbolNodes, fetchGraphRoot } from '@/utils/flowGraphEnrich'
import {
  ENRICHMENT_HORIZON_DEPTH,
  SILENT_BUFFER_STEPS,
  entryOnlyRevealedIds,
  enrichmentHorizon,
  revealedPathLength,
  silentPrefetchTargets,
  type SymbolFlowState,
  visibleFrontierId,
} from '@/utils/flowGraphUtils'
import type { useFlowGraphCache } from '@/composables/useFlowGraphCache'
import { api } from '@/api'

const MAX_EXPAND_DEPTH = 4

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
  const mappingFullFlow = ref(false)
  const fullyExpanded = ref(false)
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
      fullyExpanded: fullyExpanded.value,
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
    enrichedIds.value = new Set(state.enrichedIds)
    isMock.value = state.isMock
    parentPath.value = [...state.parentPath]
    fullyExpanded.value = state.fullyExpanded ?? false
    symbol.value = sym

    if (fullyExpanded.value) {
      revealedIds.value = new Set(state.allNodes.map((n) => n.id))
    } else {
      revealedIds.value = entryOnlyRevealedIds(state.rootId, state.allNodes, state.allEdges)
    }
    syncVisible()
  }

  function syncVisible(): void {
    const revealed = revealedIds.value
    nodes.value = allNodes.value.filter((n) => revealed.has(n.id))
    edges.value = allEdges.value.filter((e) => revealed.has(e.from) && revealed.has(e.to))
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

  function maintainSilentBuffer(): void {
    if (fullyExpanded.value || !rootId.value || !lastPayload.value) return

    const visibleCount = revealedPathLength(
      rootId.value,
      allNodes.value,
      allEdges.value,
      revealedIds.value,
    )
    const buffer = silentPrefetchTargets(
      rootId.value,
      allNodes.value,
      allEdges.value,
      visibleCount,
      SILENT_BUFFER_STEPS,
    )
    const frontier =
      visibleFrontierId(rootId.value, allNodes.value, allEdges.value, revealedIds.value) ||
      rootId.value
    const horizon = enrichmentHorizon(frontier, allEdges.value, ENRICHMENT_HORIZON_DEPTH, enrichedIds.value)
    const toEnrich = [...new Set([...buffer, ...horizon])].filter((id) => !enrichedIds.value.has(id))
    if (toEnrich.length) void enrichNodes(toEnrich, { background: true })
  }

  function prefetchAroundNode(_nodeId: string): void {
    maintainSilentBuffer()
  }

  function hasHiddenChildren(nodeId: string): boolean {
    if (fullyExpanded.value) return false
    const revealed = revealedIds.value
    return allEdges.value.some((e) => e.from === nodeId && !revealed.has(e.to))
  }

  function revealFromNode(nodeId: string): string[] {
    if (fullyExpanded.value) return []

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
    void enrichNodes(newly, { background: true })
    maintainSilentBuffer()
    return newly
  }

  function mergeFragment(fragment: FlowGraph, revealAll = false): void {
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
    if (revealAll || fullyExpanded.value) {
      for (const node of fragment.nodes) {
        revealedIds.value.add(node.id)
      }
      revealedIds.value = new Set(revealedIds.value)
    }
    syncVisible()
    persistToCache()
  }

  function revealAllKnownNodes(): void {
    for (const node of allNodes.value) {
      revealedIds.value.add(node.id)
    }
    revealedIds.value = new Set(revealedIds.value)
    syncVisible()
  }

  function collapsedExpandableNodes(): FlowNode[] {
    return allNodes.value.filter((n) => n.collapsed && n.expandable)
  }

  async function fetchExpandFragment(
    nodeId: string,
    payload: Omit<GraphExpandPayload, 'nodeId' | 'parentPath' | 'expandLimit'>,
    path: string[],
    limit: number,
  ): Promise<FlowGraph | null> {
    try {
      return await api.expand({ ...payload, nodeId, parentPath: path, expandLimit: limit })
    } catch {
      return null
    }
  }

  async function revealFullFlow(
    payload: Omit<GraphExpandPayload, 'nodeId' | 'parentPath' | 'expandLimit'>,
  ): Promise<void> {
    if (fullyExpanded.value || mappingFullFlow.value) return

    mappingFullFlow.value = true
    error.value = null

    try {
      let guard = 0
      while (guard < 32) {
        guard++
        revealAllKnownNodes()

        const collapsed = collapsedExpandableNodes()
        if (collapsed.length === 0) break

        let expandedAny = false
        for (const node of collapsed) {
          if (parentPath.value.length >= MAX_EXPAND_DEPTH) continue
          const path = [...parentPath.value, node.id]
          const fragment = await fetchExpandFragment(
            node.id,
            payload,
            path,
            node.childCount || 6,
          )
          if (!fragment) continue
          mergeFragment(fragment, true)
          parentPath.value = path
          expandedAny = true
        }
        if (!expandedAny) break
      }

      revealAllKnownNodes()

      const allIds = allNodes.value.map((n) => n.id)
      await enrichNodes(allIds, { background: false })

      fullyExpanded.value = true
      persistToCache()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to map full flow'
    } finally {
      mappingFullFlow.value = false
    }
  }

  function activateSymbol(sym: string, payload: GraphRootPayload): boolean {
    const cached = cache.get(payload.filePath, sym)
    if (!cached) return false
    error.value = null
    currentFilePath.value = payload.filePath
    currentWorkspace.value = payload.workspaceId
    lastPayload.value = { ...payload, symbol: sym }
    hydrateFromState(cached, sym)
    maintainSilentBuffer()
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
    currentWorkspace.value = payload.workspaceId
    symbol.value = payload.symbol
    lastPayload.value = payload
    enrichedIds.value = new Set()
    inFlightIds.value = new Set()
    fullyExpanded.value = false
    revealedIds.value = new Set()

    try {
      const data = await fetchGraphRoot(payload)
      allNodes.value = data.nodes
      allEdges.value = data.edges
      rootId.value = data.rootId
      isMock.value = Boolean(data.mock) && data.nodes.length === 0

      if (data.rootId) {
        revealedIds.value = entryOnlyRevealedIds(data.rootId, data.nodes, data.edges)
      }
      syncVisible()
      loading.value = false
      persistToCache()
      maintainSilentBuffer()
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
    if (fullyExpanded.value) return

    expanding.value = true
    error.value = null
    const path = [...parentPath.value, nodeId]

    try {
      const fragment = await fetchExpandFragment(nodeId, payload, path, limit)
      if (!fragment) throw new Error('Expand failed')
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
        void enrichNodes(newIds, { background: true })
        maintainSilentBuffer()
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
    fullyExpanded.value = false
    mappingFullFlow.value = false
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
    mappingFullFlow,
    fullyExpanded,
    error,
    isMock,
    currentFilePath,
    currentWorkspace,
    loadRoot,
    activateSymbol,
    expandNode,
    revealFromNode,
    revealFullFlow,
    hasHiddenChildren,
    prefetchAroundNode,
    reset,
  }
}
