<script setup lang="ts">
/**
 * Landing page for OnBober: hero, purpose, how-it-works, and feature overview.
 */
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import logo from '@/assets/logo.png'
import wordmark from '@/assets/dark_headline.png'
import { SOURCE_REPO_LABEL, SOURCE_REPO_URL } from '@/constants/projectRepo'
import { TEAM_MEMBERS } from '@/constants/teamMembers'

const router = useRouter()

/** Navigate to the onboarding flow. */
function initializeWorkspace(): void {
  router.push('/onboarding')
}

const steps = [
  {
    title: 'Drop in your codebase',
    body: 'Point to a local folder, paste a GitHub URL, or upload a zip. No setup beyond that.',
  },
  {
    title: 'Pick a symbol',
    body: 'Open any file and click a function or symbol chip in the code drawer.',
  },
  {
    title: 'Walk the execution flow',
    body: 'Get an instant graph of what calls what, built straight from a source scan.',
  },
  {
    title: 'Expand & understand',
    body: 'Lazily expand branches and get explanations tailored to your language and experience level.',
  },
]
</script>

<template>
  <div class="flex min-h-screen flex-col bg-slate-950">
    <!-- Hero -->
    <section class="flex flex-col items-center justify-center px-6 pt-24 pb-20 text-center">
      <img :src="logo" alt="OnBober mascot" class="mb-8 h-28 w-auto" />
      <img :src="wordmark" alt="OnBober" class="mb-3 h-16 w-auto max-w-full sm:h-20 md:h-24" />
      <p class="mb-4 max-w-xl text-lg text-slate-400">
        An onboarding compass for complex codebases.
      </p>
      <p class="mb-10 max-w-lg text-sm text-slate-500">
        Walk the execution path of any function. Expand branches on demand and get explanations
        calibrated to your background — instead of reading files top to bottom and hoping it clicks.
      </p>
      <Button variant="primary" @click="initializeWorkspace">Initialize Workspace</Button>
      <a
        :href="SOURCE_REPO_URL"
        target="_blank"
        rel="noopener noreferrer"
        class="mt-5 inline-flex items-center gap-1.5 text-sm text-slate-500 transition hover:text-onbober-primary"
      >
        View source repo
        <span class="font-mono text-xs text-slate-600">({{ SOURCE_REPO_LABEL }})</span>
        <svg class="h-3.5 w-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
          <path d="M6 3H3v10h10v-3M9 2h5v5M14 2L7 9" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </a>
    </section>

    <!-- How it works -->
    <section class="border-t border-slate-900 px-6 py-20">
      <div class="mx-auto max-w-5xl">
        <h2 class="mb-2 text-center text-2xl font-bold text-white">How it works</h2>
        <p class="mb-12 text-center text-sm text-slate-500">Four steps from unfamiliar repo to mental model.</p>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div
            v-for="(step, i) in steps"
            :key="step.title"
            class="rounded-xl border border-slate-800 bg-slate-900 p-5"
          >
            <span class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-onbober-primary/10 text-sm font-semibold text-onbober-primary">
              {{ i + 1 }}
            </span>
            <h3 class="mb-1.5 font-semibold text-white">{{ step.title }}</h3>
            <p class="text-sm text-slate-400">{{ step.body }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Team -->
    <footer class="mt-auto border-t border-slate-900 px-6 py-10">
      <div class="mx-auto max-w-5xl">
        <p class="mb-5 text-center text-xs font-semibold uppercase tracking-wider text-slate-500">Built by</p>
        <ul class="flex flex-nowrap items-start justify-center gap-6 overflow-x-auto pb-1 sm:gap-8 md:gap-10">
          <li
            v-for="member in TEAM_MEMBERS"
            :key="member.name"
            class="flex shrink-0 flex-col items-center gap-2 text-center"
          >
            <span class="whitespace-nowrap text-sm font-medium text-slate-300">{{ member.name }}</span>
            <div class="flex items-center justify-center gap-3">
              <a
                v-if="member.github"
                :href="member.github"
                target="_blank"
                rel="noopener noreferrer"
                class="text-slate-500 transition hover:text-white"
                :aria-label="`${member.name} on GitHub`"
                title="GitHub"
              >
                <svg class="h-4 w-4" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
                  <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8z" />
                </svg>
              </a>
              <a
                v-if="member.linkedin"
                :href="member.linkedin"
                target="_blank"
                rel="noopener noreferrer"
                class="text-slate-500 transition hover:text-[#0a66c2]"
                :aria-label="`${member.name} on LinkedIn`"
                title="LinkedIn"
              >
                <svg class="h-4 w-4" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
                  <path d="M0 1.146C0 .513.526 0 1.175 0h13.65C15.474 0 16 .513 16 1.146v13.708c0 .633-.526 1.146-1.175 1.146H1.175A1.148 1.148 0 0 1 0 14.854V1.146zm4.943 12.248V6.169H2.542v7.225h2.401zm-1.2-8.212c.837 0 1.358-.554 1.358-1.248-.015-.709-.52-1.248-1.342-1.248S4.4 3.226 4.4 3.934c0 .694.521 1.248 1.327 1.248h.016zm4.908 8.212V9.359c0-.216.016-.432.08-.586.173-.431.568-.878 1.232-.878.869 0 1.216.662 1.216 1.634v3.865h2.401V9.25c0-2.22-1.184-3.252-2.764-3.252-1.274 0-1.845.7-2.165 1.193v.025h-.016a5.54 5.54 0 0 1 .016-.025V6.169h-2.4c.03.678 0 7.225 0 7.225h2.4z" />
                </svg>
              </a>
            </div>
          </li>
        </ul>
      </div>
    </footer>
  </div>
</template>
