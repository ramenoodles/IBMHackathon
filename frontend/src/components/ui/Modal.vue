<script setup lang="ts">
/**
 * Accessible modal overlay with a default slot for content.
 */
import { onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    /** Controls modal visibility. */
    open: boolean
    /** Optional title shown in the modal header. */
    title?: string
  }>(),
  {
    title: '',
  },
)

const emit = defineEmits<{
  /** Emitted when the user requests to close the modal. */
  close: []
}>()

/**
 * Close the modal when Escape is pressed.
 * @param event - Keyboard event from the document listener.
 */
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && props.open) {
    emit('close')
  }
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      @click.self="emit('close')"
    >
      <div
        class="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl"
        role="dialog"
        aria-modal="true"
      >
        <header v-if="title" class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-slate-100">{{ title }}</h2>
          <button
            type="button"
            class="text-slate-400 hover:text-white"
            aria-label="Close modal"
            @click="emit('close')"
          >
            ✕
          </button>
        </header>
        <slot />
      </div>
    </div>
  </Teleport>
</template>
