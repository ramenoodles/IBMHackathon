<script setup lang="ts">
/**
 * Compact segmented control for explanation complexity (experience level).
 */
import { EXPERIENCE_LEVELS } from '@/constants/experienceLevel'
import type { ExperienceLevel } from '@/store/userContext'

const model = defineModel<ExperienceLevel>({ required: true })

const emit = defineEmits<{
  change: [level: ExperienceLevel]
}>()

function select(level: ExperienceLevel): void {
  if (model.value === level) return
  model.value = level
  emit('change', level)
}
</script>

<template>
  <div
    class="flex items-center rounded-lg border border-slate-700 bg-slate-800/80 p-0.5"
    role="group"
    aria-label="Explanation complexity"
    title="Adjust explanation depth for flow labels and step details"
  >
    <button
      v-for="opt in EXPERIENCE_LEVELS"
      :key="opt.value"
      type="button"
      class="rounded-md px-2.5 py-1 text-xs font-medium transition"
      :class="
        model === opt.value
          ? 'bg-onbober-primary/20 text-white'
          : 'text-slate-400 hover:text-slate-200'
      "
      :aria-pressed="model === opt.value"
      :title="opt.description"
      @click="select(opt.value)"
    >
      {{ opt.shortLabel }}
    </button>
  </div>
</template>
