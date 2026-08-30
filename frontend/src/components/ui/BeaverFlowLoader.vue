<script setup lang="ts">
/**
 * Minimal beaver log bar for flow graph loading — SVG animation only.
 */
import { computed } from 'vue'
import BeaverLogBar from '@/components/ui/BeaverLogBar.vue'

const props = withDefaults(
  defineProps<{
    mode: 'indeterminate' | 'progress'
    active: boolean
    progress?: number
    compact?: boolean
  }>(),
  {
    progress: 0,
    compact: false,
  },
)

const animationMode = computed(() => (props.mode === 'progress' ? 'progress' : 'indeterminate'))
</script>

<template>
  <BeaverLogBar
    bar-only
    :mode="animationMode"
    :active="active"
    :external-progress="progress"
    :compact="compact"
  />
</template>
