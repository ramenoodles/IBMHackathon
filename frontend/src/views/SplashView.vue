<script setup lang="ts">
/**
 * Landing page for OnBober: hero, purpose, how-it-works, and feature overview.
 */
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import logo from '@/assets/logo.png'
import wordmark from '@/assets/dark_headline.png'

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

const features = [
  {
    title: 'Instant, scan-first graphs',
    body: 'Flow graphs are built from your source directly, so they load immediately even if the LLM is offline.',
  },
  {
    title: 'Safe on huge repos',
    body: 'Branch expansion is bounded and requires your confirmation, so you never trigger a compute bomb by accident.',
  },
  {
    title: 'Tailored to you',
    body: 'Explanation depth is calibrated to your primary language and experience level from onboarding.',
  },
  {
    title: 'Bring any codebase',
    body: 'Local path, GitHub clone, or zip upload — start from wherever the code already lives.',
  },
]
</script>

<template>
  <div class="min-h-screen bg-slate-950">
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

    <!-- Why OnBober -->
    <section class="border-t border-slate-900 px-6 py-20">
      <div class="mx-auto max-w-5xl">
        <h2 class="mb-2 text-center text-2xl font-bold text-white">Why OnBober</h2>
        <p class="mb-12 text-center text-sm text-slate-500">
          Built for the moment you're dropped into code you didn't write.
        </p>
        <div class="grid gap-4 sm:grid-cols-2">
          <div
            v-for="feature in features"
            :key="feature.title"
            class="rounded-xl border border-slate-800 bg-slate-900 p-5"
          >
            <h3 class="mb-1.5 font-semibold text-white">{{ feature.title }}</h3>
            <p class="text-sm text-slate-400">{{ feature.body }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Final CTA -->
    <section class="border-t border-slate-900 px-6 py-20 text-center">
      <h2 class="mb-3 text-2xl font-bold text-white">Ready to onboard faster?</h2>
      <p class="mb-8 text-sm text-slate-500">Point OnBober at a codebase and pick your first symbol.</p>
      <Button variant="primary" @click="initializeWorkspace">Initialize Workspace</Button>
    </section>
  </div>
</template>
