import { onMounted, onUnmounted, ref } from 'vue'

/**
 * Reactive layout state for the graph-first workspace.
 */
export function useWorkspaceLayout() {
  const sidebarOpen = ref(true)
  const detailPanelOpen = ref(true)
  const codeDrawerOpen = ref(false)
  const isMobile = ref(false)

  function updateBreakpoint(): void {
    isMobile.value = window.innerWidth < 768
    if (isMobile.value) {
      sidebarOpen.value = false
      detailPanelOpen.value = false
    }
  }

  function toggleSidebar(): void {
    sidebarOpen.value = !sidebarOpen.value
  }

  function toggleDetailPanel(): void {
    detailPanelOpen.value = !detailPanelOpen.value
  }

  function toggleCodeDrawer(): void {
    codeDrawerOpen.value = !codeDrawerOpen.value
  }

  onMounted(() => {
    updateBreakpoint()
    window.addEventListener('resize', updateBreakpoint)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', updateBreakpoint)
  })

  return {
    sidebarOpen,
    detailPanelOpen,
    codeDrawerOpen,
    isMobile,
    toggleSidebar,
    toggleDetailPanel,
    toggleCodeDrawer,
  }
}
