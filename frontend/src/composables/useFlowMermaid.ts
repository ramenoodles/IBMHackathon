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

/**
 * Escape text for Mermaid quoted node labels ("...").
 * Inside a double-quoted label the only characters that break parsing are
 * double-quotes (which close the string) and raw newlines.
 * Everything else — <, >, [, ], #, ; — is safe to pass through as-is.
 * @param text - Raw label text.
 */
export function escapeMermaidLabel(text: string): string {
  return text
    .replace(/\r?\n/g, ' ')    // collapse newlines to a space
    .replace(/"/g, "'")         // swap " for ' to avoid closing the label string
}

/**
 * Build stable numeric Mermaid node ids (avoids reserved words like `if`).
 * @param index - Node index in the graph list.
 */
export function mermaidNodeId(index: number): string {
  return `n${index}`
}

/**
 * Legacy sanitizer kept for click-handler matching on older SVG ids.
 * @param id - Raw node id.
 */
export function mermaidId(id: string): string {
  const base = id.replace(/[^a-zA-Z0-9_]/g, '_')
  if (!base || MERMAID_RESERVED.has(base.toLowerCase()) || /^\d/.test(base)) {
    return `node_${base || 'x'}`
  }
  return base
}

/**
 * Plain label for graph boxes — prefer LLM title, never raw code literals.
 */
export function nodeDisplayTitle(node: FlowNode): string {
  if (node.title) return node.title
  const s = node.summary?.trim()
  if (s && !looksLikeCode(s)) return s
  return node.label.replace(/^L\d+\s+/, '')
}

function looksLikeCode(text: string): boolean {
  return /[{}"'`;=]/.test(text) || text.startsWith('return ')
}

/**
 * Compile flow graph nodes and edges into Mermaid flowchart syntax.
 * Selection highlighting is applied via DOM after render — not in DSL.
 */
export function compileToMermaid(nodes: FlowNode[], edges: FlowEdge[]): string {
  if (nodes.length === 0) return ''

  const indexById = new Map(nodes.map((n, i) => [n.id, i]))

  const lines = [
    'flowchart TD',
    'classDef verified fill:#1a2e1a,stroke:#4ade80,color:#e2e8f0',
    'classDef inferred fill:#2a1f0a,stroke:#fbbf24,color:#e2e8f0,stroke-dasharray:5 5',
    'classDef collapsed fill:#1e293b,stroke:#ff3366,color:#f8fafc',
  ]

  for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i]!
    const mid = mermaidNodeId(i)
    const display = nodeDisplayTitle(node)
    const label = node.collapsed ? `${display} (+${node.childCount})` : display
    const safe = escapeMermaidLabel(label)
    if (node.kind === 'branch') {
      lines.push(`  ${mid}{"${safe}"}`)
    } else {
      lines.push(`  ${mid}["${safe}"]`)
    }

    const classes: string[] = [node.confidence]
    if (node.collapsed) classes.push('collapsed')
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

/**
 * Highlight the selected node in the rendered SVG without re-rendering Mermaid.
 */
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

/**
 * Wire click handlers on rendered Mermaid node groups.
 */
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

/**
 * Reactive Mermaid renderer for flow graphs with node click support.
 */
export function useFlowMermaid(
  nodes: MaybeRefOrGetter<FlowNode[]>,
  edges: MaybeRefOrGetter<FlowEdge[]>,
  selectedNodeId: MaybeRefOrGetter<string> = '',
  onNodeClick?: (node: FlowNode) => void,
  onStructuralRender?: () => void,
) {
  const mermaidCode = ref('')
  const renderError = ref<string | null>(null)
  const containerRef = ref<HTMLElement | null>(null)
  let structuralTimer: ReturnType<typeof setTimeout> | null = null

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
      onStructuralRender?.()
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

  watch(
    [() => toValue(nodes), () => toValue(edges)],
    () => {
      scheduleStructuralRender()
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
    void renderStructural()
  }

  return { mermaidCode, renderError, setContainer, renderStructural }
}
