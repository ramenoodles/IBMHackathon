import { ref } from 'vue'
import type { FileSymbol } from '@/utils/flowGraphUtils'
import { LARGE_FILE_SYMBOL_THRESHOLD } from '@/utils/flowGraphUtils'

/**
 * Fetches structured symbols for the file flow brief screen.
 */
export function useSymbolBrief() {
  const symbols = ref<FileSymbol[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isLargeFile = ref(false)

  async function load(workspacePath: string, filePath: string): Promise<void> {
    if (!workspacePath || !filePath) {
      symbols.value = []
      isLargeFile.value = false
      return
    }
    loading.value = true
    error.value = null
    try {
      const params = new URLSearchParams({ workspace: workspacePath, path: filePath })
      const res = await fetch(`/api/file/symbols?${params}`)
      if (!res.ok) throw new Error(`Symbol scan failed (${res.status})`)
      const data = (await res.json()) as { symbols: FileSymbol[]; count: number }
      symbols.value = data.symbols ?? []
      isLargeFile.value = symbols.value.length > LARGE_FILE_SYMBOL_THRESHOLD
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load symbols'
      symbols.value = []
      isLargeFile.value = false
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    symbols.value = []
    loading.value = false
    error.value = null
    isLargeFile.value = false
  }

  return { symbols, loading, error, isLargeFile, load, reset }
}
