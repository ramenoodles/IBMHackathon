<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { FileSymbol } from '@/utils/flowGraphUtils'
import { SYMBOL_PAGE_SIZE } from '@/utils/flowGraphUtils'

const props = defineProps<{
  fileName: string
  symbols: FileSymbol[]
  loading: boolean
  error: string | null
  isLargeFile: boolean
}>()

const emit = defineEmits<{
  confirm: [selected: string[]]
  decline: []
}>()

const selected = ref<Set<string>>(new Set())

// Pre-select all (or first page if large file) whenever symbols change
watch(
  () => props.symbols,
  (syms) => {
    const initial = props.isLargeFile ? syms.slice(0, SYMBOL_PAGE_SIZE) : syms
    selected.value = new Set(initial.map((s) => s.name))
  },
  { immediate: true },
)

function toggle(name: string): void {
  if (selected.value.has(name)) {
    selected.value.delete(name)
  } else {
    selected.value.add(name)
  }
  // trigger reactivity
  selected.value = new Set(selected.value)
}

const canConfirm = computed(() => selected.value.size > 0)

function onConfirm(): void {
  emit('confirm', [...selected.value])
}

const kindColour: Record<string, string> = {
  function: 'text-sky-400 bg-sky-950',
  method: 'text-violet-400 bg-violet-950',
  class: 'text-amber-400 bg-amber-950',
  constructor: 'text-emerald-400 bg-emerald-950',
}

function kindClass(kind: string): string {
  return kindColour[kind.toLowerCase()] ?? 'text-slate-400 bg-slate-800'
}
</script>

<template>
  <div class="flex flex-1 flex-col items-center justify-center p-8">
    <div class="w-full max-w-lg rounded-xl border border-slate-800 bg-slate-900/80 p-6 shadow-xl">
      <p class="text-xs font-medium uppercase tracking-wide text-onbober-primary">Flow mapping</p>
      <h2 class="mt-1 truncate text-xl font-semibold text-white">{{ fileName }}</h2>

      <!-- Loading -->
      <p v-if="loading" class="mt-4 text-sm text-slate-500">Scanning symbols…</p>

      <!-- Error -->
      <p v-else-if="error" class="mt-4 text-sm text-red-400">{{ error }}</p>

      <!-- No symbols -->
      <template v-else-if="symbols.length === 0">
        <p class="mt-4 text-sm text-slate-500">No traceable symbols found in this file.</p>
        <div class="mt-6">
          <button
            type="button"
            class="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300 transition hover:bg-slate-800"
            @click="emit('decline')"
          >
            Close
          </button>
        </div>
      </template>

      <!-- Symbol list -->
      <template v-else>
        <p class="mt-3 text-sm text-slate-400">
          Select the symbols you'd like to map, then start tracing.
          <span v-if="isLargeFile" class="ml-1 text-amber-400"
            >Large file — first {{ symbols.length }} symbols shown.</span
          >
        </p>

        <ul class="mt-4 max-h-64 space-y-1 overflow-y-auto pr-1">
          <li
            v-for="sym in symbols"
            :key="sym.name"
            class="flex cursor-pointer items-start gap-3 rounded-lg px-3 py-2 transition hover:bg-slate-800"
            @click="toggle(sym.name)"
          >
            <input
              type="checkbox"
              class="mt-0.5 shrink-0 accent-onbober-primary"
              :checked="selected.has(sym.name)"
              @click.stop="toggle(sym.name)"
            />
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="truncate text-sm font-medium text-white">{{ sym.name }}</span>
                <span
                  :class="['shrink-0 rounded px-1.5 py-0.5 text-[10px] font-medium uppercase', kindClass(sym.kind)]"
                >{{ sym.kind }}</span>
              </div>
              <p v-if="sym.hint" class="mt-0.5 truncate text-xs text-slate-500">{{ sym.hint }}</p>
              <p v-else class="mt-0.5 truncate text-xs text-slate-600">{{ sym.signature }}</p>
            </div>
          </li>
        </ul>

        <div class="mt-6 flex flex-wrap gap-3">
          <button
            type="button"
            class="rounded-lg bg-onbober-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="!canConfirm"
            @click="onConfirm"
          >
            Start tracing
          </button>
          <button
            type="button"
            class="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300 transition hover:bg-slate-800"
            @click="emit('decline')"
          >
            Skip
          </button>
        </div>
      </template>
    </div>
  </div>
</template>
