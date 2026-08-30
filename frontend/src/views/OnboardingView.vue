<script setup lang="ts">
/**
 * Multi-step onboarding flow collecting developer context for tailored AI analysis.
 */
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import LoadingStatus from '@/components/ui/LoadingStatus.vue'
import logo from '@/assets/logo.png'
import {
  type ExperienceLevel,
  updateUserContext,
  userContext,
} from '@/store/userContext'
import {
  type WorkspaceSource,
  useWorkspaceSetup,
} from '@/composables/useWorkspaceSetup'
import {
  GITHUB_SETUP_PHRASES,
  LOCAL_SETUP_PHRASES,
  WORKSPACE_SETUP_PHRASES,
  ZIP_SETUP_PHRASES,
} from '@/constants/workspaceSetupPhrases'

const router = useRouter()
const { loading, error, setupLocal, setupGitHub, setupZip } = useWorkspaceSetup()

const step = ref(1)
const totalSteps = 3

/**
 * All language options. `value` must match a backend language profile name or
 * alias (auto, c, cpp, csharp, go, java, javascript, python, rust, typescript).
 * Use "auto" for anything not in the list so the backend searches all files.
 */
const ALL_LANGUAGES: { label: string; value: string }[] = [
  { label: 'Python',      value: 'python' },
  { label: 'JavaScript',  value: 'javascript' },
  { label: 'TypeScript',  value: 'typescript' },
  { label: 'Go',          value: 'go' },
  { label: 'Rust',        value: 'rust' },
  { label: 'Java',        value: 'java' },
  { label: 'C',           value: 'c' },
  { label: 'C++',         value: 'cpp' },
  { label: 'C#',          value: 'csharp' },
  { label: 'Ruby',        value: 'auto' },
  { label: 'PHP',         value: 'auto' },
  { label: 'Swift',       value: 'auto' },
  { label: 'Kotlin',      value: 'auto' },
  { label: 'Scala',       value: 'auto' },
  { label: 'Haskell',     value: 'auto' },
  { label: 'Lua',         value: 'auto' },
  { label: 'Elixir',      value: 'auto' },
  { label: 'Clojure',     value: 'auto' },
  { label: 'Dart',        value: 'auto' },
  { label: 'R',           value: 'auto' },
  { label: 'MATLAB',      value: 'auto' },
  { label: 'Shell/Bash',  value: 'auto' },
  { label: 'Other',       value: 'auto' },
]

const langSearch = ref('')
const filteredLanguages = computed(() => {
  const q = langSearch.value.toLowerCase().trim()
  if (!q) return ALL_LANGUAGES
  return ALL_LANGUAGES.filter((l) => l.label.toLowerCase().includes(q))
})

const levels: { value: ExperienceLevel; label: string }[] = [
  { value: 'junior', label: 'Junior SWE' },
  { value: 'mid', label: 'Mid-level SWE' },
  { value: 'senior', label: 'Senior SWE' },
]

// Restore from stored comma-separated labels, matching both stored value and label.
const storedLang = userContext.value.primaryLanguage || ''
const selectedLangLabels = ref<Set<string>>(
  new Set(
    storedLang
      .split(',')
      .map((s) => s.trim())
      .map((s) => ALL_LANGUAGES.find((l) => l.value === s || l.label === s)?.label ?? '')
      .filter(Boolean),
  ),
)

function toggleLang(label: string): void {
  if (selectedLangLabels.value.has(label)) {
    selectedLangLabels.value.delete(label)
  } else {
    selectedLangLabels.value.add(label)
  }
  // Trigger reactivity — Set mutations are not tracked automatically.
  selectedLangLabels.value = new Set(selectedLangLabels.value)
}

// The value stored in userContext: comma-separated backend-safe values, deduplicated.
// First non-auto value wins for the explain endpoint; auto is the fallback.
const selectedLanguagesValue = computed(() => {
  const values = [...selectedLangLabels.value]
    .map((label) => ALL_LANGUAGES.find((l) => l.label === label)?.value ?? 'auto')
  const unique = [...new Set(values)]
  return unique.join(',')
})
const selectedLevel = ref<ExperienceLevel>(userContext.value.experienceLevel || 'junior')

const sourceTab = ref<WorkspaceSource>('local')
const localPath = ref('')
const githubUrl = ref('')
const zipFile = ref<File | null>(null)

const sourceOptions: { value: WorkspaceSource; label: string }[] = [
  { value: 'local', label: 'Local path' },
  { value: 'github', label: 'GitHub repo' },
  { value: 'zip', label: 'Upload ZIP' },
]

const setupPhrases = computed(() => {
  if (sourceTab.value === 'zip') return ZIP_SETUP_PHRASES
  if (sourceTab.value === 'github') return GITHUB_SETUP_PHRASES
  if (sourceTab.value === 'local') return LOCAL_SETUP_PHRASES
  return WORKSPACE_SETUP_PHRASES
})

const setupSourceLabel = computed(() => {
  if (sourceTab.value === 'zip' && zipFile.value) return zipFile.value.name
  if (sourceTab.value === 'github' && githubUrl.value.trim()) return githubUrl.value.trim()
  if (sourceTab.value === 'local' && localPath.value.trim()) return localPath.value.trim()
  return ''
})

const canProceed = computed(() => {
  if (step.value === 1) return selectedLangLabels.value.size > 0
  if (step.value === 2) return Boolean(selectedLevel.value)
  if (step.value === 3) {
    if (sourceTab.value === 'local') return localPath.value.trim().length > 0
    if (sourceTab.value === 'github') return githubUrl.value.trim().length > 0
    if (sourceTab.value === 'zip') return zipFile.value !== null
  }
  return false
})

/** Handle zip file selection from the file input. */
function onZipChange(event: Event): void {
  const input = event.target as HTMLInputElement
  zipFile.value = input.files?.[0] ?? null
}

/** Advance to the next onboarding step or finish and enter the workspace. */
async function nextStep(): Promise<void> {
  if (step.value === 1) {
    updateUserContext({ primaryLanguage: selectedLanguagesValue.value || 'auto' })
  } else if (step.value === 2) {
    updateUserContext({ experienceLevel: selectedLevel.value })
  } else if (step.value === 3) {
    try {
      let workspace: { id: string; name: string } | undefined
      if (sourceTab.value === 'local') {
        workspace = await setupLocal(localPath.value.trim())
      } else if (sourceTab.value === 'github') {
        workspace = await setupGitHub(githubUrl.value.trim())
      } else if (sourceTab.value === 'zip' && zipFile.value) {
        workspace = await setupZip(zipFile.value)
      }
      if (!workspace) return
      updateUserContext({ workspaceId: workspace.id, workspaceName: workspace.name })
      router.push('/workspace')
      return
    } catch {
      return
    }
  }
  step.value += 1
}

/** Return to the previous onboarding step. */
function prevStep(): void {
  if (step.value > 1) step.value -= 1
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-950 px-6">
    <div class="w-full max-w-lg rounded-xl border border-slate-800 bg-slate-900 p-8">
      <div class="mb-6 flex justify-center">
        <img
          :src="logo"
          alt="OnBober mascot"
          class="h-24 w-auto transition-opacity"
          :class="loading ? 'animate-pulse' : ''"
        />
      </div>
      <p class="mb-2 text-sm text-onbober-primary">Step {{ step }} of {{ totalSteps }}</p>

      <div v-if="step === 1">
        <h2 class="mb-2 text-2xl font-bold text-white">What are your programming languages?</h2>
        <p class="mb-4 text-slate-400">Select all that apply. We will tailor explanations to your background.</p>
        <input
          v-model="langSearch"
          type="search"
          placeholder="Search languages…"
          class="mb-3 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 focus:border-onbober-primary focus:outline-none"
        />
        <div class="grid max-h-56 grid-cols-2 gap-2 overflow-y-auto pr-1">
          <button
            v-for="lang in filteredLanguages"
            :key="lang.label"
            type="button"
            class="flex items-center gap-2 rounded-lg border px-3 py-2.5 text-left text-sm transition"
            :class="
              selectedLangLabels.has(lang.label)
                ? 'border-onbober-primary bg-onbober-primary/10 text-white'
                : 'border-slate-700 text-slate-300 hover:border-slate-500'
            "
            @click="toggleLang(lang.label)"
          >
            <span
              class="flex h-4 w-4 shrink-0 items-center justify-center rounded border text-xs"
              :class="
                selectedLangLabels.has(lang.label)
                  ? 'border-onbober-primary bg-onbober-primary text-white'
                  : 'border-slate-600'
              "
            >{{ selectedLangLabels.has(lang.label) ? '✓' : '' }}</span>
            {{ lang.label }}
          </button>
          <p v-if="filteredLanguages.length === 0" class="col-span-2 py-2 text-center text-sm text-slate-500">No match — will use auto-detect.</p>
        </div>
        <p v-if="selectedLangLabels.size > 0" class="mt-2 text-xs text-slate-500">
          Selected:
          <span
            v-for="label in selectedLangLabels"
            :key="label"
            class="ml-1 inline-flex items-center gap-1 rounded-full border border-slate-600 px-2 py-0.5 text-slate-300"
          >
            {{ label }}
            <button type="button" class="text-slate-500 hover:text-white" @click="toggleLang(label)">✕</button>
          </span>
        </p>
      </div>

      <div v-else-if="step === 2">
        <h2 class="mb-2 text-2xl font-bold text-white">What is your experience level?</h2>
        <p class="mb-6 text-slate-400">This helps calibrate explanation depth.</p>
        <div class="space-y-3">
          <button
            v-for="level in levels"
            :key="level.value"
            type="button"
            class="w-full rounded-lg border px-4 py-3 text-left transition"
            :class="
              selectedLevel === level.value
                ? 'border-onbober-primary bg-onbober-primary/10 text-white'
                : 'border-slate-700 text-slate-300 hover:border-slate-500'
            "
            @click="selectedLevel = level.value"
          >
            {{ level.label }}
          </button>
        </div>
      </div>

      <div v-else>
        <template v-if="loading">
          <h2 class="mb-2 text-2xl font-bold text-white">Setting up your workspace</h2>
          <p class="mb-6 text-slate-400">This may take a moment for larger archives.</p>
          <div class="rounded-lg border border-slate-800 bg-slate-950 px-6 py-8">
            <p v-if="setupSourceLabel" class="mb-4 truncate text-center font-mono text-sm text-slate-400">
              {{ setupSourceLabel }}
            </p>
            <div class="flex justify-center">
              <LoadingStatus
                :active="true"
                :phrases="setupPhrases"
                :show-shimmer="false"
                phrase-class="text-sm text-slate-400"
              />
            </div>
          </div>
        </template>

        <template v-else>
          <h2 class="mb-2 text-2xl font-bold text-white">Where is your codebase?</h2>
          <p class="mb-4 text-slate-400">
            Point to a local folder, paste a GitHub link, or upload a zip archive.
          </p>

          <div class="mb-4 flex gap-2">
            <button
              v-for="opt in sourceOptions"
              :key="opt.value"
              type="button"
              class="flex-1 rounded-lg border px-3 py-2 text-sm transition"
              :class="
                sourceTab === opt.value
                  ? 'border-onbober-primary bg-onbober-primary/10 text-white'
                  : 'border-slate-700 text-slate-400 hover:border-slate-500'
              "
              @click="sourceTab = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>

          <div v-if="sourceTab === 'local'">
            <input
              v-model="localPath"
              type="text"
              placeholder="C:\codebases\linux"
              class="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 placeholder:text-slate-500 focus:border-onbober-primary focus:outline-none"
            />
            <p class="mt-2 text-xs text-slate-500">Absolute path to an existing folder on your machine.</p>
          </div>

          <div v-else-if="sourceTab === 'github'">
            <input
              v-model="githubUrl"
              type="text"
              placeholder="https://github.com/torvalds/linux"
              class="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 placeholder:text-slate-500 focus:border-onbober-primary focus:outline-none"
            />
            <p class="mt-2 text-xs text-slate-500">
              Public repos only. Shallow clone — requires git on the server machine.
            </p>
          </div>

          <div v-else>
            <label
              class="flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-slate-700 bg-slate-950 px-4 py-8 transition hover:border-onbober-primary"
            >
              <span class="mb-2 text-3xl">📦</span>
              <span class="text-sm text-slate-300">
                {{ zipFile ? zipFile.name : 'Click to choose a .zip file' }}
              </span>
              <span class="mt-1 text-xs text-slate-500">Max 200 MB</span>
              <input type="file" accept=".zip" class="hidden" @change="onZipChange" />
            </label>
          </div>
        </template>

        <p v-if="error" class="mt-3 text-sm text-red-400">{{ error }}</p>
      </div>

      <div class="mt-8 flex justify-between">
        <Button v-if="step > 1" variant="ghost" :disabled="loading" @click="prevStep">Back</Button>
        <div v-else />
        <Button variant="primary" :disabled="!canProceed || loading" @click="nextStep">
          {{ step === totalSteps ? (loading ? 'Setting up...' : 'Enter Workspace') : 'Continue' }}
        </Button>
      </div>
    </div>
  </div>
</template>
