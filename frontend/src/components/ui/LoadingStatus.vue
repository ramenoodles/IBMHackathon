<script setup lang="ts">
/**
 * Reusable loading indicator: blink dots, rotating phrases, shimmer skeleton, optional progress bar.
 */
import { toRef } from 'vue'
import { useRotatingPhrase } from '@/composables/useRotatingPhrase'

const props = withDefaults(
  defineProps<{
    phrases: readonly string[]
    active: boolean
    label?: string
    showShimmer?: boolean
    showProgressBar?: boolean
    phraseClass?: string
  }>(),
  {
    showShimmer: true,
    showProgressBar: false,
    phraseClass: 'text-[11px] italic text-slate-500',
  },
)

const { phraseIndex, currentPhrase } = useRotatingPhrase(
  toRef(props, 'active'),
  toRef(props, 'phrases'),
)
</script>

<template>
  <div v-if="active" class="space-y-2">
    <p v-if="label" class="text-xs font-medium uppercase tracking-wide text-onbober-primary">
      {{ label }}
    </p>

    <div class="flex items-center gap-2">
      <span class="h-1.5 w-1.5 animate-[blink_1s_ease-in-out_0s_infinite] rounded-full bg-onbober-primary" />
      <span class="h-1.5 w-1.5 animate-[blink_1s_ease-in-out_0.2s_infinite] rounded-full bg-onbober-primary" />
      <span class="h-1.5 w-1.5 animate-[blink_1s_ease-in-out_0.4s_infinite] rounded-full bg-onbober-primary" />
      <Transition name="phrase-fade" mode="out-in">
        <span :key="phraseIndex" :class="phraseClass">{{ currentPhrase }}</span>
      </Transition>
    </div>

    <template v-if="showShimmer">
      <div class="h-2.5 w-full animate-[shimmer_1.5s_ease-in-out_infinite] rounded bg-gradient-to-r from-slate-800 via-slate-700 to-slate-800 bg-[length:200%_100%]" />
      <div class="h-2.5 w-4/5 animate-[shimmer_1.5s_ease-in-out_0.15s_infinite] rounded bg-gradient-to-r from-slate-800 via-slate-700 to-slate-800 bg-[length:200%_100%]" />
      <div class="h-2.5 w-full animate-[shimmer_1.5s_ease-in-out_0.3s_infinite] rounded bg-gradient-to-r from-slate-800 via-slate-700 to-slate-800 bg-[length:200%_100%]" />
      <div class="h-2.5 w-3/5 animate-[shimmer_1.5s_ease-in-out_0.45s_infinite] rounded bg-gradient-to-r from-slate-800 via-slate-700 to-slate-800 bg-[length:200%_100%]" />
    </template>

    <div v-if="showProgressBar" class="mt-3 h-1.5 overflow-hidden rounded-full bg-slate-800">
      <div class="progress-indeterminate h-full rounded-full bg-onbober-primary" />
    </div>
  </div>
</template>

<style scoped>
.phrase-fade-enter-active,
.phrase-fade-leave-active {
  transition: opacity 0.35s ease;
}
.phrase-fade-enter-from,
.phrase-fade-leave-to {
  opacity: 0;
}

.progress-indeterminate {
  width: 40%;
  animation: progress-indeterminate 1.5s ease-in-out infinite;
}

@keyframes progress-indeterminate {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(350%);
  }
}
</style>
