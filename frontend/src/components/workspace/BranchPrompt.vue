<script setup lang="ts">
/**
 * Modal prompting the user to expand collapsed branch nodes.
 */
import Button from '@/components/ui/Button.vue'
import { ref } from 'vue'

const props = defineProps<{
  open: boolean
  nodeLabel: string
  childCount: number
}>()

const emit = defineEmits<{
  close: []
  expand: [limit: number]
}>()

const selectedLimit = ref(3)

/** Confirm expansion with the selected branch limit. */
function confirm(): void {
  emit('expand', selectedLimit.value)
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      @click.self="emit('close')"
    >
      <div class="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6">
        <h2 class="mb-2 text-lg font-semibold text-white">Expand branches?</h2>
        <p class="mb-4 text-sm text-slate-400">
          <span class="text-slate-200">{{ nodeLabel }}</span> has
          {{ childCount }} possible paths. How many should we load?
        </p>
        <div class="mb-6 flex gap-2">
          <button
            v-for="n in [3, 5, childCount > 6 ? 6 : childCount]"
            :key="n"
            type="button"
            class="flex-1 rounded-lg border px-3 py-2 text-sm transition"
            :class="
              selectedLimit === n
                ? 'border-onbober-primary bg-onbober-primary/10 text-white'
                : 'border-slate-700 text-slate-400 hover:border-slate-500'
            "
            @click="selectedLimit = n"
          >
            {{ n }}
          </button>
        </div>
        <div class="flex justify-end gap-2">
          <Button variant="ghost" @click="emit('close')">Cancel</Button>
          <Button variant="primary" @click="confirm">Expand</Button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
