import { computed, ref } from 'vue'
import { api } from '@/api'
import type { FileSymbol } from '@/utils/flowGraphUtils'
import { LARGE_FILE_SYMBOL_THRESHOLD, SYMBOL_PAGE_SIZE } from '@/utils/flowGraphUtils'

/**
 * Fetches structured symbols for a file and paginates them client-side.
 * All symbol names are fetched once; page navigation is pure index arithmetic.
 */
export function useSymbolBrief() {
  const allSymbols = ref<FileSymbol[]>([])
  const currentPage = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const isLargeFile = ref(false)

  const totalPages = computed(() => Math.max(1, Math.ceil(allSymbols.value.length / SYMBOL_PAGE_SIZE)))

  const currentPageSymbols = computed(() => {
    const start = currentPage.value * SYMBOL_PAGE_SIZE
    return allSymbols.value.slice(start, start + SYMBOL_PAGE_SIZE)
  })

  const hasNextPage = computed(() => currentPage.value < totalPages.value - 1)
  const hasPrevPage = computed(() => currentPage.value > 0)

  function advancePage(): void {
    if (hasNextPage.value) currentPage.value++
  }

  function prevPage(): void {
    if (hasPrevPage.value) currentPage.value--
  }

  function goToPage(n: number): void {
    const clamped = Math.max(0, Math.min(n, totalPages.value - 1))
    currentPage.value = clamped
  }

  async function load(workspaceId: string, filePath: string): Promise<void> {
    if (!workspaceId || !filePath) {
      allSymbols.value = []
      currentPage.value = 0
      isLargeFile.value = false
      return
    }
    loading.value = true
    error.value = null
    try {
      const data = await api.symbols<FileSymbol>(workspaceId, filePath)
      allSymbols.value = data.symbols ?? []
      currentPage.value = 0
      isLargeFile.value = allSymbols.value.length > LARGE_FILE_SYMBOL_THRESHOLD
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load symbols'
      allSymbols.value = []
      currentPage.value = 0
      isLargeFile.value = false
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    allSymbols.value = []
    currentPage.value = 0
    loading.value = false
    error.value = null
    isLargeFile.value = false
  }

  return {
    symbols: allSymbols,
    currentPageSymbols,
    currentPage,
    totalPages,
    hasNextPage,
    hasPrevPage,
    loading,
    error,
    isLargeFile,
    load,
    reset,
    advancePage,
    prevPage,
    goToPage,
  }
}
