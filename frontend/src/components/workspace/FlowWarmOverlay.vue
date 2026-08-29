<script setup lang="ts">
/** Progress overlay while warming flow maps for multiple symbols. */
defineProps<{
  open: boolean
  fileName: string
  done: number
  total: number
  currentSymbol: string
}>()
</script>

<template>
  <div
    v-if="open"
    class="absolute inset-0 z-20 flex items-center justify-center bg-slate-950/85 backdrop-blur-sm"
  >
    <div class="w-full max-w-md rounded-xl border border-slate-800 bg-slate-900 p-6 shadow-2xl">
      <p class="text-xs font-medium uppercase tracking-wide text-onbober-primary">Initializing</p>
      <h3 class="mt-1 text-lg font-semibold text-white">Flow maps for {{ fileName }}</h3>
      <p class="mt-2 text-sm text-slate-400">
        Preparing the first steps so navigation stays instant.
      </p>

      <div class="mt-4 h-2 overflow-hidden rounded-full bg-slate-800">
        <div
          class="h-full rounded-full bg-onbober-primary transition-all duration-300"
          :style="{ width: total ? `${(done / total) * 100}%` : '0%' }"
        />
      </div>
      <p class="mt-2 text-xs text-slate-500">{{ done }} / {{ total }} symbols</p>
      <p v-if="currentSymbol" class="mt-3 font-mono text-sm text-slate-300">
        Currently: {{ currentSymbol }}
      </p>
    </div>
  </div>
</template>
