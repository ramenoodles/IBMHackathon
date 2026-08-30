<script setup lang="ts">
/**
 * Multi-step onboarding flow collecting developer context for tailored AI analysis.
 */
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import BeaverRepoLoader from '@/components/ui/BeaverRepoLoader.vue'
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
import { DEMO_REPO_LABEL, DEMO_REPO_URL, formatRepoDisplayLabel } from '@/constants/demoRepo'
import {
  BEAVER_LOADER_DURATION_MS,
  BEAVER_LOADER_HOLD_MS,
  delay,
} from '@/constants/beaverLoader'
import {
  EXPERIENCE_EFFECT_BULLETS,
  EXPERIENCE_EFFECT_SUMMARY,
  EXPERIENCE_LEVELS,
} from '@/constants/experienceLevel'
import {
  PROGRAMMING_LANGUAGES,
  languageLabelsFromStored,
  storedLanguagesFromLabels,
} from '@/constants/programmingLanguages'

const router = useRouter()
const { loading, error, setupLocal, setupGitHub, setupDemo, setupZip } = useWorkspaceSetup()

const step = ref(1)
const totalSteps = 3

const langSearch = ref('')
const filteredLanguages = computed(() => {
  const q = langSearch.value.toLowerCase().trim()
  if (!q) return PROGRAMMING_LANGUAGES
  return PROGRAMMING_LANGUAGES.filter((l) => l.label.toLowerCase().includes(q))
})

const selectedLangLabels = ref<Set<string>>(new Set(languageLabelsFromStored(userContext.value.primaryLanguage || '')))

function toggleLang(label: string): void {
  if (selectedLangLabels.value.has(label)) {
    selectedLangLabels.value.delete(label)
  } else {
    selectedLangLabels.value.add(label)
  }
  // Trigger reactivity — Set mutations are not tracked automatically.
  selectedLangLabels.value = new Set(selectedLangLabels.value)
}

const selectedLanguagesValue = computed(() => storedLanguagesFromLabels(selectedLangLabels.value))
const selectedLevel = ref<ExperienceLevel>(userContext.value.experienceLevel || 'junior')

const sourceTab = ref<WorkspaceSource>('demo')
const localPath = ref('')
const githubUrl = ref('')
const zipFile = ref<File | null>(null)

const sourceOptions: { value: WorkspaceSource; label: string }[] = [
  { value: 'demo', label: 'Try demo' },
  { value: 'local', label: 'Local path' },
  { value: 'github', label: 'GitHub repo' },
  { value: 'zip', label: 'Upload ZIP' },
]

const setupPhrases = computed(() => {
  if (sourceTab.value === 'zip') return ZIP_SETUP_PHRASES
  if (sourceTab.value === 'demo' || sourceTab.value === 'github') return GITHUB_SETUP_PHRASES
  if (sourceTab.value === 'local') return LOCAL_SETUP_PHRASES
  return WORKSPACE_SETUP_PHRASES
})

const setupSourceLabel = computed(() => {
  if (sourceTab.value === 'demo') return DEMO_REPO_LABEL
  if (sourceTab.value === 'zip' && zipFile.value) return zipFile.value.name
  if (sourceTab.value === 'github' && githubUrl.value.trim()) {
    return formatRepoDisplayLabel(githubUrl.value)
  }
  if (sourceTab.value === 'local' && localPath.value.trim()) return localPath.value.trim()
  return ''
})

const githubPlaceholder = `${DEMO_REPO_LABEL} or ${DEMO_REPO_URL}`

const useBeaverLoader = computed(
  () => sourceTab.value === 'demo' || sourceTab.value === 'github' || sourceTab.value === 'zip',
)

/** Keeps the beaver loader visible until the animation finishes, even if the API returns early. */
const setupAnimating = ref(false)

const showBeaverSetup = computed(() => loading.value || setupAnimating.value)

const canProceed = computed(() => {
  if (step.value === 1) return selectedLangLabels.value.size > 0
  if (step.value === 2) return Boolean(selectedLevel.value)
  if (step.value === 3) {
    if (sourceTab.value === 'demo') return true
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
    const withBeaver = useBeaverLoader.value
    if (withBeaver) setupAnimating.value = true
    const startedAt = performance.now()
    try {
      let workspace: { id: string; name: string } | undefined
      if (sourceTab.value === 'demo') {
        workspace = await setupDemo()
      } else if (sourceTab.value === 'local') {
        workspace = await setupLocal(localPath.value.trim())
      } else if (sourceTab.value === 'github') {
        workspace = await setupGitHub(githubUrl.value.trim())
      } else if (sourceTab.value === 'zip' && zipFile.value) {
        workspace = await setupZip(zipFile.value)
      }
      if (!workspace) {
        setupAnimating.value = false
        return
      }

      if (withBeaver) {
        const elapsed = performance.now() - startedAt
        const remaining = BEAVER_LOADER_DURATION_MS + BEAVER_LOADER_HOLD_MS - elapsed
        if (remaining > 0) await delay(remaining)
      }

      updateUserContext({ workspaceId: workspace.id, workspaceName: workspace.name })
      await router.push('/workspace')
    } catch {
      setupAnimating.value = false
    }
  }
  step.value += 1
}

/** Return to the previous onboarding step. */
function prevStep(): void {
  if (step.value > 1) step.value -= 1
}

/** Submit step 3 from a text input when Enter is pressed. */
function onSourceKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Enter' || step.value !== totalSteps || !canProceed.value || loading.value || setupAnimating.value) return
  event.preventDefault()
  void nextStep()
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-950 px-6">
    <div
      class="w-full rounded-xl border border-slate-800 bg-slate-900 p-8"
      :class="showBeaverSetup && step === totalSteps && useBeaverLoader ? 'max-w-2xl' : 'max-w-lg'"
    >
      <div v-if="!(showBeaverSetup && step === totalSteps && useBeaverLoader)" class="mb-6 flex justify-center">
        <img
          :src="logo"
          alt="OnBober mascot"
          class="h-24 w-auto transition-opacity"
          :class="loading ? 'animate-pulse' : ''"
        />
      </div>
      <p v-if="!(showBeaverSetup && step === totalSteps && useBeaverLoader)" class="mb-2 text-sm text-onbober-primary">
        Step {{ step }} of {{ totalSteps }}
      </p>

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
        <p class="mb-4 text-slate-400">{{ EXPERIENCE_EFFECT_SUMMARY }}</p>
        <div class="space-y-3">
          <button
            v-for="level in EXPERIENCE_LEVELS"
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
            <span class="block font-medium">{{ level.label }}</span>
            <span class="mt-1 block text-sm text-slate-400">{{ level.description }}</span>
          </button>
        </div>
        <div class="mt-5 rounded-lg border border-slate-800 bg-slate-950/60 px-4 py-3">
          <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">What this affects</p>
          <ul class="mt-2 list-inside list-disc space-y-1 text-sm text-slate-400">
            <li v-for="item in EXPERIENCE_EFFECT_BULLETS" :key="item">{{ item }}</li>
          </ul>
        </div>
      </div>

      <div v-else>
        <template v-if="showBeaverSetup && useBeaverLoader">
          <BeaverRepoLoader
            :active="showBeaverSetup"
            :phrases="setupPhrases"
            :source-label="setupSourceLabel"
          />
        </template>
        <template v-else-if="loading">
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
            Try the IBM Bob demo repo, or point to your own codebase.
          </p>

          <div class="mb-4 grid grid-cols-2 gap-2">
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

          <div v-if="sourceTab === 'demo'" class="rounded-lg border border-onbober-primary/30 bg-onbober-primary/5 px-4 py-4">
            <p class="text-sm font-medium text-white">{{ DEMO_REPO_LABEL }}</p>
            <p class="mt-1 text-sm text-slate-400">
              IBM's Go Kafka client library — explore flow tracing with no setup required.
            </p>
            <p class="mt-3 font-mono text-xs text-slate-500">{{ DEMO_REPO_URL }}</p>
          </div>

          <div v-else-if="sourceTab === 'local'">
            <input
              v-model="localPath"
              type="text"
              placeholder="C:\codebases\linux"
              class="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 placeholder:text-slate-500 focus:border-onbober-primary focus:outline-none"
              @keydown="onSourceKeydown"
            />
            <p class="mt-2 text-xs text-slate-500">Absolute path to an existing folder on your machine.</p>
          </div>

          <div v-else-if="sourceTab === 'github'">
            <input
              v-model="githubUrl"
              type="text"
              :placeholder="githubPlaceholder"
              class="w-full rounded-lg border border-slate-700 bg-slate-950 px-4 py-3 text-slate-100 placeholder:text-slate-500 focus:border-onbober-primary focus:outline-none"
              @keydown="onSourceKeydown"
            />
            <p class="mt-2 text-xs text-slate-500">
              Public repos only — use <span class="font-mono">owner/repo</span> or a full URL. Shallow clone; requires git on the server.
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
        <Button v-if="step > 1" variant="ghost" :disabled="loading || setupAnimating" @click="prevStep">Back</Button>
        <div v-else />
        <Button variant="primary" :disabled="!canProceed || loading || setupAnimating" @click="nextStep">
          {{ step === totalSteps ? (loading || setupAnimating ? 'Setting up...' : 'Enter Workspace') : 'Continue' }}
        </Button>
      </div>
    </div>
  </div>
</template>
