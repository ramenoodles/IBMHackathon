<script setup lang="ts">
/**
 * Toggle whether LLM labels and explanations include brief analogies to familiar languages.
 */
import { computed } from 'vue'
import { familiarLanguageNames } from '@/store/userContext'

const props = defineProps<{
  primaryLanguage: string
}>()

const model = defineModel<boolean>({ required: true })

const emit = defineEmits<{
  change: [enabled: boolean]
}>()

const familiar = computed(() => familiarLanguageNames(props.primaryLanguage))
const disabled = computed(() => familiar.value.length === 0)

const title = computed(() => {
  if (disabled.value) {
    return 'Select languages you know to enable cross-language analogies'
  }
  const langs = familiar.value.join(', ')
  if (model.value) {
    return `Analogies on — summaries may relate this code to ${langs}`
  }
  return `Analogies off — enable to relate steps to ${langs}`
})

function toggle(): void {
  if (disabled.value) return
  const next = !model.value
  model.value = next
  emit('change', next)
}
</script>

<template>
  <button
    type="button"
    role="switch"
    class="flex items-center gap-2 rounded-lg border px-2.5 py-1 text-xs font-medium transition disabled:cursor-not-allowed disabled:opacity-40"
    :class="
      model && !disabled
        ? 'border-onbober-primary/40 bg-onbober-primary/10 text-white'
        : 'border-slate-700 bg-slate-800/80 text-slate-400 hover:text-slate-200'
    "
    :aria-checked="model"
    :disabled="disabled"
    :title="title"
    @click="toggle"
  >
    <span
      class="relative h-4 w-7 shrink-0 rounded-full transition"
      :class="model && !disabled ? 'bg-onbober-primary/40' : 'bg-slate-700'"
      aria-hidden="true"
    >
      <span
        class="absolute top-0.5 h-3 w-3 rounded-full bg-white shadow transition"
        :class="model && !disabled ? 'left-3.5' : 'left-0.5'"
      />
    </span>
    <span class="hidden sm:inline">Lang analogies</span>
  </button>
</template>
