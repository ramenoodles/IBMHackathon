<script setup lang="ts">
/**
 * Beaver log-chewing loader for repo clone / archive setup.
 */
import { toRef } from 'vue'
import BeaverLogBar from '@/components/ui/BeaverLogBar.vue'
import { useRotatingPhrase } from '@/composables/useRotatingPhrase'
import { BEAVER_LOADER_DURATION_MS } from '@/constants/beaverLoader'

const props = withDefaults(
  defineProps<{
    active: boolean
    phrases: readonly string[]
    sourceLabel?: string
  }>(),
  { sourceLabel: '' },
)

const { currentPhrase } = useRotatingPhrase(toRef(props, 'active'), toRef(props, 'phrases'))

const REPO_STAGES = [
  { threshold: 0, text: 'MEASURING LOG PROFILE...', badge: 'STAGE 1: TIMBER SCAN' },
  { threshold: 20, text: 'CHISELING OUTER BARK...', badge: 'STAGE 2: GNAWING' },
  { threshold: 50, text: 'RAPID CORE PROCESSING...', badge: 'STAGE 3: FAST CHEWING' },
  { threshold: 75, text: 'TRIMMING WOOD BEAM...', badge: 'STAGE 4: SHAPING' },
  { threshold: 100, text: 'TIMBER COMPLETE! READY!', badge: 'FINISHED' },
] as const
</script>

<template>
  <BeaverLogBar
    mode="timed"
    :active="active"
    :duration-ms="BEAVER_LOADER_DURATION_MS"
    :stages="REPO_STAGES"
    title="Setting up your workspace"
    :subtitle="sourceLabel"
  >
    <Transition name="phrase-fade" mode="out-in">
      <p :key="currentPhrase" class="rotating-phrase">{{ currentPhrase }}</p>
    </Transition>
  </BeaverLogBar>
</template>

<style scoped>
.rotating-phrase {
  text-align: center;
  font-size: 0.875rem;
  font-style: italic;
  color: #94a3b8;
}

.phrase-fade-enter-active,
.phrase-fade-leave-active {
  transition: opacity 0.3s ease;
}

.phrase-fade-enter-from,
.phrase-fade-leave-to {
  opacity: 0;
}
</style>
