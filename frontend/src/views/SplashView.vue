<script setup lang="ts">
/**
 * Landing page for OnBober: hero, flow preview, features, and how-it-works.
 */
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import logo from '@/assets/logo.png'
import wordmark from '@/assets/dark_headline.png'
import { FLOW_EDGE_COLORS, FLOW_EDGE_LEGEND } from '@/composables/useFlowMermaid'
import { SOURCE_REPO_LABEL, SOURCE_REPO_URL } from '@/constants/projectRepo'
import TeamCreditsBar from '@/components/ui/TeamCreditsBar.vue'

const router = useRouter()

const features = [
  {
    title: 'Verified CFG scan',
    body: 'Execution flow is built from static analysis first, so structure you can trust appears before any AI labels.',
  },
  {
    title: 'Progressive reveal',
    body: 'Steps unfold one at a time so you follow the path instead of staring at an entire call graph.',
  },
  {
    title: 'Colored branch arrows',
    body: 'True, false, and sequential edges use distinct colors so control flow reads instantly.',
  },
  {
    title: 'Brief AI labels',
    body: 'Watsonx enriches scan titles with short Brief summaries on each step.',
  },
  {
    title: 'Deep dive on demand',
    body: 'Open or hide per-step explanations anytime, with verified detail plus inferred context when available.',
  },
  {
    title: 'Experience tuning',
    body: 'Junior, mid, or senior adjusts label depth and how “Explain this step” is written.',
  },
  {
    title: 'Language analogies',
    body: 'Optional cross-language comparisons relate unfamiliar code to languages you already know.',
  },
  {
    title: 'Callee flow preview',
    body: 'Folded calls stay compact in the main graph. Open View code flow for a scan-only callee map.',
  },
]

const steps = [
  {
    title: 'Share your background',
    body: 'Pick languages you know and a junior / mid / senior level so labels and explanations match how you think.',
    beaver: false,
  },
  {
    title: 'Load any codebase',
    body: 'Try the demo repo, paste a GitHub URL, point at a local folder, or upload a zip.',
    beaver: false,
  },
  {
    title: 'Pick a file & symbol',
    body: 'Browse the explorer, open a file, then choose a function from the symbol bar. Bober scans the CFG right away.',
    beaver: true,
  },
  {
    title: 'Trace, expand, understand',
    body: 'Reveal steps progressively, expand branches inline, preview callee flows, jump to source, or ask Bober for a deep dive when you want more.',
    beaver: false,
  },
]

/** Navigate to the onboarding flow. */
function initializeWorkspace(): void {
  router.push('/onboarding')
}
</script>

<template>
  <div class="relative flex min-h-screen flex-col overflow-hidden bg-slate-950">
    <!-- Atmospheric background -->
    <div class="pointer-events-none absolute inset-0" aria-hidden="true">
      <div class="absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-10%,rgba(255,51,102,0.12),transparent)]" />
      <div class="absolute -left-32 top-1/4 h-96 w-96 rounded-full bg-slate-800/20 blur-3xl" />
      <div class="absolute -right-32 bottom-1/4 h-96 w-96 rounded-full bg-slate-800/15 blur-3xl" />
      <div class="landing-grid absolute inset-0 opacity-40" />
      <img :src="logo" alt="" class="beaver-float absolute -left-6 top-28 h-28 w-auto opacity-[0.06] sm:h-36" />
      <img :src="logo" alt="" class="beaver-float-delayed absolute -right-4 top-[42%] h-20 w-auto opacity-[0.05]" />
    </div>

    <div class="relative z-10 flex min-h-screen flex-col">
      <!-- Hero -->
      <section class="mx-auto w-full max-w-6xl px-6 pt-10 pb-6 lg:grid lg:grid-cols-2 lg:items-start lg:gap-10 lg:pt-12 lg:pb-8">
        <div class="flex flex-col items-center text-center lg:items-start lg:text-left">
          <div class="relative mb-4 flex items-end justify-center gap-3 lg:justify-start">
            <img
              :src="logo"
              alt="Bober the Bober"
              class="beaver-bob h-14 w-auto shrink-0 sm:h-16 lg:h-[4.5rem]"
            />
            <img :src="wordmark" alt="OnBober" class="h-12 w-auto max-w-full sm:h-14 lg:h-16" />
          </div>
          <p class="mb-1 text-base text-slate-300 lg:text-lg">
            An onboarding compass for complex codebases.
          </p>
          <p class="mb-4 text-sm font-medium text-onbober-primary/90">
            Bober the Bober is here to help you chew through the hard parts.
          </p>
          <p class="mb-5 max-w-md text-sm leading-relaxed text-slate-400">
            Scan a function’s control-flow graph, walk it step by step, and layer on AI Brief labels
            and explanations tuned to your languages and experience. Bober maps the path, labels the
            steps, and sticks around in the workspace whenever you need a nudge.
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
        </div>

        <!-- Workspace-style flow preview -->
        <div class="relative mt-8 lg:mt-0">
          <img
            :src="logo"
            alt=""
            class="beaver-peek absolute -left-2 -top-5 z-10 h-11 w-auto drop-shadow-lg sm:-left-3 sm:h-12"
            aria-hidden="true"
          />
          <div class="flow-preview-card relative overflow-hidden rounded-2xl border border-slate-800 bg-slate-900/95 shadow-2xl shadow-black/40">
            <!-- Toolbar -->
            <div class="flex flex-wrap items-center justify-between gap-2 border-b border-slate-800/80 bg-slate-950/60 px-3 py-2">
              <div class="flex items-center gap-2">
                <span class="h-2 w-2 rounded-full bg-onbober-primary/80" />
                <p class="text-[10px] font-semibold uppercase tracking-wider text-onbober-primary">Workspace preview</p>
              </div>
              <div class="flex items-center gap-2 text-[9px] font-medium uppercase tracking-wide text-slate-500" aria-label="Flow arrow legend">
                <span
                  v-for="item in FLOW_EDGE_LEGEND"
                  :key="item.key"
                  class="flex items-center gap-0.5"
                  :title="item.title"
                >
                  <svg class="h-1.5 w-3.5" viewBox="0 0 20 8" fill="none" aria-hidden="true">
                    <path
                      d="M0 4h14M11 1l4 3-4 3"
                      :stroke="item.color"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  {{ item.label }}
                </span>
              </div>
            </div>

            <div class="grid grid-cols-[72px_1fr_88px] border-b border-slate-800/60">
              <!-- Steps strip -->
              <div class="border-r border-slate-800/60 bg-slate-950/40 px-1.5 py-2">
                <p class="mb-1.5 text-[8px] font-semibold uppercase tracking-wider text-slate-600">Steps</p>
                <ol class="space-y-1">
                  <li class="rounded border border-onbober-primary/40 bg-onbober-primary/5 px-1 py-0.5">
                    <span class="text-[8px] font-bold text-onbober-primary">1</span>
                    <p class="truncate text-[7px] leading-tight text-slate-300">Entry</p>
                  </li>
                  <li class="rounded px-1 py-0.5">
                    <span class="text-[8px] text-slate-600">2</span>
                    <p class="truncate text-[7px] leading-tight text-slate-500">Lock mutex</p>
                  </li>
                  <li class="rounded px-1 py-0.5">
                    <span class="text-[8px] text-slate-600">3</span>
                    <p class="truncate text-[7px] leading-tight text-slate-500">Branch</p>
                  </li>
                </ol>
              </div>

              <!-- Flow diagram -->
              <svg class="w-full bg-slate-950/30" viewBox="0 0 280 168" role="img" aria-label="Example execution flow with colored branch arrows">
                <defs>
                  <marker id="arrow-flow" markerWidth="7" markerHeight="7" refX="5" refY="3.5" orient="auto">
                    <path d="M0,0 L7,3.5 L0,7 Z" :fill="FLOW_EDGE_COLORS.default" />
                  </marker>
                  <marker id="arrow-true" markerWidth="7" markerHeight="7" refX="5" refY="3.5" orient="auto">
                    <path d="M0,0 L7,3.5 L0,7 Z" :fill="FLOW_EDGE_COLORS.true" />
                  </marker>
                  <marker id="arrow-false" markerWidth="7" markerHeight="7" refX="5" refY="3.5" orient="auto">
                    <path d="M0,0 L7,3.5 L0,7 Z" :fill="FLOW_EDGE_COLORS.false" />
                  </marker>
                </defs>

                <!-- start -->
                <path
                  d="M 140 8 L 140 28"
                  fill="none"
                  :stroke="FLOW_EDGE_COLORS.default"
                  stroke-width="2"
                  marker-end="url(#arrow-flow)"
                />
                <rect x="88" y="28" width="104" height="28" rx="5" fill="#1a2e1a" :stroke="FLOW_EDGE_COLORS.true" stroke-width="1.5" />
                <text x="140" y="46" text-anchor="middle" fill="#86efac" font-size="9" font-weight="600" font-family="ui-monospace, monospace">AsyncClose()</text>

                <!-- to branch diamond -->
                <path
                  d="M 140 56 L 140 72"
                  fill="none"
                  :stroke="FLOW_EDGE_COLORS.default"
                  stroke-width="2"
                  marker-end="url(#arrow-flow)"
                />
                <polygon
                  points="140,72 168,92 140,112 112,92"
                  fill="#1a2e1a"
                  :stroke="FLOW_EDGE_COLORS.true"
                  stroke-width="1.5"
                />
                <text x="140" y="95" text-anchor="middle" fill="#86efac" font-size="8" font-family="ui-monospace, monospace">closed?</text>

                <!-- true branch -->
                <path
                  d="M 112 92 L 56 92 L 56 118"
                  fill="none"
                  :stroke="FLOW_EDGE_COLORS.true"
                  stroke-width="2"
                  marker-end="url(#arrow-true)"
                />
                <rect x="8" y="118" width="96" height="26" rx="5" fill="#2a1f0a" stroke="#fbbf24" stroke-width="1.5" stroke-dasharray="4 2" />
                <text x="56" y="134" text-anchor="middle" fill="#fde68a" font-size="8" font-weight="600" font-family="ui-monospace, monospace">return early</text>

                <!-- false branch -->
                <path
                  d="M 168 92 L 224 92 L 224 118"
                  fill="none"
                  :stroke="FLOW_EDGE_COLORS.false"
                  stroke-width="2"
                  marker-end="url(#arrow-false)"
                />
                <rect x="176" y="118" width="96" height="26" rx="5" fill="#1e293b" :stroke="FLOW_EDGE_COLORS.false" stroke-width="1.5" opacity="0.9" />
                <text x="224" y="132" text-anchor="middle" fill="#fca5a5" font-size="8" font-family="ui-monospace, monospace">close ch.</text>
                <text x="224" y="141" text-anchor="middle" fill="#64748b" font-size="7" font-family="system-ui, sans-serif">false</text>

                <!-- merge to compact -->
                <path
                  d="M 56 144 L 56 152 L 140 152"
                  fill="none"
                  :stroke="FLOW_EDGE_COLORS.default"
                  stroke-width="1.5"
                  marker-end="url(#arrow-flow)"
                />
                <path
                  d="M 224 144 L 224 152 L 140 152"
                  fill="none"
                  :stroke="FLOW_EDGE_COLORS.default"
                  stroke-width="1.5"
                />

                <!-- compact callee -->
                <rect x="92" y="148" width="96" height="18" rx="4" fill="#1e293b" stroke="#ff3366" stroke-width="1.5" />
                <text x="140" y="160" text-anchor="middle" fill="#f1f5f9" font-size="8" font-weight="600" font-family="ui-monospace, monospace">handleClose()</text>

                <!-- edge labels -->
                <rect x="72" y="78" width="22" height="12" rx="3" fill="rgba(21,45,28,0.96)" :stroke="FLOW_EDGE_COLORS.true" stroke-width="0.75" />
                <text x="83" y="87" text-anchor="middle" :fill="FLOW_EDGE_COLORS.true" font-size="7" font-weight="600">true</text>
                <rect x="186" y="78" width="24" height="12" rx="3" fill="rgba(52,22,22,0.96)" :stroke="FLOW_EDGE_COLORS.false" stroke-width="0.75" />
                <text x="198" y="87" text-anchor="middle" :fill="FLOW_EDGE_COLORS.false" font-size="7" font-weight="600">false</text>
                <rect x="128" y="14" width="24" height="12" rx="3" fill="rgba(13,30,56,0.96)" :stroke="FLOW_EDGE_COLORS.default" stroke-width="0.75" />
                <text x="140" y="23" text-anchor="middle" :fill="FLOW_EDGE_COLORS.default" font-size="7" font-weight="600">start</text>
              </svg>

              <!-- Details strip -->
              <div class="border-l border-slate-800/60 bg-slate-950/40 px-1.5 py-2">
                <p class="mb-1.5 text-[8px] font-semibold uppercase tracking-wider text-slate-600">Details</p>
                <p class="text-[8px] font-semibold text-white">Lock mutex</p>
                <p class="mt-1 rounded border border-dashed border-amber-600/40 bg-amber-900/20 px-1 py-0.5 text-[7px] font-bold text-amber-400">Brief</p>
                <p class="mt-1 text-[7px] leading-snug text-amber-300/80">Acquires lock before mutating state.</p>
                <p class="mt-2 text-[7px] text-onbober-primary">View code flow ↗</p>
              </div>
            </div>

            <div class="flex items-center justify-between border-t border-slate-800/80 bg-slate-950/40 px-3 py-1.5 text-[9px] text-slate-500">
              <span class="font-mono">async_producer.go · NewAsyncProducer</span>
              <span class="flex items-center gap-1 text-slate-600">
                <img :src="logo" alt="" class="h-3.5 w-auto opacity-60" aria-hidden="true" />
                Bober's on it
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- Meet Bober -->
      <section id="meet-bober" class="border-t border-slate-900 px-6 py-8">
        <div class="mx-auto flex max-w-6xl flex-col items-center gap-6 sm:flex-row sm:items-start">
          <img
            :src="logo"
            alt="Bober the Bober"
            class="beaver-bob h-20 w-auto shrink-0 sm:h-24"
          />
          <div class="text-center sm:text-left">
            <h2 class="text-lg font-bold text-white">Meet Bober the Bober</h2>
            <p class="mt-2 max-w-2xl text-sm leading-relaxed text-slate-400">
              That’s the little beaver dude. Bober is OnBober’s mascot and your guide: gnawing through
              repo setup, cheering while flow maps load, and peeking from corners of the workspace when
              you toggle on little bobers. He’s not here to replace reading code. He’s here to point at
              the right branch, explain the step in your words, and make a scary codebase feel walkable.
            </p>
            <p class="mt-2 text-sm text-slate-500">
              Initialize a workspace and Bober comes with you the whole way.
            </p>
          </div>
        </div>
      </section>

      <!-- Features -->
      <section id="features" class="border-t border-slate-900 px-6 py-10">
        <div class="mx-auto max-w-6xl">
          <div class="mb-8 text-center sm:text-left">
            <h2 class="text-xl font-bold text-white">What you get in the workspace</h2>
            <p class="mt-1 text-sm text-slate-400">
              Everything below is live today: scan-first flow maps with AI layered on when you need it.
            </p>
          </div>
          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <article
              v-for="feature in features"
              :key="feature.title"
              class="rounded-lg border border-slate-800 bg-slate-900/80 p-4 transition hover:border-slate-700"
            >
              <h3 class="mb-1.5 text-sm font-semibold text-white">{{ feature.title }}</h3>
              <p class="text-xs leading-relaxed text-slate-400">{{ feature.body }}</p>
            </article>
          </div>
        </div>
      </section>

      <!-- How it works -->
      <section id="how-it-works" class="border-t border-slate-900 px-6 py-10">
        <div class="mx-auto max-w-6xl">
          <div class="mb-8 flex items-end justify-center gap-3 sm:justify-start">
            <img :src="logo" alt="Bober the Bober" class="h-9 w-auto opacity-80" />
            <div class="text-center sm:text-left">
              <h2 class="text-xl font-bold text-white">How it works</h2>
              <p class="text-sm text-slate-400">Four steps with Bober from first open to understanding a function.</p>
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
      <div class="mt-auto border-t border-slate-900 px-6 py-3">
        <TeamCreditsBar variant="splash" />
      </div>
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
