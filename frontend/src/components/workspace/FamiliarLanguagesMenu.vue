<script setup lang="ts">
/**
 * Dropdown for editing familiar programming languages in the workspace header.
 */
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  PROGRAMMING_LANGUAGES,
  languageLabelsFromStored,
  storedLanguagesFromLabels,
} from '@/constants/programmingLanguages'
import { familiarLanguageNames } from '@/store/userContext'

const props = defineProps<{
  primaryLanguage: string
}>()

const emit = defineEmits<{
  change: [stored: string]
}>()

const open = ref(false)
const langSearch = ref('')
const draft = ref<Set<string>>(new Set())
const root = ref<HTMLElement | null>(null)

const summary = computed(() => {
  const names = familiarLanguageNames(props.primaryLanguage)
  if (!names.length) return 'Languages'
  if (names.length <= 2) return names.join(', ')
  return `${names.slice(0, 2).join(', ')} +${names.length - 2}`
})

const filteredLanguages = computed(() => {
  const q = langSearch.value.toLowerCase().trim()
  if (!q) return PROGRAMMING_LANGUAGES
  return PROGRAMMING_LANGUAGES.filter((l) => l.label.toLowerCase().includes(q))
})

const canApply = computed(() => draft.value.size > 0)

watch(open, (isOpen) => {
  if (isOpen) {
    draft.value = new Set(languageLabelsFromStored(props.primaryLanguage))
    langSearch.value = ''
  }
})

function toggleOpen(): void {
  open.value = !open.value
}

function toggleLang(label: string): void {
  const next = new Set(draft.value)
  if (next.has(label)) next.delete(label)
  else next.add(label)
  draft.value = next
}

function cancel(): void {
  open.value = false
}

function apply(): void {
  if (!canApply.value) return
  const stored = storedLanguagesFromLabels(draft.value)
  if (stored !== props.primaryLanguage) {
    emit('change', stored)
  }
  open.value = false
}

function onDocumentClick(event: MouseEvent): void {
  if (!open.value || !root.value) return
  if (!root.value.contains(event.target as Node)) {
    open.value = false
  }
}

function onDocumentKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') open.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onDocumentKeydown)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onDocumentKeydown)
})
</script>

<template>
  <div ref="root" class="relative">
    <button
      type="button"
      class="flex max-w-[9.5rem] items-center gap-1.5 rounded-lg border border-slate-700 bg-slate-800/80 px-2.5 py-1 text-xs font-medium text-slate-300 transition hover:border-slate-600 hover:text-white sm:max-w-[11rem]"
      :aria-expanded="open"
      aria-haspopup="listbox"
      title="Languages you know — used for AI labels and analogies"
      @click.stop="toggleOpen"
    >
      <svg class="h-3.5 w-3.5 shrink-0 text-slate-500" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
        <path d="M2 4h12M2 8h8M2 12h10" stroke-linecap="round" />
      </svg>
      <span class="truncate">{{ summary }}</span>
      <svg
        class="h-3 w-3 shrink-0 text-slate-500 transition"
        :class="{ 'rotate-180': open }"
        viewBox="0 0 12 12"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        aria-hidden="true"
      >
        <path d="M3 4.5 6 7.5 9 4.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>

    <div
      v-if="open"
      class="absolute right-0 z-50 mt-1.5 w-72 overflow-hidden rounded-xl border border-slate-700 bg-slate-900 shadow-xl shadow-black/40"
      role="listbox"
      aria-label="Familiar programming languages"
      @click.stop
    >
      <div class="border-b border-slate-800 px-3 py-2.5">
        <p class="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-500">Languages you know</p>
        <input
          v-model="langSearch"
          type="search"
          placeholder="Search languages…"
          class="w-full rounded-lg border border-slate-700 bg-slate-950 px-2.5 py-1.5 text-xs text-slate-100 placeholder:text-slate-500 focus:border-onbober-primary focus:outline-none"
        />
      </div>

      <ul class="max-h-52 overflow-y-auto p-2">
        <li v-for="lang in filteredLanguages" :key="lang.label">
          <button
            type="button"
            class="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-xs transition"
            :class="
              draft.has(lang.label)
                ? 'bg-onbober-primary/10 text-white'
                : 'text-slate-300 hover:bg-slate-800'
            "
            role="option"
            :aria-selected="draft.has(lang.label)"
            @click="toggleLang(lang.label)"
          >
            <span
              class="flex h-4 w-4 shrink-0 items-center justify-center rounded border"
              :class="draft.has(lang.label) ? 'border-onbober-primary bg-onbober-primary text-white' : 'border-slate-600'"
              aria-hidden="true"
            >
              <svg v-if="draft.has(lang.label)" class="h-2.5 w-2.5" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M2 6l3 3 5-5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </span>
            {{ lang.label }}
          </button>
        </li>
      </ul>

      <div class="flex items-center justify-between gap-2 border-t border-slate-800 px-3 py-2">
        <p v-if="!canApply" class="text-[10px] text-amber-400/90">Select at least one</p>
        <span v-else class="text-[10px] text-slate-500">{{ draft.size }} selected</span>
        <div class="ml-auto flex gap-1.5">
          <button
            type="button"
            class="rounded-md px-2 py-1 text-xs text-slate-400 transition hover:bg-slate-800 hover:text-white"
            @click="cancel"
          >
            Cancel
          </button>
          <button
            type="button"
            class="rounded-md bg-onbober-primary px-2.5 py-1 text-xs font-medium text-white transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="!canApply"
            @click="apply"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
