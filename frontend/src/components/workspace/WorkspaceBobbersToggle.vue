<script setup lang="ts">
/**
 * Toggle decorative workspace bobbers on or off.
 */
const model = defineModel<boolean>({ required: true })

function toggle(): void {
  model.value = !model.value
}
</script>

<template>
  <button
    type="button"
    role="switch"
    class="flex items-center gap-2 rounded-lg border px-2.5 py-1 text-xs font-medium transition"
    :class="
      model
        ? 'border-onbober-primary/40 bg-onbober-primary/10 text-white'
        : 'border-slate-700 bg-slate-800/80 text-slate-400 hover:text-slate-200'
    "
    :aria-checked="model"
    :title="model ? 'Hide little bobers in the workspace' : 'Show little bobers in the workspace'"
    @click="toggle"
  >
    <span
      class="relative h-4 w-7 shrink-0 rounded-full transition"
      :class="model ? 'bg-onbober-primary/40' : 'bg-slate-700'"
      aria-hidden="true"
    >
      <span
        class="absolute top-0.5 h-3 w-3 rounded-full bg-white shadow transition"
        :class="model ? 'left-3.5' : 'left-0.5'"
      />
    </span>
    <span class="hidden sm:inline">Little bobers</span>
  </button>
</template>
