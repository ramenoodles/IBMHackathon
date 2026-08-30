import { type MaybeRefOrGetter, ref, toValue, watch } from 'vue'
import type { FlowEdge, FlowNode } from '@/types/flowGraph'
import { renderMermaid } from '@/composables/useMermaid'

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

const STRUCTURAL_RENDER_DEBOUNCE_MS = 75
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

export function compileToMermaid(nodes: FlowNode[], edges: FlowEdge[]): string {
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
    const safe = escapeMermaidLabel(nodeDisplayTitle(node))
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
) {
  const mermaidCode = ref('')
  const renderError = ref<string | null>(null)
  const containerRef = ref<HTMLElement | null>(null)
  let structuralTimer: ReturnType<typeof setTimeout> | null = null
  let lastStructureKey = ''
  let lastLabelKey = ''

  async function renderStructural(): Promise<void> {
    const nodeList = toValue(nodes)
    const edgeList = toValue(edges)
    mermaidCode.value = compileToMermaid(nodeList, edgeList)
    if (!containerRef.value || !mermaidCode.value) return
    try {
      renderError.value = null
      await renderMermaid(containerRef.value, mermaidCode.value)
      if (onNodeClick) attachNodeClicks(containerRef.value, nodeList, onNodeClick)
      applySelectionHighlight(containerRef.value, nodeList, toValue(selectedNodeId))
      onRender?.('structure')
    } catch (err) {
      renderError.value = err instanceof Error ? err.message : 'Render failed'
    }
  }

  function scheduleStructuralRender(): void {
    if (structuralTimer) clearTimeout(structuralTimer)
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
      lastLabelKey = labelKey
      scheduleStructuralRender()
      return
    }

    if (labelKey !== lastLabelKey) {
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
