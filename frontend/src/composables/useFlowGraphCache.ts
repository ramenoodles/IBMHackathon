import { ref } from 'vue'
import type { SymbolFlowState } from '@/utils/flowGraphUtils'
import { cacheKey, cloneSymbolFlowState } from '@/utils/flowGraphUtils'

/**
 * Per-file symbol flow cache for instant symbol switching after warm-up.
 */
export function useFlowGraphCache() {
  const byKey = ref<Map<string, SymbolFlowState>>(new Map())
  const warmedFiles = ref<Set<string>>(new Set())

  function get(filePath: string, symbol: string): SymbolFlowState | undefined {
    const state = byKey.value.get(cacheKey(filePath, symbol))
    return state ? cloneSymbolFlowState(state) : undefined
  }

  function set(filePath: string, symbol: string, state: SymbolFlowState): void {
    const next = new Map(byKey.value)
    next.set(cacheKey(filePath, symbol), cloneSymbolFlowState(state))
    byKey.value = next
  }

  function has(filePath: string, symbol: string): boolean {
    return byKey.value.has(cacheKey(filePath, symbol))
  }

  function markFileWarmed(filePath: string): void {
    warmedFiles.value = new Set(warmedFiles.value).add(filePath)
  }

  function isFileWarmed(filePath: string): boolean {
    return warmedFiles.value.has(filePath)
  }

  function clearFile(filePath: string): void {
    const next = new Map(byKey.value)
    for (const key of next.keys()) {
      if (key.startsWith(`${filePath}::`)) next.delete(key)
    }
    byKey.value = next
    const files = new Set(warmedFiles.value)
    files.delete(filePath)
    warmedFiles.value = files
  }

  function listSymbolsForFile(filePath: string): string[] {
    const prefix = `${filePath}::`
    const names: string[] = []
    for (const key of byKey.value.keys()) {
      if (key.startsWith(prefix)) {
        names.push(key.slice(prefix.length))
      }
    }
    return names
  }

  function reset(): void {
    byKey.value = new Map()
    warmedFiles.value = new Set()
  }

  return {
    get,
    set,
    has,
    markFileWarmed,
    isFileWarmed,
    clearFile,
    listSymbolsForFile,
    reset,
  }
}
