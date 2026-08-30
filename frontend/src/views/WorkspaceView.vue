<script setup lang="ts">
/**
 * Clean graph-first workspace: explorer + timeline + on-demand source modal.
 */
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Sidebar from '@/components/workspace/Sidebar.vue'
import SymbolBar from '@/components/workspace/SymbolBar.vue'
import FlowCanvas from '@/components/workspace/FlowCanvas.vue'
import FlowWarmOverlay from '@/components/workspace/FlowWarmOverlay.vue'
import BranchPrompt from '@/components/workspace/BranchPrompt.vue'
import CompactFlowPreview from '@/components/workspace/CompactFlowPreview.vue'
import Modal from '@/components/ui/Modal.vue'
import CodePanel from '@/components/workspace/CodePanel.vue'
import ResizeHandle from '@/components/ui/ResizeHandle.vue'
import { useWorkspaceLayout } from '@/composables/useWorkspaceLayout'
import { useHorizontalResize } from '@/composables/usePanelResize'
import { useFlowGraphCache } from '@/composables/useFlowGraphCache'
import { useFlowGraph } from '@/composables/useFlowGraph'
import { useSymbolBrief } from '@/composables/useSymbolBrief'
import { useFileFlowWarm } from '@/composables/useFileFlowWarm'
import { useNodeDetail } from '@/composables/useNodeDetail'
import { userContext, normalizeLanguage } from '@/store/userContext'
import type { FlowNode } from '@/types/flowGraph'

type WorkspacePhase = 'idle' | 'tracing'

const {
  sidebarOpen,
  isMobile,
  explorerWidth,
  EXPLORER_MIN,
  EXPLORER_MAX,
  persistWidths,
  toggleSidebar,
} = useWorkspaceLayout()

const resizeEnabled = computed(() => !isMobile.value)

const { onPointerDown: onExplorerResize } = useHorizontalResize({
  width: explorerWidth,
  min: EXPLORER_MIN,
  max: EXPLORER_MAX,
  side: 'right',
  enabled: resizeEnabled,
  onEnd: persistWidths,
})

const graphCache = useFlowGraphCache()
const {
  nodes,
  edges,
  rootId,
  loading,
  enriching,
  enrichError,
  expanding,
  mappingFullFlow,
  mappingProgress,
  fullyExpanded,
  error: graphError,
  isMock,
  loadRoot,
  activateSymbol,
  expandNode,
  revealFromNode,
  revealFullFlow,
  hasHiddenChildren,
  prefetchAroundNode,
  reset: resetGraph,
} = useFlowGraph(graphCache)

const {
  symbols,
  loading: symbolsLoading,
  error: symbolsError,
  load: loadSymbols,
  reset: resetSymbols,
} = useSymbolBrief()
const { warming, progress } = useFileFlowWarm(graphCache)
const { detail, loading: detailLoading, streaming: detailStreaming, error: detailError, loadDetail, clear: clearDetail } = useNodeDetail()

const workspacePhase = ref<WorkspacePhase>('idle')
const selectedPath = ref('')
const symbol = ref('')
const selectedNodeId = ref('')
const sourceOpen = ref(false)
const sourcePath = ref('')
const sourceLine = ref<number | undefined>(undefined)
const branchPromptOpen = ref(false)
const branchNode = ref<FlowNode | null>(null)
const compactPreviewOpen = ref(false)
const compactPreviewNode = ref<FlowNode | null>(null)

const router = useRouter()
const leaveConfirmOpen = ref(false)

const showSymbolBar = computed(
  () => !!(selectedPath.value && workspacePhase.value === 'tracing'),
)

const allSymbolNames = computed(() => symbols.value.map((s) => s.name))

function graphPayload() {
  return {
    workspaceId: userContext.value.workspaceId,
    filePath: selectedPath.value,
    symbol: symbol.value,
    userContext: { ...userContext.value },
  }
}

async function onSelectFile(path: string): Promise<void> {
  selectedPath.value = path
  symbol.value = ''
  selectedNodeId.value = ''
  clearDetail()
  resetSymbols()
  resetGraph()
  workspacePhase.value = 'tracing'
  await loadSymbols(userContext.value.workspaceId, path)
  const first = symbols.value[0]
  if (first) void onPickSymbol(first.name)
}

async function onPickSymbol(name: string): Promise<void> {
  if (!selectedPath.value) return
  symbol.value = name
  selectedNodeId.value = ''
  clearDetail()
  const payload = { ...graphPayload(), symbol: name }
  if (!activateSymbol(name, payload)) {
    await loadRoot(payload)
  }
}

function onRevealNode(node: FlowNode): void {
  revealFromNode(node.id)
}

function onSelectNode(node: FlowNode): void {
  selectedNodeId.value = node.id
  prefetchAroundNode(node.id)
}

// Auto-select the root node whenever a new graph loads
watch(rootId, (id) => {
  if (!id) return
  const root = nodes.value.find((n) => n.id === id)
  if (root) onSelectNode(root)
})

function onRequestDetail(node: FlowNode): void {
  void loadDetail(
    {
       workspaceId: userContext.value.workspaceId,
      nodeId: node.id,
      symbol: symbol.value,
      file: node.file ?? selectedPath.value,
      line: node.line,
      title: node.title ?? node.label,
      confidence: node.confidence,
      code: node.code,
      kind: node.kind,
      summary: node.summary,
      experience: userContext.value.experienceLevel,
      language: normalizeLanguage(userContext.value.primaryLanguage),
    },
    { stream: true },
  )
}

function onViewSource(file?: string, line?: number): void {
  sourcePath.value = file ?? selectedPath.value
  sourceLine.value = line
  sourceOpen.value = true
}

async function onGoToDefinition(file: string, symbolName: string, line?: number): Promise<void> {
if (!file || !symbolName) return

sourceOpen.value = false
sourcePath.value = ''
sourceLine.value = undefined

if (selectedPath.value === file) {
  await onPickSymbol(symbolName)
  return
}

selectedPath.value = file
symbol.value = ''
selectedNodeId.value = ''
clearDetail()
resetGraph()
workspacePhase.value = 'tracing'
await loadSymbols(userContext.value.workspaceId, file)

const nextSymbol = symbols.value.find((s) => s.name === symbolName) ?? symbols.value[0]
if (!nextSymbol) return
await onPickSymbol(nextSymbol.name)
}

function onExpandNode(node: FlowNode): void {
  if (node.childCount > 3) {
    branchNode.value = node
    branchPromptOpen.value = true
  } else {
    void expandNode(node.id, graphPayload(), node.childCount ? node.childCount + 1 : 3)
  }
}

function onConfirmExpand(limit: number): void {
  if (!branchNode.value) return
  void expandNode(branchNode.value.id, graphPayload(), limit)
  branchNode.value = null
}

function onShowFullFlow(): void {
  void revealFullFlow(graphPayload())
}

function onPreviewCompacted(node: FlowNode): void {
  compactPreviewNode.value = node
  compactPreviewOpen.value = true
}

function fileName(): string {
  return selectedPath.value.split('/').pop() ?? selectedPath.value
}
</script>

<template>
  <div class="flex h-screen flex-col overflow-hidden bg-slate-950">
    <header class="flex items-center gap-3 border-b border-slate-800 bg-slate-900 px-4 py-2.5">
      <button
        type="button"
        class="rounded p-1 text-slate-400 hover:bg-slate-800 hover:text-white"
        aria-label="Toggle files"
        @click="toggleSidebar"
      >
        <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" aria-hidden="true">
          <path d="M2 4h12M2 8h12M2 12h12" />
        </svg>
      </button>
      <button
        type="button"
        class="cursor-pointer transition-opacity hover:opacity-70"
        title="Go home"
        @click="leaveConfirmOpen = true"
      >
        <img src="@/assets/dark_headline.png" alt="OnBober" class="h-6 w-auto" />
      </button>
      <span v-if="selectedPath" class="truncate text-xs text-slate-500">
        {{ fileName() }}
      </span>
    </header>

    <div class="flex min-h-0 flex-1">
      <Sidebar
        :workspace-id="userContext.workspaceId"
        :selected-path="selectedPath"
        :open="sidebarOpen"
        :width="explorerWidth"
        @select="onSelectFile"
        @toggle="toggleSidebar"
      />
      <ResizeHandle
        v-if="sidebarOpen && !isMobile"
        label="Resize file explorer"
        @pointerdown="onExplorerResize"
      />

      <div class="relative flex min-w-0 flex-1 flex-col">
        <SymbolBar
          :workspace-id="userContext.workspaceId"
          :file-path="selectedPath"
          :active-symbol="symbol"
          :all-symbols="allSymbolNames"
          :visible="showSymbolBar"
          @pick="onPickSymbol"
        />

        <Transition name="fade" mode="out-in">
          <FlowCanvas
            v-if="workspacePhase === 'tracing'"
            key="tracing"
            :nodes="nodes"
            :edges="edges"
            :root-id="rootId"
            :loading="loading"
            :enriching="enriching"
            :enrich-error="enrichError"
            :expanding="expanding"
            :mapping-full-flow="mappingFullFlow"
            :mapping-progress="mappingProgress"
            :fully-expanded="fullyExpanded"
            :error="graphError"
            :is-mock="isMock"
            :symbol="symbol"
            :selected-node-id="selectedNodeId"
            :detail="detail"
            :detail-loading="detailLoading"
            :detail-streaming="detailStreaming"
            :detail-error="detailError"
            :has-hidden-children="hasHiddenChildren"
            @select-node="onSelectNode"
            @request-detail="onRequestDetail"
            @reveal-node="onRevealNode"
            @expand-node="onExpandNode"
            @show-full-flow="onShowFullFlow"
            @view-source="onViewSource"
            @go-to-definition="onGoToDefinition"
            @preview-compacted="onPreviewCompacted"
          />

          <div
            v-else
            key="idle"
            class="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-center text-slate-500"
          >
            <p class="text-lg font-medium text-slate-300">Select a file to begin</p>
            <p class="max-w-sm text-sm">Choose a file from the explorer to see traceable symbols.</p>
          </div>
        </Transition>
      </div>
    </div>

    <Modal :open="leaveConfirmOpen" title="Leave workspace?" @close="leaveConfirmOpen = false">
      <p class="mb-5 text-sm text-slate-300">Are you sure you would like to leave the workspace?</p>
      <div class="flex justify-end gap-2">
        <button
          type="button"
          class="rounded-md border border-slate-700 px-3 py-1.5 text-sm text-slate-300 transition hover:border-slate-500 hover:text-white"
          @click="leaveConfirmOpen = false"
        >
          Return to workspace
        </button>
        <button
          type="button"
          class="rounded-md bg-onbober-primary px-3 py-1.5 text-sm font-medium text-white transition hover:opacity-90"
          @click="router.push('/')"
        >
          Yes, go home
        </button>
      </div>
    </Modal>

    <Modal :open="sourceOpen" title="Source" size="xl" @close="sourceOpen = false">
      <CodePanel
        v-if="sourcePath || selectedPath"
        :workspace-id="userContext.workspaceId"
        :file-path="sourcePath || selectedPath"
        :highlight-line="sourceLine"
      />
    </Modal>

    <BranchPrompt
      :open="branchPromptOpen"
      :node-label="branchNode?.label ?? ''"
      :child-count="branchNode?.childCount ?? 0"
      @close="branchPromptOpen = false"
      @expand="onConfirmExpand"
    />

    <CompactFlowPreview
      :open="compactPreviewOpen"
      :workspace-id="userContext.workspaceId"
      :file-path="compactPreviewNode?.calleeFile ?? ''"
      :symbol="compactPreviewNode?.calleeSymbol ?? ''"
      :line="compactPreviewNode?.calleeLine"
      :user-context="userContext"
      @close="compactPreviewOpen = false"
    />
  </div>
</template>

<style scoped>
/* Phase content: simple fade */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
