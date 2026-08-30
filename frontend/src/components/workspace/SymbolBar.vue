<script setup lang="ts">
/**
 * Horizontal symbol picker — scrollable chip row with arrow buttons.
 * No pagination. Arrow buttons appear only when there is overflow in that
 * direction. Active chip scrolls into view automatically.
 */
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'

const props = defineProps<{
  workspaceId: string
  filePath: string
  activeSymbol: string
  allSymbols?: string[]
  visible?: boolean
}>()

const emit = defineEmits<{
  pick: [symbol: string]
}>()

const scrollEl = ref<HTMLElement | null>(null)
const canScrollLeft = ref(false)
const canScrollRight = ref(false)

const SCROLL_STEP = 240

function updateScrollState(): void {
  const el = scrollEl.value
  if (!el) return
  canScrollLeft.value = el.scrollLeft > 1
  canScrollRight.value = el.scrollLeft + el.clientWidth < el.scrollWidth - 1
}

function scrollLeft(): void {
  scrollEl.value?.scrollBy({ left: -SCROLL_STEP, behavior: 'smooth' })
  setTimeout(updateScrollState, 320)
}

function scrollRight(): void {
  scrollEl.value?.scrollBy({ left: SCROLL_STEP, behavior: 'smooth' })
  setTimeout(updateScrollState, 320)
}

// Scroll the active chip into view whenever it changes
watch(
  () => props.activeSymbol,
  () => {
    void nextTick(() => {
      const el = scrollEl.value?.querySelector<HTMLElement>('[data-active="true"]')
      el?.scrollIntoView({ inline: 'nearest', block: 'nearest' })
      updateScrollState()
    })
  },
  { immediate: true },
)

// Re-evaluate overflow when symbol list changes (e.g. file switch)
watch(() => props.allSymbols, () => void nextTick(updateScrollState))

let ro: ResizeObserver | null = null

onMounted(() => {
  const el = scrollEl.value
  if (!el) return
  updateScrollState()
  el.addEventListener('scroll', updateScrollState, { passive: true })
  ro = new ResizeObserver(updateScrollState)
  ro.observe(el)
})

onUnmounted(() => {
  scrollEl.value?.removeEventListener('scroll', updateScrollState)
  ro?.disconnect()
})
</script>

<template>
  <div
    class="symbolbar-wrap"
    :class="visible ? 'symbolbar-visible' : 'symbolbar-hidden'"
  >
    <div class="symbolbar-inner">
      <div v-if="filePath" class="flex items-center border-b border-slate-800 bg-slate-900/80">
        <span class="shrink-0 px-3 text-xs text-slate-500">Trace</span>

        <!-- Left arrow — always in flow, invisible when not needed to avoid layout shift -->
        <button
          type="button"
          class="scroll-arrow shrink-0"
          :class="canScrollLeft ? '' : 'invisible'"
          aria-label="Scroll left"
          :tabindex="canScrollLeft ? 0 : -1"
          @click="scrollLeft"
        >
          <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10 3L5 8l5 5"/></svg>
        </button>

        <!-- Chip scroll area -->
        <div class="relative min-w-0 flex-1">
          <div ref="scrollEl" class="chip-track flex gap-1.5 overflow-x-auto px-1 py-2">
            <button
              v-for="sym in allSymbols"
              :key="sym"
              type="button"
              class="shrink-0 rounded-md px-3 py-1 text-xs font-medium transition"
              :class="
                activeSymbol === sym
                  ? 'bg-onbober-primary text-white'
                  : 'bg-slate-800 text-slate-300 hover:bg-slate-700 hover:text-white'
              "
              :data-active="activeSymbol === sym ? 'true' : undefined"
              @click="emit('pick', sym)"
            >
              {{ sym }}
            </button>
            <span v-if="!allSymbols?.length" class="py-1 text-xs text-slate-600">No symbols</span>
          </div>
        </div>

        <!-- Right arrow — always in flow, invisible when not needed -->
        <button
          type="button"
          class="scroll-arrow shrink-0"
          :class="canScrollRight ? '' : 'invisible'"
          aria-label="Scroll right"
          :tabindex="canScrollRight ? 0 : -1"
          @click="scrollRight"
        >
          <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 3l5 5-5 5"/></svg>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.symbolbar-wrap {
  display: grid;
  transition: grid-template-rows 0.22s ease, opacity 0.22s ease;
}
.symbolbar-visible {
  grid-template-rows: 1fr;
  opacity: 1;
}
.symbolbar-hidden {
  grid-template-rows: 0fr;
  opacity: 0;
  pointer-events: none;
}
.symbolbar-inner {
  overflow: hidden;
}

/* Hide scrollbar — arrows + gesture scrolling are the affordance */
.chip-track {
  scrollbar-width: none;
}
.chip-track::-webkit-scrollbar {
  display: none;
}

.scroll-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  margin: 0 2px;
  flex-shrink: 0;
  border-radius: 6px;
  border: 1px solid #334155;
  background: #1e293b;
  color: #94a3b8;
  transition: background 0.15s, color 0.15s, border-color 0.15s;
}
.scroll-arrow svg {
  width: 13px;
  height: 13px;
}
.scroll-arrow:hover {
  background: #334155;
  border-color: #475569;
  color: #f1f5f9;
}
</style>
