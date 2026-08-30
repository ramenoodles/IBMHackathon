<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'

withDefaults(
  defineProps<{
    variant?: 'workspace' | 'splash'
    /** Where the popover opens relative to the button. */
    placement?: 'above' | 'below'
  }>(),
  { variant: 'workspace', placement: 'above' },
)

const open = ref(false)
const root = ref<HTMLElement | null>(null)

interface HelpRow {
  label: string
  hint: string
}

const workspaceTips: HelpRow[] = [
  { label: 'Trace chips', hint: 'Another function in this file' },
  { label: 'A step', hint: 'In the list or a box in the graph' },
  { label: 'Panel arrows', hint: 'Hide or show Steps / Details' },
  { label: '− / Center / +', hint: 'Zoom the graph' },
  { label: 'Mouse wheel', hint: 'Scroll to zoom in and out' },
  { label: 'Show full flow', hint: 'Open hidden branches' },
  { label: 'Explain this step', hint: 'More detail on the selected step' },
  { label: 'Expand / View flow', hint: 'On folded call steps' },
  { label: 'Go to / View source', hint: 'Jump to code' },
]

const splashTips: HelpRow[] = [
  { label: 'Get started', hint: 'Load a repo and open the workspace' },
]

function toggle(): void {
  open.value = !open.value
}

function close(): void {
  open.value = false
}

function onDocumentClick(event: MouseEvent): void {
  if (!open.value || !root.value) return
  if (!root.value.contains(event.target as Node)) {
    close()
  }
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') close()
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div ref="root" class="relative shrink-0">
    <button
      type="button"
      class="rounded-md border-2 border-onbober-primary bg-slate-900 px-3 py-1.5 text-sm font-semibold text-onbober-primary shadow-[0_0_10px_rgba(255,51,102,0.35)] ring-2 ring-fuchsia-500/30 transition hover:border-fuchsia-400 hover:text-fuchsia-300 hover:shadow-[0_0_14px_rgba(255,51,102,0.5)]"
      :aria-expanded="open"
      aria-haspopup="dialog"
      aria-label="Help — what you can click"
      title="Help"
      @click.stop="toggle"
    >
      Help
    </button>

    <div
      v-if="open"
      role="dialog"
      aria-label="What you can click"
      class="absolute z-50 w-[min(22rem,calc(100vw-1.5rem))] rounded-lg border border-slate-700 bg-slate-900 p-3 text-left shadow-xl shadow-black/40"
      :class="
        placement === 'below'
          ? 'top-full left-0 mt-2'
          : 'bottom-full right-0 mb-2'
      "
      @click.stop
    >
      <p class="text-xs font-semibold text-slate-200">What you can click</p>

      <div
        class="mt-2 overflow-hidden rounded-md border border-slate-800/80 text-[11px] leading-snug"
        role="table"
        aria-label="Clickable controls"
      >
        <div
          v-for="(row, index) in variant === 'splash' ? splashTips : workspaceTips"
          :key="row.label"
          class="grid grid-cols-[8.25rem_1fr] gap-x-3 px-2.5 py-1.5"
          :class="index > 0 ? 'border-t border-slate-800/60' : ''"
          role="row"
        >
          <span class="font-medium text-slate-300" role="cell">{{ row.label }}</span>
          <span class="text-slate-500" role="cell">{{ row.hint }}</span>
        </div>
      </div>

      <button
        type="button"
        class="mt-3 w-full rounded border border-slate-700 px-2 py-1 text-[10px] font-medium text-slate-400 transition hover:border-slate-500 hover:text-white"
        @click="close"
      >
        Got it
      </button>
    </div>
  </div>
</template>
