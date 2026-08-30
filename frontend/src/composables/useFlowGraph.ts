import { ref } from 'vue'
import type {
  FlowEdge,
  FlowGraph,
  FlowNode,
  GraphExpandPayload,
  GraphRootPayload,
} from '@/types/flowGraph'
import { enrichSymbolNodes, fetchGraphRoot, type EnrichPatchResult } from '@/utils/flowGraphEnrich'
import {
  ENRICHMENT_HORIZON_DEPTH,
  SILENT_BUFFER_STEPS,
  entryOnlyRevealedIds,
  enrichmentHorizon,
  pruneGraphToRoot,
  revealedPathLength,
  silentPrefetchTargets,
  type SymbolFlowState,
  visibleFrontierId,
} from '@/utils/flowGraphUtils'
import type { useFlowGraphCache } from '@/composables/useFlowGraphCache'
import { api } from '@/api'

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
  const mappingProgress = ref(0)
  const fullyExpanded = ref(false)
  const error = ref<string | null>(null)
  const isMock = ref(false)
  const parentPath = ref<string[]>([])
  const currentFilePath = ref('')
  const currentWorkspace = ref('')
  const lastPayload = ref<GraphRootPayload | null>(null)
  const enrichError = ref<string | null>(null)

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
    const pruned = pruneGraphToRoot(state.rootId, state.allNodes, state.allEdges)
    allNodes.value = pruned.nodes.map((n) => ({ ...n }))
    allEdges.value = [...pruned.edges]
    rootId.value = state.rootId
    enrichedIds.value = new Set(state.enrichedIds)
    isMock.value = state.isMock
    parentPath.value = [...state.parentPath]
    fullyExpanded.value = state.fullyExpanded ?? false
    symbol.value = sym

    const validRevealed = [...new Set(state.revealedIds ?? [])].filter((id) =>
      allNodes.value.some((node) => node.id === id),
    )

    if (fullyExpanded.value) {
      revealedIds.value = new Set(allNodes.value.map((n) => n.id))
    } else if (validRevealed.length > 0) {
      revealedIds.value = new Set(validRevealed)
    } else {
      revealedIds.value = entryOnlyRevealedIds(state.rootId, allNodes.value, allEdges.value)
    }
    syncVisible()
  }

  function syncVisible(): void {
    const revealed = revealedIds.value
    nodes.value = allNodes.value.filter((n) => revealed.has(n.id))
    edges.value = allEdges.value.filter((e) => revealed.has(e.from) && revealed.has(e.to))
  }

  async function enrichNodes(
    nodeIds: string[],
    opts?: { background?: boolean; priorityVisible?: boolean },
  ): Promise<void> {
    const payload = lastPayload.value
    if (!payload || nodeIds.length === 0) return

    const pending = nodeIds.filter((id) => !enrichedIds.value.has(id) && !inFlightIds.value.has(id))
    if (pending.length === 0) return

    if (!opts?.background) enriching.value = true
    try {
      const priorityIds = opts?.priorityVisible ? new Set(revealedIds.value) : undefined
      // Pass a snapshot of the current nodes for input building only.
      // Patches are applied directly to the live allNodes.value after the await
      // so concurrent enrichment calls cannot clobber each other's results.
      const result: EnrichPatchResult = await enrichSymbolNodes(
        allNodes.value,
        payload,
        pending,
        inFlightIds.value,
        enrichedIds.value,
        priorityIds,
      )
      if (result.enrichError) {
        enrichError.value = result.enrichError
      }
      // Apply patches onto live allNodes — never replace the whole array.
      const byId = new Map(allNodes.value.map((n) => [n.id, n]))
      const newEnrichedIds = new Set(enrichedIds.value)
      for (const patch of result.patches) {
        const node = byId.get(patch.id)
        if (!node) continue
        if (patch.title) node.title = patch.title
        if (patch.summary) node.summary = patch.summary
        if (patch.labelSource) {
          node.labelSource = patch.labelSource as FlowNode['labelSource']
          if (patch.labelSource === 'ai' || patch.labelSource === 'heuristic') {
            node.confidence = 'inferred'
          }
        }
        newEnrichedIds.add(patch.id)
      }
      // Trigger reactivity by replacing the array reference.
      allNodes.value = [...allNodes.value]
      enrichedIds.value = newEnrichedIds
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
    if (toEnrich.length) void enrichNodes(toEnrich, { background: true, priorityVisible: true })
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
    void enrichNodes(newly, { background: true, priorityVisible: true })
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

    const root = rootId.value || fragment.rootId
    if (root) {
      const pruned = pruneGraphToRoot(root, allNodes.value, allEdges.value)
      allNodes.value = pruned.nodes
      allEdges.value = pruned.edges
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
    if (rootId.value) {
      const pruned = pruneGraphToRoot(rootId.value, allNodes.value, allEdges.value)
      allNodes.value = pruned.nodes
      allEdges.value = pruned.edges
    }
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

    const EXPAND_WEIGHT = 0.72
    const ENRICH_WEIGHT = 0.28

    function setMappingProgress(
      expandDone: number,
      expandGoal: number,
      enrichDone: number,
      enrichGoal: number,
    ): void {
      const expandRatio = expandGoal > 0 ? Math.min(1, expandDone / expandGoal) : 1
      const enrichRatio = enrichGoal > 0 ? Math.min(1, enrichDone / enrichGoal) : 1
      const pct = expandRatio * EXPAND_WEIGHT + enrichRatio * ENRICH_WEIGHT
      mappingProgress.value = Math.min(99, Math.round(pct * 100))
    }

    if (rootId.value) {
      const pruned = pruneGraphToRoot(rootId.value, allNodes.value, allEdges.value)
      allNodes.value = pruned.nodes
      allEdges.value = pruned.edges
    }

    mappingFullFlow.value = true
    mappingProgress.value = 0
    error.value = null

    try {
      let expandedCount = 0
      let guard = 0
      while (guard < 32) {
        guard++
        revealAllKnownNodes()

        const collapsed = collapsedExpandableNodes()
        if (collapsed.length === 0) break

        let expandedAny = false
        for (const node of collapsed) {
          const path = [node.id]
          const fragment = await fetchExpandFragment(
            node.id,
            payload,
            path,
            node.childCount ? node.childCount + 1 : 6,
          )
          if (!fragment) continue
          mergeFragment(fragment, true)
          expandedCount += 1
          expandedAny = true
          const remaining = collapsedExpandableNodes().length
          setMappingProgress(expandedCount, expandedCount + remaining, 0, 1)
        }
        if (!expandedAny) break
      }

      revealAllKnownNodes()
      const expandGoal = Math.max(expandedCount, 1)
      const allIds = allNodes.value.map((n) => n.id)
      const enrichGoal = allIds.length

      for (let i = 0; i < allIds.length; i += 8) {
        const batch = allIds.slice(i, i + 8)
        await enrichNodes(batch, { background: false })
        setMappingProgress(expandGoal, expandGoal, Math.min(enrichGoal, i + batch.length), enrichGoal)
      }

      fullyExpanded.value = collapsedExpandableNodes().length === 0
      if (!fullyExpanded.value) {
        error.value = 'Some flow branches could not be expanded'
      }
      persistToCache()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to map full flow'
    } finally {
      mappingProgress.value = 100
      await new Promise((resolve) => setTimeout(resolve, 400))
      mappingFullFlow.value = false
      mappingProgress.value = 0
    }
  }

  function activateSymbol(sym: string, payload: GraphRootPayload): boolean {
    const cached = cache.get(payload.filePath, sym)
    if (!cached) return false
    error.value = null
    loading.value = false
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
    enrichError.value = null
    parentPath.value = []
    currentFilePath.value = payload.filePath
    currentWorkspace.value = payload.workspaceId
    symbol.value = payload.symbol
    lastPayload.value = payload
    enrichedIds.value = new Set()
    inFlightIds.value = new Set()
    fullyExpanded.value = false
    revealedIds.value = new Set()

    // Capture the symbol we are loading so we can detect if the user switched away.
    const requestedSymbol = payload.symbol

    try {
      const data = await fetchGraphRoot(payload)

      // A newer symbol was activated while this request was in-flight — discard.
      if (symbol.value !== requestedSymbol) {
        loading.value = false
        return
      }

      const pruned = pruneGraphToRoot(data.rootId, data.nodes, data.edges)
      allNodes.value = pruned.nodes
      allEdges.value = pruned.edges
      rootId.value = data.rootId
      isMock.value = Boolean(data.mock) && data.nodes.length === 0

      if (data.rootId) {
        revealedIds.value = entryOnlyRevealedIds(data.rootId, allNodes.value, allEdges.value)
      }
      syncVisible()
      loading.value = false
      persistToCache()
      const visibleIds = [...revealedIds.value]
      void enrichNodes(visibleIds, { background: false, priorityVisible: true })
      maintainSilentBuffer()
    } catch (err) {
      if (symbol.value === requestedSymbol) {
        error.value = err instanceof Error ? err.message : 'Failed to load graph'
        allNodes.value = []
        allEdges.value = []
        nodes.value = []
        edges.value = []
      }
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
    const path = [nodeId]

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
        void enrichNodes(newIds, { background: true, priorityVisible: true })
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
    mappingProgress.value = 0
    parentPath.value = []
    lastPayload.value = null
  }

  function stripEnrichedLabels(): void {
    for (const node of allNodes.value) {
      if (node.labelSource !== 'ai' && node.labelSource !== 'heuristic') continue
      delete node.title
      node.summary = node.code?.trim() || node.label || ''
      delete node.labelSource
      node.confidence = 'verified'
    }
    enrichedIds.value = new Set()
    inFlightIds.value = new Set()
  }

  async function reapplyExperience(payload: GraphRootPayload): Promise<void> {
    if (!symbol.value || allNodes.value.length === 0) return
    lastPayload.value = payload
    enrichError.value = null
    stripEnrichedLabels()
    syncVisible()
    persistToCache()
    const allIds = allNodes.value.map((n) => n.id)
    await enrichNodes(allIds, { background: false, priorityVisible: true })
  }

  return {
    nodes,
    edges,
    rootId,
    symbol,
    loading,
    enrichError,
    enriching,
    expanding,
    mappingFullFlow,
    mappingProgress,
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
    reapplyExperience,
  }
}
