<script setup lang="ts">
/**
 * Multi-step onboarding flow collecting developer context for tailored AI analysis.
 */
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import {
  type ExperienceLevel,
  updateUserContext,
  userContext,
} from '@/store/userContext'
import {
  type WorkspaceSource,
  useWorkspaceSetup,
} from '@/composables/useWorkspaceSetup'

const router = useRouter()
const { loading, error, setupLocal, setupGitHub, setupZip } = useWorkspaceSetup()

const step = ref(1)
const totalSteps = 3

const languages = ['C/C++', 'Python', 'Rust', 'Go', 'Other']
const levels: { value: ExperienceLevel; label: string }[] = [
  { value: 'junior', label: 'Junior SWE' },
  { value: 'mid', label: 'Mid-level SWE' },
  { value: 'senior', label: 'Senior SWE' },
]

const selectedLanguage = ref(userContext.value.primaryLanguage || '')
const selectedLevel = ref<ExperienceLevel>(userContext.value.experienceLevel || 'junior')

const sourceTab = ref<WorkspaceSource>('local')
const localPath = ref(userContext.value.workspacePath || '')
const githubUrl = ref('')
const zipFile = ref<File | null>(null)

const sourceOptions: { value: WorkspaceSource; label: string }[] = [
  { value: 'local', label: 'Local path' },
  { value: 'github', label: 'GitHub repo' },
  { value: 'zip', label: 'Upload ZIP' },
]

const canProceed = computed(() => {
  if (step.value === 1) return Boolean(selectedLanguage.value)
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
    updateUserContext({ primaryLanguage: selectedLanguage.value })
  } else if (step.value === 2) {
    updateUserContext({ experienceLevel: selectedLevel.value })
  } else if (step.value === 3) {
    try {
      let workspacePath = ''
      if (sourceTab.value === 'local') {
        workspacePath = await setupLocal(localPath.value.trim())
      } else if (sourceTab.value === 'github') {
        workspacePath = await setupGitHub(githubUrl.value.trim())
      } else if (sourceTab.value === 'zip' && zipFile.value) {
        workspacePath = await setupZip(zipFile.value)
      }
      updateUserContext({ workspacePath })
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
      <p class="mb-2 text-sm text-onbober-primary">Step {{ step }} of {{ totalSteps }}</p>

      <div v-if="step === 1">
        <h2 class="mb-2 text-2xl font-bold text-white">What is your primary programming language?</h2>
        <p class="mb-6 text-slate-400">We will tailor explanations to your background.</p>
        <div class="grid grid-cols-2 gap-3">
          <button
            v-for="lang in languages"
            :key="lang"
            type="button"
            class="rounded-lg border px-4 py-3 text-left transition"
            :class="
              selectedLanguage === lang
                ? 'border-onbober-primary bg-onbober-primary/10 text-white'
                : 'border-slate-700 text-slate-300 hover:border-slate-500'
            "
            @click="selectedLanguage = lang"
          >
            {{ lang }}
          </button>
        </div>
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

        <p v-if="error" class="mt-3 text-sm text-red-400">{{ error }}</p>
        <p v-if="loading" class="mt-3 text-sm text-onbober-primary">
          {{ sourceTab === 'github' ? 'Cloning repository...' : sourceTab === 'zip' ? 'Extracting archive...' : 'Validating path...' }}
        </p>
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
