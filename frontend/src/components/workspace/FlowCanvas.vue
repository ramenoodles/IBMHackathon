<script setup lang="ts">
/**
 * Flow graph canvas: Mermaid flowchart + source-anchored execution trace.
 */
import type { FlowEdge, FlowNode, NodeDetail } from '@/types/flowGraph'
import { computed, ref, watch } from 'vue'
import { useFlowMermaid } from '@/composables/useFlowMermaid'

const props = defineProps<{
  nodes: FlowNode[]
  edges: FlowEdge[]
  rootId: string
  loading: boolean
  expanding: boolean
  error: string | null
  isMock: boolean
  symbol: string
  selectedNodeId: string
  detail: NodeDetail | null
  detailLoading: boolean
}>()

const emit = defineEmits<{
  selectNode: [node: FlowNode]
  expandNode: [node: FlowNode]
  viewSource: []
}>()

const mermaidContainer = ref<HTMLElement | null>(null)

function onNodeClick(node: FlowNode): void {
  if (node.collapsed && node.expandable) {
    emit('expandNode', node)
  } else {
    emit('selectNode', node)
  }
}

const { renderError, setContainer } = useFlowMermaid(
  () => props.nodes,
  () => props.edges,
  () => props.selectedNodeId,
  onNodeClick,
)

watch(mermaidContainer, (el) => setContainer(el))

/** Nodes ordered from root following edges, then by line number. */
const orderedNodes = computed(() => {
  if (!props.nodes.length) return []
  const byId = new Map(props.nodes.map((n) => [n.id, n]))
  const used = new Set<string>()
  const ordered: FlowNode[] = []

  let startId = props.rootId || props.nodes[0]?.id
  let current: FlowNode | undefined = startId ? byId.get(startId) : props.nodes[0]
  while (current && !used.has(current.id)) {
    ordered.push(current)
    used.add(current.id)
    const out = props.edges.find((e) => e.from === current!.id)
    current = out ? byId.get(out.to) : undefined
  }

  const rest = props.nodes.filter((n) => !used.has(n.id))
  rest.sort((a, b) => (a.line ?? 0) - (b.line ?? 0))
  return [...ordered, ...rest]
})

const selectedNode = computed(() => props.nodes.find((n) => n.id === props.selectedNodeId))

function kindLabel(kind: string): string {
  const labels: Record<string, string> = {
    entry: 'entry',
    call: 'call',
    branch: 'branch',
    return: 'return',
    assign: 'assign',
  }
  return labels[kind] ?? kind
}
</script>

<template>
  <div class="flex h-full min-w-0 flex-1">
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <div v-if="isMock" class="shrink-0 bg-amber-900/20 px-4 py-1.5 text-xs text-amber-300">
        Demo mode — Ollama offline
      </div>

      <div v-if="loading" class="flex flex-1 items-center justify-center text-sm text-slate-500">
        Mapping execution path...
      </div>

      <div v-else-if="!symbol" class="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-center">
        <p class="text-lg font-medium text-slate-300">Pick a function to trace</p>
        <p class="max-w-sm text-sm text-slate-500">Select a file, then choose a function from the bar above.</p>
      </div>

      <div v-else-if="!nodes.length" class="flex flex-1 items-center justify-center text-sm text-slate-500">
        No flow data for <span class="ml-1 font-mono text-onbober-primary">{{ symbol }}</span>
      </div>

      <template v-else>
        <div class="flex shrink-0 items-center justify-between border-b border-slate-800 px-4 py-2">
          <p class="text-xs text-slate-500">
            Tracing <span class="font-mono text-slate-300">{{ symbol }}</span>
            <span v-if="expanding" class="ml-2 text-onbober-primary">Expanding...</span>
          </p>
          <div class="flex items-center gap-3 text-[10px] text-slate-500">
            <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-sm border border-green-500 bg-green-900/40" /> verified</span>
            <span class="flex items-center gap-1"><span class="inline-block h-2 w-2 rounded-sm border border-dashed border-amber-500 bg-amber-900/40" /> inferred</span>
          </div>
        </div>

        <p v-if="error" class="shrink-0 px-4 py-2 text-sm text-red-400">{{ error }}</p>

        <div class="flex min-h-0 flex-1">
          <!-- Source-anchored trace list -->
          <ol class="w-72 shrink-0 overflow-y-auto border-r border-slate-800 bg-slate-900/40 p-2">
            <li v-for="(node, i) in orderedNodes" :key="node.id" class="mb-1">
              <button
                type="button"
                class="w-full rounded-md border px-2 py-1.5 text-left transition"
                :class="
                  selectedNodeId === node.id
                    ? 'border-onbober-primary/50 bg-onbober-primary/5'
                    : 'border-transparent hover:border-slate-700 hover:bg-slate-800/50'
                "
                @click="onNodeClick(node)"
              >
                <div class="flex items-center gap-2">
                  <span
                    class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-[10px] font-bold"
                    :class="
                      selectedNodeId === node.id
                        ? 'bg-onbober-primary text-white'
                        : 'bg-slate-800 text-slate-400'
                    "
                  >
                    {{ i + 1 }}
                  </span>
                  <span class="min-w-0 flex-1 truncate text-xs font-medium text-slate-200">{{ node.label }}</span>
                  <span class="shrink-0 text-[10px] uppercase text-slate-600">{{ kindLabel(node.kind) }}</span>
                </div>
                <p v-if="node.code" class="mt-1 truncate font-mono text-[10px] text-slate-500">{{ node.code }}</p>
                <p v-else class="mt-1 truncate text-[10px] text-slate-500">{{ node.summary }}</p>
              </button>
            </li>
          </ol>

          <!-- Flowchart -->
          <div class="min-w-0 flex-1 overflow-auto bg-slate-950/50 p-4">
            <div
              ref="mermaidContainer"
              class="mermaid-flow mx-auto flex min-h-[180px] min-w-fit items-center justify-center"
            />
            <p v-if="renderError" class="mt-2 text-center text-xs text-red-400">{{ renderError }}</p>
          </div>
        </div>
      </template>
    </div>

    <aside
      v-if="selectedNodeId && (detail || detailLoading || selectedNode)"
      class="w-72 shrink-0 overflow-y-auto border-l border-slate-800 bg-slate-900 p-4"
    >
      <p v-if="detailLoading" class="text-sm text-slate-500">Loading...</p>
      <template v-else>
        <h3 class="font-semibold text-white">{{ selectedNode?.label ?? detail?.title }}</h3>
        <span
          v-if="selectedNode"
          class="mt-1 inline-block rounded px-1.5 py-0.5 text-[10px] font-bold uppercase"
          :class="selectedNode.confidence === 'verified' ? 'bg-green-900/40 text-green-400' : 'bg-amber-900/40 text-amber-400'"
        >
          {{ selectedNode.confidence }}
        </span>
        <pre
          v-if="selectedNode?.code"
          class="mt-3 overflow-x-auto rounded border border-slate-800 bg-slate-950 p-2 font-mono text-[11px] leading-relaxed text-slate-300"
        >{{ selectedNode.code }}</pre>
        <p v-if="detail" class="mt-3 text-sm leading-relaxed text-slate-300">{{ detail.explanation }}</p>
        <p v-else-if="selectedNode" class="mt-3 text-sm leading-relaxed text-slate-400">{{ selectedNode.summary }}</p>
        <button
          v-if="selectedNode?.file || detail?.file"
          type="button"
          class="mt-4 text-xs text-onbober-primary hover:underline"
          @click="emit('viewSource')"
        >
          View source{{ selectedNode?.line ? ` (line ${selectedNode.line})` : '' }}
        </button>
      </template>
    </aside>
  </div>
</template>

<style scoped>
.mermaid-flow :deep(svg) {
  max-width: 100%;
  height: auto;
}
</style>
