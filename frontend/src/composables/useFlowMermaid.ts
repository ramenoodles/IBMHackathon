import { type MaybeRefOrGetter, ref, toValue, watch } from 'vue'
import type { FlowEdge, FlowNode } from '@/types/flowGraph'
import { compileMermaidSvg } from '@/composables/useMermaid'

/** Mermaid reserved words that break when used as node ids. */
const MERMAID_RESERVED = new Set([
  'end',
  'subgraph',
  'graph',
  'flowchart',
  'click',
  'class',
  'style',
  'linkstyle',
  'direction',
  'if',
  'else',
  'elseif',
  'then',
  'for',
  'while',
  'switch',
  'case',
  'default',
])

const STRUCTURAL_RENDER_DEBOUNCE_MS = 150
const MERMAID_STYLE_CLASSES = ['verified', 'inferred', 'heuristic', 'collapsed'] as const

export type FlowRenderReason = 'structure' | 'label'

/**
 * Escape text for Mermaid quoted node labels ("...").
 */
export function escapeMermaidLabel(text: string): string {
  return text
    .replace(/\r?\n/g, ' ')
    .replace(/"/g, "'")
}

export function mermaidNodeId(index: number): string {
  return `n${index}`
}

export function mermaidId(id: string): string {
  const base = id.replace(/[^a-zA-Z0-9_]/g, '_')
  if (!base || MERMAID_RESERVED.has(base.toLowerCase()) || /^\d/.test(base)) {
    return `node_${base || 'x'}`
  }
  return base
}

export function nodeDisplayTitle(node: FlowNode): string {
  if (node.title) return node.title
  const s = node.summary?.trim()
  if (s && !looksLikeCode(s)) return s
  return node.label.replace(/^L\d+\s+/, '')
}

/** Raw scan label from the CFG — ignores enriched title/summary. */
export function nodeScanTitle(node: FlowNode): string {
  return node.label.replace(/^L\d+\s+/, '')
}

export type MermaidLabelMode = 'display' | 'scan'

function nodeMermaidLabel(node: FlowNode, labelMode: MermaidLabelMode): string {
  return labelMode === 'scan' ? nodeScanTitle(node) : nodeDisplayTitle(node)
}

function looksLikeCode(text: string): boolean {
  return /[{}"'`;=]/.test(text) || text.startsWith('return ')
}

export function nodeMermaidClasses(node: FlowNode): string[] {
  if (node.collapsed) return ['collapsed']
  switch (node.labelSource) {
    case 'ai':
      return ['inferred']
    case 'heuristic':
      return ['heuristic']
    default:
      return ['verified']
  }
}

/** Stable key for graph topology — ids, kinds, compact state, edges. */
export function graphStructureKey(nodes: FlowNode[], edges: FlowEdge[]): string {
  const nodePart = nodes
    .map((n) => `${n.id}:${n.kind}:${n.collapsed ? 1 : 0}`)
    .join('|')
  const edgePart = edges.map((e) => `${e.from}->${e.to}:${e.label ?? ''}`).join('|')
  return `${nodePart};;${edgePart}`
}

/** Stable key for display labels only. */
export function graphLabelKey(nodes: FlowNode[]): string {
  return nodes
    .map((n) => `${n.id}:${nodeDisplayTitle(n)}:${n.labelSource ?? 'scan'}`)
    .join('|')
}

export function compileToMermaid(
  nodes: FlowNode[],
  edges: FlowEdge[],
  labelMode: MermaidLabelMode = 'display',
): string {
  if (nodes.length === 0) return ''

  const indexById = new Map(nodes.map((n, i) => [n.id, i]))

  const lines = [
    'flowchart TD',
    'classDef verified fill:#1a2e1a,stroke:#4ade80,color:#e2e8f0',
    'classDef inferred fill:#2a1f0a,stroke:#fbbf24,color:#e2e8f0,stroke-dasharray:5 5',
    'classDef heuristic fill:#0c2a2e,stroke:#22d3ee,color:#e2e8f0',
    'classDef collapsed fill:#1e293b,stroke:#ff3366,color:#f8fafc',
  ]

  for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i]!
    const mid = mermaidNodeId(i)
    const safe = escapeMermaidLabel(nodeMermaidLabel(node, labelMode))
    if (node.kind === 'branch') {
      lines.push(`  ${mid}{"${safe}"}`)
    } else {
      lines.push(`  ${mid}["${safe}"]`)
    }

    const classes = nodeMermaidClasses(node)
    lines.push(`  class ${mid} ${classes.join(',')}`)
  }

  for (const edge of edges) {
    const fromIdx = indexById.get(edge.from)
    const toIdx = indexById.get(edge.to)
    if (fromIdx === undefined || toIdx === undefined) continue
    const from = mermaidNodeId(fromIdx)
    const to = mermaidNodeId(toIdx)
    if (edge.label) {
      const edgeLabel = escapeMermaidLabel(edge.label)
      lines.push(`  ${from} -->|"${edgeLabel}"| ${to}`)
    } else {
      lines.push(`  ${from} --> ${to}`)
    }
  }

  return lines.join('\n')
}

function findNodeGroup(container: HTMLElement, nodes: FlowNode[], nodeId: string): SVGGElement | null {
  const idx = nodes.findIndex((n) => n.id === nodeId)
  if (idx >= 0) {
    const byIdx = container.querySelector(`g[id^="flowchart-n${idx}-"]`)
    if (byIdx) return byIdx as SVGGElement
  }
  const legacy = nodes.find((n) => n.id === nodeId)
  if (!legacy) return null
  for (const g of container.querySelectorAll('g.node')) {
    if (g.id.includes(mermaidId(legacy.id))) return g as SVGGElement
  }
  return null
}

function applyNodeStyleClasses(group: SVGGElement, classes: string[]): void {
  const targets = [group, ...group.querySelectorAll('rect, polygon, path, circle, ellipse')]
  for (const el of targets) {
    for (const cls of MERMAID_STYLE_CLASSES) el.classList.remove(cls)
    for (const cls of classes) el.classList.add(cls)
  }
}

/**
 * Update box labels and style classes without re-running Mermaid.
 */
export function applyLabelPatches(container: HTMLElement | null, nodes: FlowNode[]): void {
  if (!container) return
  for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i]!
    const group = container.querySelector<SVGGElement>(`g[id^="flowchart-n${i}-"]`)
    if (!group) continue

    const title = nodeDisplayTitle(node)
    const labelEl =
      group.querySelector<HTMLElement>('.nodeLabel') ??
      group.querySelector<HTMLElement>('foreignObject div') ??
      group.querySelector<HTMLElement>('.label')
    if (labelEl) labelEl.textContent = title

    applyNodeStyleClasses(group, nodeMermaidClasses(node))
  }
}

export function applySelectionHighlight(
  container: HTMLElement | null,
  nodes: FlowNode[],
  selectedNodeId: string,
): void {
  if (!container) return
  for (const g of container.querySelectorAll('g.node.is-selected')) {
    g.classList.remove('is-selected')
  }
  if (!selectedNodeId) return
  const group = findNodeGroup(container, nodes, selectedNodeId)
  group?.classList.add('is-selected')
}

function attachNodeClicks(
  container: HTMLElement,
  nodes: FlowNode[],
  onNodeClick: (node: FlowNode) => void,
): void {
  for (const g of container.querySelectorAll('g.node')) {
    const idxMatch = g.id.match(/flowchart-n(\d+)-/)
    let node = idxMatch ? nodes[Number(idxMatch[1])] : undefined
    if (!node) {
      node = nodes.find((n) => g.id.includes(mermaidId(n.id)))
    }
    if (!node) continue
    bindClick(g as SVGGElement, node, onNodeClick)
  }
}

function bindClick(
  el: SVGGElement,
  node: FlowNode,
  onNodeClick: (node: FlowNode) => void,
): void {
  el.style.cursor = 'pointer'
  el.onclick = (e) => {
    e.stopPropagation()
    onNodeClick(node)
  }
}

export function useFlowMermaid(
  nodes: MaybeRefOrGetter<FlowNode[]>,
  edges: MaybeRefOrGetter<FlowEdge[]>,
  selectedNodeId: MaybeRefOrGetter<string> = '',
  onNodeClick?: (node: FlowNode) => void,
  onRender?: (reason: FlowRenderReason) => void,
  labelMode: MermaidLabelMode = 'display',
) {
  const mermaidCode = ref('')
  const renderError = ref<string | null>(null)
  const containerRef = ref<HTMLElement | null>(null)
  let structuralTimer: ReturnType<typeof setTimeout> | null = null
  let renderGeneration = 0
  let lastStructureKey = ''
  let lastLabelKey = ''

  async function renderStructural(): Promise<void> {
    // Cancel any pending debounce — this call supersedes it.
    cancelStructuralRender()
    // Bump the generation so any older in-flight render knows to discard its result.
    const generation = ++renderGeneration
    const edgeList = toValue(edges)
    // Re-read nodes at render time (not at schedule time) so any enrichment
    // patches applied during the debounce window are included in the first SVG.
    const latestNodeList = toValue(nodes)
    mermaidCode.value = compileToMermaid(latestNodeList, edgeList, labelMode)
    if (!containerRef.value || !mermaidCode.value) return
    try {
      renderError.value = null
      const svg = await compileMermaidSvg(mermaidCode.value)
      // A newer render started while we were awaiting Mermaid — discard entirely.
      if (generation !== renderGeneration || !containerRef.value) return
      containerRef.value.innerHTML = svg
      // Re-read nodes AFTER the await so any enrichment patches that arrived
      // while Mermaid was compiling are included in the label pass.  Using a
      // stale snapshot here would overwrite freshly-enriched labels and leave
      // lastLabelKey pointing at the unenriched key, causing the chart to
      // permanently show raw function names for the current symbol.
      const currentNodeList = toValue(nodes)
      const currentEdgeList = toValue(edges)
      applyLabelPatches(containerRef.value, currentNodeList)
      if (onNodeClick) attachNodeClicks(containerRef.value, currentNodeList, onNodeClick)
      applySelectionHighlight(containerRef.value, currentNodeList, toValue(selectedNodeId))
      // Sync both keys so subsequent syncFromGraph calls correctly classify
      // any further changes as label-only rather than structural.
      lastStructureKey = graphStructureKey(currentNodeList, currentEdgeList)
      lastLabelKey = graphLabelKey(currentNodeList)
      onRender?.('structure')
    } catch (err) {
      if (generation === renderGeneration) {
        renderError.value = err instanceof Error ? err.message : 'Render failed'
      }
    }
  }

  function cancelStructuralRender(): void {
    if (structuralTimer) {
      clearTimeout(structuralTimer)
      structuralTimer = null
    }
  }

  function scheduleStructuralRender(): void {
    cancelStructuralRender()
    structuralTimer = setTimeout(() => {
      structuralTimer = null
      void renderStructural()
    }, STRUCTURAL_RENDER_DEBOUNCE_MS)
  }

  function syncFromGraph(): void {
    const nodeList = toValue(nodes)
    const edgeList = toValue(edges)
    const structureKey = graphStructureKey(nodeList, edgeList)
    const labelKey = graphLabelKey(nodeList)

    if (structureKey !== lastStructureKey) {
      lastStructureKey = structureKey
      // Don't update lastLabelKey here — renderStructural will set it after
      // reading the latest labels, so a concurrent enrichment patch during the
      // debounce window doesn't cause a spurious second applyLabelPatches call.
      scheduleStructuralRender()
      return
    }

    if (labelKey !== lastLabelKey) {
      // Only commit the new label key if the container exists and the patch
      // can actually be applied.  If the SVG hasn't been written yet (container
      // is null or the structural render is still in-flight), leave lastLabelKey
      // dirty so that the next syncFromGraph call — or the structural render
      // itself — will pick it up.
      if (!containerRef.value) return
      lastLabelKey = labelKey
      applyLabelPatches(containerRef.value, nodeList)
      applySelectionHighlight(containerRef.value, nodeList, toValue(selectedNodeId))
      onRender?.('label')
    }
  }

  watch(
    [() => toValue(nodes), () => toValue(edges)],
    () => {
      syncFromGraph()
    },
    { deep: true },
  )

  watch(
    () => toValue(selectedNodeId),
    (id) => {
      const nodeList = toValue(nodes)
      applySelectionHighlight(containerRef.value, nodeList, id)
    },
  )

  function setContainer(el: HTMLElement | null): void {
    containerRef.value = el
    lastStructureKey = ''
    lastLabelKey = ''
    void renderStructural()
  }

  function resetKeys(): void {
    lastStructureKey = ''
    lastLabelKey = ''
  }

  return { mermaidCode, renderError, setContainer, renderStructural, resetKeys }
}
