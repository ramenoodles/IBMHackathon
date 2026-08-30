<script setup lang="ts">
/**
 * Minimal yes/no prompt shown when a file is first selected.
 * Asks the user whether to load the first page of symbols.
 */
const props = defineProps<{
  fileName: string
  loading: boolean
  error: string | null
  symbolCount: number
}>()

const emit = defineEmits<{
  confirm: []
  decline: []
}>()
</script>

<template>
  <div class="flex flex-1 flex-col items-center justify-center p-8">
    <div class="w-full max-w-md rounded-xl border border-slate-800 bg-slate-900/80 p-6 shadow-xl">
      <p class="text-xs font-medium uppercase tracking-wide text-onbober-primary">Flow mapping</p>
      <h2 class="mt-1 text-xl font-semibold text-white">{{ fileName }}</h2>

      <p v-if="loading" class="mt-4 text-sm text-slate-500">Scanning symbols…</p>
      <p v-else-if="error" class="mt-4 text-sm text-red-400">{{ error }}</p>
      <p v-else-if="symbolCount === 0" class="mt-4 text-sm text-slate-500">No traceable symbols found in this file.</p>
      <p v-else class="mt-4 text-sm leading-relaxed text-slate-400">
        Would you like to load the first 8 symbols in this file?
      </p>

      <div class="mt-6 flex flex-wrap gap-3">
        <button
          type="button"
          class="rounded-lg bg-onbober-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="loading || symbolCount === 0"
          @click="emit('confirm')"
        >
          Yes
        </button>
        <button
          type="button"
          class="rounded-lg border border-slate-700 px-4 py-2 text-sm text-slate-300 transition hover:bg-slate-800"
          @click="emit('decline')"
        >
          No
        </button>
      </div>
    </div>
  </div>
</template>
