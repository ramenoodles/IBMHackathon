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
    <Transition name="modal">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
        @click.self="emit('close')"
      >
        <div
          class="modal-dialog w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl"
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
              <svg class="h-4 w-4" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" aria-hidden="true">
                <path d="M3 3l10 10M13 3L3 13" />
              </svg>
            </button>
          </header>
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
/* Backdrop fades, dialog scales up */
.modal-enter-active {
  transition: opacity 0.2s ease;
}
.modal-leave-active {
  transition: opacity 0.15s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-dialog {
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.2s ease;
}
.modal-leave-active .modal-dialog {
  transition: transform 0.15s ease, opacity 0.15s ease;
}
.modal-enter-from .modal-dialog {
  transform: scale(0.95) translateY(8px);
  opacity: 0;
}
.modal-leave-to .modal-dialog {
  transform: scale(0.97) translateY(4px);
  opacity: 0;
}
</style>
