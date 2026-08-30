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

const featurePills = ['Source scan', 'Execution flow', 'AI explanations']

/** Navigate to the onboarding flow. */
function initializeWorkspace(): void {
  router.push('/onboarding')
}

const steps = [
  {
    title: 'Drop in your codebase',
    body: 'Point to a local folder, paste a GitHub URL, or upload a zip. No setup beyond that.',
    beaver: false,
  },
  {
    title: 'Pick a symbol',
    body: 'Open any file and click a function or symbol chip in the code drawer.',
    beaver: false,
  },
  {
    title: 'Walk the execution flow',
    body: 'Get an instant graph of what calls what, built straight from a source scan.',
    beaver: true,
  },
  {
    title: 'Expand & understand',
    body: 'Lazily expand branches and get explanations tailored to your language and experience level.',
    beaver: false,
  },
]
</script>

<template>
  <div class="relative flex min-h-screen flex-col overflow-hidden bg-slate-950">
    <!-- Atmospheric background + wandering bober -->
    <div class="pointer-events-none absolute inset-0" aria-hidden="true">
      <div class="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-10%,rgba(255,51,102,0.12),transparent)]" />
      <div class="absolute -left-32 top-1/4 h-96 w-96 rounded-full bg-slate-800/20 blur-3xl" />
      <div class="absolute -right-32 bottom-1/4 h-96 w-96 rounded-full bg-slate-800/15 blur-3xl" />
      <div class="landing-grid absolute inset-0 opacity-40" />
      <img :src="logo" alt="" class="beaver-float absolute -left-6 top-28 h-28 w-auto opacity-[0.06] sm:h-36" />
      <img :src="logo" alt="" class="beaver-float-delayed absolute -right-4 top-[42%] h-20 w-auto opacity-[0.05]" />
      <img :src="logo" alt="" class="absolute bottom-32 left-[8%] h-14 w-auto opacity-[0.04] rotate-6" />
    </div>

    <div class="relative z-10 flex min-h-screen flex-col">
      <!-- Hero -->
      <section class="mx-auto w-full max-w-6xl px-6 pt-10 pb-6 lg:grid lg:grid-cols-2 lg:items-start lg:gap-10 lg:pt-12 lg:pb-6">
        <!-- Copy -->
        <div class="flex flex-col items-center text-center lg:items-start lg:text-left">
          <div class="relative mb-4 flex items-end justify-center gap-3 lg:justify-start">
            <img
              :src="logo"
              alt=""
              class="beaver-bob h-14 w-auto shrink-0 sm:h-16 lg:h-[4.5rem]"
              aria-hidden="true"
            />
            <img :src="wordmark" alt="OnBober" class="h-12 w-auto max-w-full sm:h-14 lg:h-16" />
          </div>
          <p class="mb-2 text-base text-slate-300 lg:text-lg">
            An onboarding compass for complex codebases.
          </p>
          <p class="mb-5 max-w-md text-sm leading-relaxed text-slate-400">
            Walk the execution path of any function. Expand branches on demand and get explanations
            calibrated to your background — instead of reading files top to bottom and hoping it clicks.
          </p>
          <div class="flex flex-wrap items-center justify-center gap-3 lg:justify-start">
            <Button variant="primary" @click="initializeWorkspace">Initialize Workspace</Button>
            <a
              :href="SOURCE_REPO_URL"
              target="_blank"
              rel="noopener noreferrer"
              class="group inline-flex items-center gap-1.5 rounded-lg border border-slate-700/80 bg-slate-900/40 px-3 py-2 text-sm text-slate-300 transition hover:border-slate-600 hover:bg-slate-900/70 hover:text-white"
            >
              <span>Source repo</span>
              <span class="font-mono text-xs text-slate-400 transition group-hover:text-slate-300">{{ SOURCE_REPO_LABEL }}</span>
              <svg class="h-3.5 w-3.5 shrink-0 text-slate-500 transition group-hover:text-slate-300" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
                <path d="M6 3H3v10h10v-3M9 2h5v5M14 2L7 9" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </a>
          </div>
          <ul class="mt-4 flex flex-wrap items-center justify-center gap-1.5 lg:justify-start">
            <li
              v-for="pill in featurePills"
              :key="pill"
              class="rounded-full border border-slate-700/80 bg-slate-900/50 px-2.5 py-0.5 text-[11px] text-slate-300"
            >
              {{ pill }}
            </li>
          </ul>
        </div>

        <!-- Hero visual: flow graph -->
        <div class="relative mt-8 lg:mt-2">
          <img
            :src="logo"
            alt=""
            class="beaver-peek absolute -left-2 -top-5 z-10 h-11 w-auto drop-shadow-lg sm:-left-3 sm:h-12"
            aria-hidden="true"
          />
          <div class="flow-preview-card relative overflow-hidden rounded-2xl border border-slate-800 bg-slate-900/90 shadow-2xl shadow-black/40">
            <div class="flex items-center justify-between gap-3 border-b border-slate-800/80 bg-slate-950/50 px-4 py-2.5">
              <div class="flex items-center gap-2">
                <span class="h-2 w-2 rounded-full bg-onbober-primary/80" />
                <p class="text-[11px] font-semibold uppercase tracking-wider text-onbober-primary">Flow preview</p>
              </div>
              <div class="flex items-center gap-2.5 text-[10px] font-medium uppercase tracking-wide text-slate-500">
                <span class="flex items-center gap-1">
                  <span class="h-2 w-2 rounded-sm border-2 border-green-400 bg-green-950" />
                  Scan
                </span>
                <span class="flex items-center gap-1">
                  <span class="h-2 w-2 rounded-sm border-2 border-dashed border-amber-400 bg-amber-950" />
                  AI
                </span>
                <span class="flex items-center gap-1">
                  <span class="h-2 w-2 rounded-sm border-2 border-onbober-primary bg-slate-800" />
                  Compact
                </span>
              </div>
            </div>

            <svg class="w-full" viewBox="0 0 380 210" role="img" aria-label="Example execution flow graph">
              <defs>
                <marker id="flow-arrow" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto">
                  <path d="M0,0 L6,3 L0,6 Z" fill="#475569" />
                </marker>
                <filter id="node-glow" x="-20%" y="-20%" width="140%" height="140%">
                  <feDropShadow dx="0" dy="0" stdDeviation="3" flood-color="#ff3366" flood-opacity="0.25" />
                </filter>
              </defs>

              <!-- edges -->
              <g fill="none" stroke="#475569" stroke-width="1.5" marker-end="url(#flow-arrow)">
                <path d="M 190 46 L 190 62" />
                <path d="M 190 62 L 190 74 L 108 74 L 108 88" />
                <path d="M 190 62 L 190 74 L 272 74 L 272 88" />
                <path d="M 108 118 L 108 132 L 190 132 L 190 148" />
                <path d="M 272 118 L 272 132 L 190 132" />
              </g>

              <!-- entry — scan -->
              <g>
                <rect x="128" y="10" width="124" height="36" rx="6" fill="#052e16" stroke="#4ade80" stroke-width="2" />
                <text x="190" y="26" text-anchor="middle" fill="#86efac" font-size="11" font-weight="600" font-family="ui-monospace, monospace">AsyncClose()</text>
                <text x="190" y="38" text-anchor="middle" fill="#64748b" font-size="9" font-family="system-ui, sans-serif">entry point</text>
              </g>

              <!-- AI label -->
              <g>
                <rect x="48" y="88" width="120" height="30" rx="6" fill="#451a03" stroke="#fbbf24" stroke-width="2" stroke-dasharray="5 3" />
                <text x="108" y="102" text-anchor="middle" fill="#fde68a" font-size="10" font-weight="600" font-family="ui-monospace, monospace">m.mu.Lock()</text>
                <text x="108" y="113" text-anchor="middle" fill="#64748b" font-size="8" font-family="system-ui, sans-serif">AI label</text>
              </g>

              <!-- branch — scan -->
              <g>
                <rect x="212" y="88" width="120" height="30" rx="6" fill="#052e16" stroke="#4ade80" stroke-width="2" />
                <text x="272" y="102" text-anchor="middle" fill="#86efac" font-size="10" font-weight="600" font-family="ui-monospace, monospace">if m.closed</text>
                <text x="272" y="113" text-anchor="middle" fill="#64748b" font-size="8" font-family="system-ui, sans-serif">branch</text>
              </g>

              <!-- compact callee -->
              <g filter="url(#node-glow)">
                <rect x="118" y="148" width="144" height="36" rx="6" fill="#1e293b" stroke="#ff3366" stroke-width="2" />
                <text x="190" y="164" text-anchor="middle" fill="#f1f5f9" font-size="11" font-weight="600" font-family="ui-monospace, monospace">handleClose()</text>
                <text x="190" y="176" text-anchor="middle" fill="#ff3366" font-size="9" font-family="system-ui, sans-serif">compact callee</text>
              </g>
            </svg>

            <div class="flex items-center justify-between border-t border-slate-800/80 bg-slate-950/40 px-4 py-2 text-[10px] text-slate-500">
              <span>AsyncProducer.go</span>
              <span class="flex items-center gap-1 text-slate-600">
                <img :src="logo" alt="" class="h-4 w-auto opacity-60" aria-hidden="true" />
                mapped by bober
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- How it works -->
      <section id="how-it-works" class="border-t border-slate-900 px-6 py-8">
        <div class="mx-auto max-w-6xl">
          <div class="mb-6 flex items-end justify-center gap-3 sm:justify-start">
            <img :src="logo" alt="" class="h-9 w-auto opacity-80" aria-hidden="true" />
            <div class="text-center sm:text-left">
              <h2 class="text-xl font-bold text-white">How it works</h2>
              <p class="text-sm text-slate-400">Four steps from unfamiliar repo to mental model.</p>
            </div>
          </div>
          <div class="relative grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <div
              class="pointer-events-none absolute left-[12.5%] right-[12.5%] top-3.5 hidden h-px bg-slate-800 lg:block"
              aria-hidden="true"
            />
            <div
              v-for="(step, i) in steps"
              :key="step.title"
              class="relative overflow-hidden rounded-lg border border-slate-800 bg-slate-900 p-4"
            >
              <img
                v-if="step.beaver"
                :src="logo"
                alt=""
                class="pointer-events-none absolute -bottom-2 -right-2 h-10 w-auto opacity-[0.15]"
                aria-hidden="true"
              />
              <span class="relative z-10 mb-2 inline-flex h-7 w-7 items-center justify-center rounded-full bg-onbober-primary/10 text-xs font-semibold text-onbober-primary">
                {{ i + 1 }}
              </span>
              <h3 class="relative z-10 mb-1 text-sm font-semibold text-white">{{ step.title }}</h3>
              <p class="relative z-10 text-xs leading-relaxed text-slate-300">{{ step.body }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Team -->
      <footer id="built-by" class="mt-auto border-t border-slate-900 px-6 py-6">
        <div class="mx-auto max-w-6xl">
          <div class="mb-4 flex flex-col items-center gap-2">
            <img :src="logo" alt="" class="h-8 w-auto opacity-70" aria-hidden="true" />
            <p class="text-xs font-semibold uppercase tracking-wider text-slate-500">Built by</p>
          </div>
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
  </div>
</template>

<style scoped>
.landing-grid {
  background-image: radial-gradient(circle, #1e293b 1px, transparent 1px);
  background-size: 24px 24px;
}

.beaver-float {
  animation: beaver-drift 9s ease-in-out infinite;
}

.beaver-float-delayed {
  animation: beaver-drift-reverse 11s ease-in-out infinite;
}

.beaver-bob {
  animation: beaver-bob 3s ease-in-out infinite;
}

.beaver-peek {
  animation: beaver-peek 4s ease-in-out infinite;
}

@keyframes beaver-drift {
  0%, 100% { transform: translateY(0) rotate(-12deg); }
  50% { transform: translateY(-10px) rotate(-8deg); }
}

@keyframes beaver-drift-reverse {
  0%, 100% { transform: translateY(0) rotate(12deg); }
  50% { transform: translateY(-8px) rotate(8deg); }
}

@keyframes beaver-bob {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}

@keyframes beaver-peek {
  0%, 100% { transform: translateY(0) rotate(-12deg); }
  50% { transform: translateY(3px) rotate(-8deg); }
}

.flow-preview-card {
  background-image: radial-gradient(circle at 50% 0%, rgba(255, 51, 102, 0.06), transparent 55%);
}
</style>
