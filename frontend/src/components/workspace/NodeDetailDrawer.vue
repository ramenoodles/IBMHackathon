<script setup lang="ts">
/**
 * Side drawer showing detailed explanation for a selected flow node.
 */
import type { NodeDetail } from '@/types/flowGraph'

defineProps<{
  detail: NodeDetail | null
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  jumpToCode: []
  toggle: []
}>()
</script>

<template>
  <aside class="flex h-full w-80 flex-col border-l border-slate-800 bg-slate-900">
    <div class="flex items-center justify-between border-b border-slate-800 px-3 py-2">
      <span class="text-xs font-semibold uppercase tracking-wide text-slate-400">Node Detail</span>
      <button type="button" class="text-slate-400 hover:text-white md:hidden" @click="emit('toggle')">✕</button>
    </div>

    <div class="flex-1 overflow-y-auto p-4 text-sm">
      <p v-if="!detail && !loading" class="text-slate-500">Select a node to see details.</p>
      <p v-if="loading" class="text-onbober-primary">Loading...</p>
      <p v-if="error" class="text-red-400">{{ error }}</p>

      <template v-if="detail">
        <div v-if="detail.mock" class="mb-3 rounded border border-amber-800/50 bg-amber-900/20 px-2 py-1 text-xs text-amber-300">
          Watsonx explanation unavailable
        </div>
        <h2 class="text-lg font-bold text-white">{{ detail.title }}</h2>
        <span
          class="mt-1 inline-block rounded px-2 py-0.5 text-xs font-bold uppercase"
          :class="
            detail.confidence === 'verified'
              ? 'bg-green-900/40 text-green-400'
              : 'bg-amber-900/40 text-amber-400'
          "
        >
          {{ detail.confidence }}
        </span>
        <p class="mt-3 text-slate-300">{{ detail.summary }}</p>
        <p class="mt-3 leading-relaxed text-slate-400">{{ detail.explanation }}</p>

        <div v-if="detail.relatedSymbols?.length" class="mt-4">
          <p class="mb-2 text-xs uppercase text-slate-500">Related</p>
          <div class="flex flex-wrap gap-1">
            <span
              v-for="sym in detail.relatedSymbols"
              :key="sym"
              class="rounded-full border border-slate-700 px-2 py-0.5 text-xs text-slate-300"
            >
              {{ sym }}
            </span>
          </div>
        </div>

        <button
          v-if="detail.file"
          type="button"
          class="mt-4 w-full rounded border border-slate-600 px-3 py-2 text-xs text-slate-300 hover:border-onbober-primary hover:text-white"
          @click="emit('jumpToCode')"
        >
          Jump to code{{ detail.line ? ` (line ${detail.line})` : '' }}
        </button>
      </template>
    </div>
  </aside>
</template>
