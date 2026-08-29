<script setup lang="ts">
/**
 * File flow brief: explains detected symbols and offers to initialize flow maps.
 */
import { computed, ref, watch } from 'vue'
import type { FileSymbol } from '@/utils/flowGraphUtils'
import { DEFAULT_WARM_SELECTION, LARGE_FILE_SYMBOL_THRESHOLD } from '@/utils/flowGraphUtils'

const props = defineProps<{
  fileName: string
  symbols: FileSymbol[]
  loading: boolean
  error: string | null
  isLargeFile: boolean
}>()

const emit = defineEmits<{
  confirm: [symbols: string[]]
  decline: []
}>()

const selected = ref<Set<string>>(new Set())

watch(
  () => props.symbols,
  (list) => {
    const initial = list.slice(0, DEFAULT_WARM_SELECTION).map((s) => s.name)
    selected.value = new Set(initial)
  },
  { immediate: true },
)

const canInit = computed(() => props.symbols.length > 0 && !props.loading)

function toggle(name: string): void {
  const next = new Set(selected.value)
  if (next.has(name)) next.delete(name)
  else next.add(name)
  selected.value = next
}

function onConfirm(): void {
  if (props.isLargeFile) {
    emit('confirm', [...selected.value])
  } else {
    emit(
      'confirm',
      props.symbols.map((s) => s.name),
    )
  }
}

function kindLabel(kind: string): string {
  return kind === 'class' ? 'class' : 'fn'
}
</script>

<template>
  <div class="flex flex-1 flex-col items-center justify-center overflow-auto p-8">
    <div class="w-full max-w-xl rounded-xl border border-slate-800 bg-slate-900/80 p-6 shadow-xl">
      <p class="text-xs font-medium uppercase tracking-wide text-onbober-primary">Flow mapping</p>
      <h2 class="mt-1 text-xl font-semibold text-white">{{ fileName }}</h2>

      <p v-if="loading" class="mt-4 text-sm text-slate-500">Scanning symbols...</p>
      <p v-else-if="error" class="mt-4 text-sm text-red-400">{{ error }}</p>

      <template v-else>
        <p class="mt-4 text-sm leading-relaxed text-slate-400">
          We detected
          <span class="font-medium text-slate-200">{{ symbols.length }}</span>
          traceable
          {{ symbols.length === 1 ? 'symbol' : 'symbols' }} in this file. Flow mapping shows
          execution paths — entry points, branches, calls, and returns — so you can trace how this
          file runs step by step.
        </p>

        <p v-if="isLargeFile" class="mt-3 text-xs text-amber-400/90">
          Large file: pick up to {{ LARGE_FILE_SYMBOL_THRESHOLD }} symbols to warm ({{ selected.size }}
          selected).
        </p>

        <ul
          v-if="symbols.length"
          class="mt-4 max-h-56 space-y-2 overflow-y-auto rounded-lg border border-slate-800 bg-slate-950/50 p-2"
        >
          <li v-for="sym in symbols" :key="sym.name" class="flex items-start gap-2 rounded-md px-2 py-1.5">
            <input
              v-if="isLargeFile"
              type="checkbox"
              class="mt-1 accent-onbober-primary"
              :checked="selected.has(sym.name)"
              @change="toggle(sym.name)"
            />
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="font-mono text-sm text-slate-200">{{ sym.name }}</span>
                <span class="text-[10px] uppercase text-slate-600">{{ kindLabel(sym.kind) }}</span>
                <span class="text-[10px] text-slate-600">L{{ sym.line }}</span>
              </div>
              <p class="truncate font-mono text-[11px] text-slate-500">{{ sym.signature }}</p>
              <p v-if="sym.hint" class="mt-0.5 text-xs text-slate-400">{{ sym.hint }}</p>
            </div>
          </li>
        </ul>

        <p v-else class="mt-4 text-sm text-slate-500">No traceable functions or classes in this file.</p>
      </template>

      <div class="mt-6 flex flex-wrap gap-3">
        <button
          type="button"
          class="rounded-lg bg-onbober-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="!canInit || (isLargeFile && selected.size === 0)"
          @click="onConfirm"
        >
          Initialize flow map
        </button>
        <button
          type="button"
          class="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300 transition hover:bg-slate-800"
          @click="emit('decline')"
        >
          Not now
        </button>
      </div>
    </div>
  </div>
</template>
