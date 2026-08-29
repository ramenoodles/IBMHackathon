import { onMounted, onUnmounted, ref } from 'vue'

const sidebarOpen = ref(true)
const tracePanelOpen = ref(true)
const detailPanelOpen = ref(true)
const codeDrawerOpen = ref(false)
const isMobile = ref(false)

const explorerWidth = ref(176)
const traceWidth = ref(288)
const detailWidth = ref(288)

const EXPLORER_MIN = 140
const EXPLORER_MAX = 400
const TRACE_MIN = 200
const TRACE_MAX = 480
const DETAIL_MIN = 240
const DETAIL_MAX = 520

const STORAGE_KEY = 'onbober.layout'

let layoutListenerCount = 0

function loadPersistedWidths(): void {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as {
      explorerWidth?: number
      traceWidth?: number
      detailWidth?: number
    }
    if (data.explorerWidth) explorerWidth.value = clamp(data.explorerWidth, EXPLORER_MIN, EXPLORER_MAX)
    if (data.traceWidth) traceWidth.value = clamp(data.traceWidth, TRACE_MIN, TRACE_MAX)
    if (data.detailWidth) detailWidth.value = clamp(data.detailWidth, DETAIL_MIN, DETAIL_MAX)
  } catch {
    /* ignore */
  }
}

function persistWidths(): void {
  sessionStorage.setItem(
    STORAGE_KEY,
    JSON.stringify({
      explorerWidth: explorerWidth.value,
      traceWidth: traceWidth.value,
      detailWidth: detailWidth.value,
    }),
  )
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n))
}

function updateBreakpoint(): void {
  isMobile.value = window.innerWidth < 768
  if (isMobile.value) {
    sidebarOpen.value = false
    tracePanelOpen.value = false
    detailPanelOpen.value = false
  }
}

/**
 * Reactive layout state for the graph-first workspace.
 */
export function useWorkspaceLayout() {
  function toggleSidebar(): void {
    sidebarOpen.value = !sidebarOpen.value
  }

  function toggleTracePanel(): void {
    tracePanelOpen.value = !tracePanelOpen.value
  }

  function toggleDetailPanel(): void {
    detailPanelOpen.value = !detailPanelOpen.value
  }

  function toggleCodeDrawer(): void {
    codeDrawerOpen.value = !codeDrawerOpen.value
  }

  onMounted(() => {
    if (layoutListenerCount === 0) {
      loadPersistedWidths()
      updateBreakpoint()
      window.addEventListener('resize', updateBreakpoint)
    }
    layoutListenerCount += 1
  })

  onUnmounted(() => {
    layoutListenerCount -= 1
    if (layoutListenerCount === 0) {
      window.removeEventListener('resize', updateBreakpoint)
    }
  })

  return {
    sidebarOpen,
    tracePanelOpen,
    detailPanelOpen,
    codeDrawerOpen,
    isMobile,
    explorerWidth,
    traceWidth,
    detailWidth,
    EXPLORER_MIN,
    EXPLORER_MAX,
    TRACE_MIN,
    TRACE_MAX,
    DETAIL_MIN,
    DETAIL_MAX,
    persistWidths,
    toggleSidebar,
    toggleTracePanel,
    toggleDetailPanel,
    toggleCodeDrawer,
  }
}
