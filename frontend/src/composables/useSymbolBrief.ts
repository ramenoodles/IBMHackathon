import { computed, ref } from 'vue'
import type { FileSymbol } from '@/utils/flowGraphUtils'
import { SYMBOL_PAGE_SIZE } from '@/utils/flowGraphUtils'

/**
 * Fetches structured symbols for a file and paginates them client-side.
 * All symbol names are fetched once; page navigation is pure index arithmetic.
 */
export function useSymbolBrief() {
  const allSymbols = ref<FileSymbol[]>([])
  const currentPage = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

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

  async function load(workspacePath: string, filePath: string): Promise<void> {
    if (!workspacePath || !filePath) {
      allSymbols.value = []
      currentPage.value = 0
      return
    }
    loading.value = true
    error.value = null
    try {
      const params = new URLSearchParams({ workspace: workspacePath, path: filePath })
      const res = await fetch(`/api/file/symbols?${params}`)
      if (!res.ok) throw new Error(`Symbol scan failed (${res.status})`)
      const data = (await res.json()) as { symbols: FileSymbol[]; count: number }
      allSymbols.value = data.symbols ?? []
      currentPage.value = 0
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load symbols'
      allSymbols.value = []
      currentPage.value = 0
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    allSymbols.value = []
    currentPage.value = 0
    loading.value = false
    error.value = null
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
    load,
    reset,
    advancePage,
    prevPage,
    goToPage,
  }
}
