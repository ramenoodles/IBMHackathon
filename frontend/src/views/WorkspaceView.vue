<script setup lang="ts">
/**
 * Clean graph-first workspace: explorer + timeline + on-demand source modal.
 */
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import Sidebar from '@/components/workspace/Sidebar.vue'
import SymbolBar from '@/components/workspace/SymbolBar.vue'
import FlowCanvas from '@/components/workspace/FlowCanvas.vue'
import FileFlowBrief from '@/components/workspace/FileFlowBrief.vue'
import FlowWarmOverlay from '@/components/workspace/FlowWarmOverlay.vue'
import BranchPrompt from '@/components/workspace/BranchPrompt.vue'
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
import { userContext } from '@/store/userContext'
import { SYMBOL_PAGE_SIZE } from '@/utils/flowGraphUtils'
import type { FlowNode } from '@/types/flowGraph'

type WorkspacePhase = 'idle' | 'brief' | 'warming' | 'tracing'

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
  expanding,
  mappingFullFlow,
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
  isLargeFile,
  load: loadSymbols,
  reset: resetSymbols,
} = useSymbolBrief()
const { warming, progress, warmFile } = useFileFlowWarm(graphCache)
const { detail, loading: detailLoading, streaming: detailStreaming, error: detailError, loadDetail, clear: clearDetail } = useNodeDetail()

const workspacePhase = ref<WorkspacePhase>('idle')
const selectedPath = ref('')
const symbol = ref('')
const selectedNodeId = ref('')
const sourceOpen = ref(false)
const sourcePath = ref('')
const warmedSymbolNames = ref<string[]>([])
const symbolBarPage = ref(0)

const branchPromptOpen = ref(false)
const branchNode = ref<FlowNode | null>(null)

const router = useRouter()
const leaveConfirmOpen = ref(false)

const showSymbolBar = computed(
  () =>
    selectedPath.value &&
    (workspacePhase.value === 'tracing' || workspacePhase.value === 'warming'),
)

const symbolBarSource = computed(() => {
  if (warmedSymbolNames.value.length) return warmedSymbolNames.value
  return symbols.value.map((s) => s.name)
})

const symbolBarTotalPages = computed(() =>
  Math.max(1, Math.ceil(symbolBarSource.value.length / SYMBOL_PAGE_SIZE)),
)

const symbolBarNames = computed(() => {
  const start = symbolBarPage.value * SYMBOL_PAGE_SIZE
  return symbolBarSource.value.slice(start, start + SYMBOL_PAGE_SIZE)
})

const symbolBarHasNext = computed(() => symbolBarPage.value < symbolBarTotalPages.value - 1)
const symbolBarHasPrev = computed(() => symbolBarPage.value > 0)

function symbolBarGoToPage(n: number): void {
  symbolBarPage.value = Math.max(0, Math.min(n, symbolBarTotalPages.value - 1))
}

function symbolBarAdvancePage(): void {
  if (symbolBarHasNext.value) symbolBarGoToPage(symbolBarPage.value + 1)
}

function symbolBarPrevPage(): void {
  if (symbolBarHasPrev.value) symbolBarGoToPage(symbolBarPage.value - 1)
}

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
  warmedSymbolNames.value = []
  symbolBarPage.value = 0
  resetSymbols()
  resetGraph()

  if (graphCache.isFileWarmed(path)) {
    warmedSymbolNames.value = graphCache.listSymbolsForFile(path)
    workspacePhase.value = 'tracing'
    if (warmedSymbolNames.value[0]) {
      symbol.value = warmedSymbolNames.value[0]
      activateSymbol(warmedSymbolNames.value[0], {
        workspaceId: userContext.value.workspaceId,
        filePath: path,
        symbol: warmedSymbolNames.value[0],
        userContext: { ...userContext.value },
      })
    }
    return
  }

  workspacePhase.value = 'brief'
  await loadSymbols(userContext.value.workspaceId, path)
}

function onDeclineFlowInit(): void {
  workspacePhase.value = 'tracing'
}

async function onConfirmFlowInit(selected: string[]): Promise<void> {
  if (!selectedPath.value || !selected.length) return
  workspacePhase.value = 'warming'
  warmedSymbolNames.value = selected
  const base = {
    workspaceId: userContext.value.workspaceId,
    filePath: selectedPath.value,
    userContext: { ...userContext.value },
  }
  await warmFile(base, selected)
  workspacePhase.value = 'tracing'
  symbol.value = selected[0]!
  selectedNodeId.value = ''
  activateSymbol(selected[0]!, { ...base, symbol: selected[0]! })
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
      language: userContext.value.primaryLanguage,
    },
    { stream: true },
  )
}

function onViewSource(file?: string): void {
  sourcePath.value = file ?? selectedPath.value
  sourceOpen.value = true
}

function onGoToDefinition(file: string, _line: number): void {
  sourcePath.value = file
  sourceOpen.value = true
}

function onExpandNode(node: FlowNode): void {
  if (node.childCount > 3) {
    branchNode.value = node
    branchPromptOpen.value = true
  } else {
    void expandNode(node.id, graphPayload(), node.childCount || 3)
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

function fileName(): string {
  return selectedPath.value.split('/').pop() ?? selectedPath.value
}
</script>

<template>
  <div class="flex h-screen flex-col bg-slate-950">
    <header class="flex items-center gap-3 border-b border-slate-800 bg-slate-900 px-4 py-2.5">
      <button
        type="button"
        class="rounded p-1 text-slate-400 hover:bg-slate-800 hover:text-white"
        aria-label="Toggle files"
        @click="toggleSidebar"
      >
        ☰
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
      <button
        v-if="!sidebarOpen"
        type="button"
        class="flex w-8 shrink-0 flex-col items-center justify-center gap-1 border-r border-slate-800 bg-slate-900/80 text-slate-500 transition hover:bg-slate-800 hover:text-slate-200"
        aria-label="Show file explorer"
        title="Show files"
        @click="toggleSidebar"
      >
        <span class="text-sm">›</span>
        <span class="text-[9px] font-medium uppercase tracking-wide [writing-mode:vertical-rl]">Files</span>
      </button>

      <Sidebar
        v-show="sidebarOpen"
        :workspace-id="userContext.workspaceId"
        :selected-path="selectedPath"
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
          v-if="showSymbolBar"
          :workspace-id="userContext.workspaceId"
          :file-path="selectedPath"
          :active-symbol="symbol"
          :symbol-names="symbolBarNames"
          :has-next-page="symbolBarHasNext"
          :has-prev-page="symbolBarHasPrev"
          :current-page="symbolBarPage"
          :total-pages="symbolBarTotalPages"
          @pick="onPickSymbol"
          @next-page="symbolBarAdvancePage"
          @prev-page="symbolBarPrevPage"
          @go-to-page="symbolBarGoToPage"
        />

        <FileFlowBrief
          v-if="workspacePhase === 'brief' && selectedPath"
          :file-name="fileName()"
          :symbols="symbols"
          :loading="symbolsLoading"
          :error="symbolsError"
          :is-large-file="isLargeFile"
          @confirm="onConfirmFlowInit"
          @decline="onDeclineFlowInit"
        />

        <FlowCanvas
          v-else-if="workspacePhase === 'tracing' || workspacePhase === 'warming'"
          :nodes="nodes"
          :edges="edges"
          :root-id="rootId"
          :loading="loading"
          :enriching="enriching"
          :expanding="expanding"
          :mapping-full-flow="mappingFullFlow"
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
          @view-source="onViewSource()"
          @go-to-definition="onGoToDefinition"
        />

        <div
          v-else
          class="flex flex-1 flex-col items-center justify-center gap-2 p-8 text-center text-slate-500"
        >
          <p class="text-lg font-medium text-slate-300">Select a file to begin</p>
          <p class="max-w-sm text-sm">Choose a file from the explorer to see traceable symbols.</p>
        </div>

        <FlowWarmOverlay
          :open="workspacePhase === 'warming' && warming"
          :file-name="fileName()"
          :done="progress.done"
          :total="progress.total"
          :current-symbol="progress.currentSymbol"
        />
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

    <Modal :open="sourceOpen" title="Source" @close="sourceOpen = false">
      <div class="max-h-[60vh] overflow-auto">
        <CodePanel
          v-if="sourcePath || selectedPath"
          :workspace-id="userContext.workspaceId"
          :file-path="sourcePath || selectedPath"
        />
      </div>
    </Modal>

    <BranchPrompt
      :open="branchPromptOpen"
      :node-label="branchNode?.label ?? ''"
      :child-count="branchNode?.childCount ?? 0"
      @close="branchPromptOpen = false"
      @expand="onConfirmExpand"
    />
  </div>
</template>
