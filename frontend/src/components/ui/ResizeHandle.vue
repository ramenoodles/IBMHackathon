<script setup lang="ts">
defineProps<{
  label?: string
}>()

const emit = defineEmits<{
  pointerdown: [event: PointerEvent]
}>()
</script>

<template>
  <div
    class="group relative z-10 w-1 shrink-0 cursor-col-resize bg-slate-800/40 hover:bg-onbober-primary/40"
    :title="label ?? 'Drag to resize'"
    role="separator"
    aria-orientation="vertical"
    @pointerdown="emit('pointerdown', $event)"
  >
    <div class="absolute inset-y-0 -left-1 -right-1" />
  </div>
</template>
