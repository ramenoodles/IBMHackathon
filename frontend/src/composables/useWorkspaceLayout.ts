import { onMounted, onUnmounted, ref } from 'vue'

const sidebarOpen = ref(true)
const tracePanelOpen = ref(true)
const detailPanelOpen = ref(true)
const codeDrawerOpen = ref(false)
const isMobile = ref(false)

let layoutListenerCount = 0

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
    toggleSidebar,
    toggleTracePanel,
    toggleDetailPanel,
    toggleCodeDrawer,
  }
}
