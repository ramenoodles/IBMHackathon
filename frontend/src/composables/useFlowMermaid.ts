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

/**
 * Escape text for Mermaid quoted node labels.
 * @param text - Raw label text.
 */
export function escapeMermaidLabel(text: string): string {
  return text
    .replace(/\r?\n/g, ' ')
    .replace(/\\/g, '\\\\')
    .replace(/#/g, '#35;')
    .replace(/"/g, '#quot;')
    .replace(/\[/g, '#91;')
    .replace(/\]/g, '#93;')
    .replace(/</g, '#lt;')
    .replace(/>/g, '#gt;')
    .replace(/;/g, '#59;')
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
 * @param nodes - Graph nodes.
 * @param edges - Graph edges.
 * @param selectedNodeId - Currently selected node for highlight.
 */
export function compileToMermaid(
  nodes: FlowNode[],
  edges: FlowEdge[],
  selectedNodeId = '',
): string {
  if (nodes.length === 0) return ''

  const indexById = new Map(nodes.map((n, i) => [n.id, i]))

  const lines = [
    'flowchart TD',
    'classDef verified fill:#1a2e1a,stroke:#4ade80,color:#e2e8f0',
    'classDef inferred fill:#2a1f0a,stroke:#fbbf24,color:#e2e8f0,stroke-dasharray:5 5',
    'classDef collapsed fill:#1e293b,stroke:#ff3366,color:#f8fafc',
    'classDef selected fill:#3b0d1a,stroke:#ff3366,color:#fff,stroke-width:3px',
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

    const classes: string[] = []
    if (selectedNodeId === node.id) {
      classes.push('selected')
    } else {
      classes.push(node.confidence)
      if (node.collapsed) classes.push('collapsed')
    }
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
) {
  const mermaidCode = ref('')
  const renderError = ref<string | null>(null)
  const containerRef = ref<HTMLElement | null>(null)

  async function render(): Promise<void> {
    const nodeList = toValue(nodes)
    const edgeList = toValue(edges)
    const selected = toValue(selectedNodeId)
    mermaidCode.value = compileToMermaid(nodeList, edgeList, selected)
    if (!containerRef.value || !mermaidCode.value) return
    try {
      renderError.value = null
      await renderMermaid(containerRef.value, mermaidCode.value)
      if (onNodeClick) attachNodeClicks(containerRef.value, nodeList, onNodeClick)
    } catch (err) {
      renderError.value = err instanceof Error ? err.message : 'Render failed'
    }
  }

  watch(
    [() => toValue(nodes), () => toValue(edges), () => toValue(selectedNodeId)],
    () => {
      void render()
    },
    { deep: true },
  )

  function setContainer(el: HTMLElement | null): void {
    containerRef.value = el
    void render()
  }

  return { mermaidCode, renderError, setContainer }
}
