<script setup lang="ts">
/**
 * Shared beaver log-chewing SVG bar with optional header/footer chrome.
 */
import { computed, toRef } from 'vue'
import logo from '@/assets/logo.png'
import {
  useBeaverLogAnimation,
  type BeaverLogAnimationMode,
} from '@/composables/useBeaverLogAnimation'

export interface BeaverLogStage {
  threshold: number
  text: string
  badge: string
}

const props = withDefaults(
  defineProps<{
    mode: BeaverLogAnimationMode
    active: boolean
    externalProgress?: number
    durationMs?: number
    stages?: readonly BeaverLogStage[]
    title?: string
    subtitle?: string
    compact?: boolean
    barOnly?: boolean
  }>(),
  {
    externalProgress: 0,
    durationMs: 7500,
    stages: () => [],
    title: '',
    subtitle: '',
    compact: false,
    barOnly: false,
  },
)

const uid = `beaver-${Math.random().toString(36).slice(2, 9)}`

const modeRef = toRef(props, 'mode')
const activeRef = toRef(props, 'active')
const externalProgressRef = toRef(props, 'externalProgress')

const { progress, isComplete, beaverRig, beaverBobbing, intactClipRect, gnawedClipRect, activeNotch, chipsLayer, shavingsGroup } =
  useBeaverLogAnimation({
    mode: modeRef,
    active: activeRef,
    externalProgress: externalProgressRef,
    durationMs: props.durationMs,
  })

const displayPercent = computed(() => `${Math.floor(progress.value)}%`)

const currentStage = computed(() => {
  const stages = props.stages
  if (!stages.length) return null
  for (let i = stages.length - 1; i >= 0; i--) {
    if (progress.value >= stages[i]!.threshold) return stages[i]!
  }
  return stages[0]!
})

const badgeLabel = computed(() => currentStage.value?.badge ?? '')
const statusText = computed(() => currentStage.value?.text ?? '')
</script>

<template>
  <div class="beaver-log-bar" :class="{ compact, 'bar-only': barOnly }">
    <header v-if="!barOnly && (title || badgeLabel)" class="loader-header">
      <div v-if="badgeLabel" class="status-badge">
        <span class="status-dot" :class="{ complete: isComplete }" />
        <span>{{ badgeLabel }}</span>
      </div>
      <h2 v-if="title" class="title-text" :class="{ compact }">{{ title }}</h2>
      <p v-if="subtitle" class="source-label">{{ subtitle }}</p>
    </header>

    <div class="animation-stage" :class="{ compact: compact || barOnly }">
      <svg class="main-svg" viewBox="0 0 1200 700" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMidYMid meet">
        <defs>
          <linearGradient :id="`${uid}-barkGrad`" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#734120" />
            <stop offset="12%" stop-color="#8a502c" />
            <stop offset="35%" stop-color="#552c13" />
            <stop offset="75%" stop-color="#3d1e0c" />
            <stop offset="100%" stop-color="#241105" />
          </linearGradient>
          <linearGradient :id="`${uid}-gnawedGrad`" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#faeed7" />
            <stop offset="25%" stop-color="#e6c49c" />
            <stop offset="70%" stop-color="#cca06f" />
            <stop offset="100%" stop-color="#a47543" />
          </linearGradient>
          <linearGradient :id="`${uid}-biteWoodGrad`" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stop-color="#ffe8cc" />
            <stop offset="100%" stop-color="#d4a373" />
          </linearGradient>
          <filter :id="`${uid}-logShadow`" x="-5%" y="-30%" width="110%" height="180%">
            <feDropShadow dx="0" dy="16" stdDeviation="10" flood-color="#251307" flood-opacity="0.35" />
          </filter>
          <clipPath :id="`${uid}-intactClip`">
            <rect ref="intactClipRect" x="220" y="340" width="760" height="80" />
          </clipPath>
          <clipPath :id="`${uid}-gnawedClip`">
            <rect ref="gnawedClipRect" x="215" y="340" width="0" height="80" />
          </clipPath>
          <pattern :id="`${uid}-barkTexture`" width="80" height="48" patternUnits="userSpaceOnUse">
            <path d="M 0 10 Q 20 12 40 8 T 80 11" stroke="#3d1d0a" stroke-width="1.2" fill="none" opacity="0.4" stroke-linecap="round" />
            <path d="M 10 24 Q 30 20 50 26 T 90 22" stroke="#341706" stroke-width="1.4" fill="none" opacity="0.45" stroke-linecap="round" />
            <path d="M 0 38 Q 25 41 45 35 T 80 39" stroke="#250e03" stroke-width="1.2" fill="none" opacity="0.5" stroke-linecap="round" />
          </pattern>
        </defs>

        <g opacity="0.4">
          <line x1="180" y1="416" x2="1020" y2="416" stroke="#334155" stroke-width="2" stroke-dasharray="4 6" />
        </g>

        <g ref="shavingsGroup" />

        <g :filter="`url(#${uid}-logShadow)`">
          <g :clip-path="`url(#${uid}-gnawedClip)`">
            <path
              d="M 220 372 L 240 370 L 260 374 L 280 369 L 300 373 L 320 368 L 340 374 L 360 370 L 380 373 L 400 368 L 420 374 L 440 369 L 460 373 L 480 369 L 500 374 L 520 368 L 540 374 L 560 369 L 580 373 L 600 368 L 620 374 L 640 369 L 660 373 L 680 369 L 700 374 L 720 368 L 740 374 L 760 369 L 780 373 L 800 369 L 820 374 L 840 368 L 860 374 L 880 369 L 900 373 L 920 369 L 940 374 L 960 369 L 980 372 L 980 388 L 960 391 L 940 386 L 920 391 L 900 387 L 880 392 L 860 386 L 840 391 L 820 386 L 800 391 L 780 387 L 760 392 L 740 386 L 720 391 L 700 387 L 680 392 L 660 386 L 640 391 L 620 387 L 600 392 L 580 386 L 560 391 L 540 387 L 520 392 L 500 386 L 480 391 L 460 387 L 440 392 L 420 386 L 400 391 L 380 387 L 360 392 L 340 386 L 320 391 L 300 387 L 280 392 L 260 386 L 240 391 L 220 388 Z"
              :fill="`url(#${uid}-gnawedGrad)`"
              stroke="#8d562c"
              stroke-width="1.2"
            />
            <g stroke="#9c6436" stroke-width="1.2" opacity="0.6" stroke-linecap="round">
              <path d="M 235 371 L 238 387 M 275 370 L 278 386 M 315 370 L 318 387 M 355 371 L 358 386 M 395 370 L 398 387 M 435 370 L 438 386 M 475 370 L 478 387 M 515 370 L 518 386 M 555 370 L 558 387 M 595 370 L 598 386 M 635 371 L 638 387 M 675 370 L 678 386 M 715 370 L 718 387 M 755 370 L 758 386 M 795 370 L 798 387 M 835 370 L 838 386 M 875 370 L 878 387 M 915 370 L 918 386 M 955 370 L 958 387" />
            </g>
            <path d="M 220 375 L 980 375" stroke="#fff" stroke-width="1" opacity="0.35" stroke-dasharray="8 6" />
          </g>

          <g :clip-path="`url(#${uid}-intactClip)`">
            <rect x="220" y="356" width="760" height="48" rx="4" :fill="`url(#${uid}-barkGrad)`" />
            <rect x="220" y="356" width="760" height="48" :fill="`url(#${uid}-barkTexture)`" opacity="0.8" />
            <line x1="222" y1="358" x2="978" y2="358" stroke="#a25b30" stroke-width="2" stroke-linecap="round" opacity="0.9" />
            <line x1="222" y1="402" x2="978" y2="402" stroke="#180a02" stroke-width="2.5" stroke-linecap="round" />
            <g opacity="0.75">
              <ellipse cx="430" cy="380" rx="4" ry="7" fill="#2d1305" />
              <ellipse cx="690" cy="374" rx="3.5" ry="6" fill="#2d1305" />
              <ellipse cx="880" cy="384" rx="4" ry="7" fill="#2d1305" />
            </g>
          </g>

          <g>
            <rect x="218" y="356" width="4" height="48" rx="2" fill="#8d562c" />
            <line x1="220" y1="356" x2="220" y2="404" stroke="#e0be92" stroke-width="2" stroke-linecap="round" opacity="0.8" />
          </g>
          <g>
            <rect x="978" y="356" width="4" height="48" rx="2" fill="#8d562c" />
            <line x1="980" y1="356" x2="980" y2="404" stroke="#e0be92" stroke-width="2" stroke-linecap="round" opacity="0.8" />
          </g>

          <g ref="activeNotch" transform="translate(220, 380)">
            <path d="M -16 -24 L 2 -8 L 8 0 L 2 8 L -16 24 Z" :fill="`url(#${uid}-biteWoodGrad)`" stroke="#8d562c" stroke-width="1.2" />
            <polygon points="-8,-14 0,-11 -5,-8" fill="#fff5ea" />
            <polygon points="-2,-4 6,0 -2,4" fill="#fff5ea" />
            <polygon points="-8,8 0,11 -5,14" fill="#fff5ea" />
          </g>
        </g>

        <g ref="chipsLayer" />

        <g ref="beaverRig" transform="translate(170, 252)">
          <g ref="beaverBobbing">
            <image :href="logo" x="0" y="0" width="135" height="135" preserveAspectRatio="xMidYMid meet" />
            <g transform="translate(42, 108)">
              <circle cx="0" cy="0" r="2.2" fill="#f59e0b" opacity="0.9" />
              <circle cx="12" cy="4" r="1.6" fill="#fbbf24" opacity="0.8" />
            </g>
          </g>
        </g>

        <g v-if="!compact || barOnly" opacity="0.6">
          <g transform="translate(410, 424)">
            <line x1="0" y1="0" x2="0" y2="6" stroke="#64748b" stroke-width="1.5" stroke-linecap="round" />
            <text x="0" y="18" font-size="11" font-weight="700" fill="#64748b" text-anchor="middle" font-family="ui-monospace, monospace">25%</text>
          </g>
          <g transform="translate(600, 424)">
            <line x1="0" y1="0" x2="0" y2="6" stroke="#64748b" stroke-width="1.5" stroke-linecap="round" />
            <text x="0" y="18" font-size="11" font-weight="700" fill="#64748b" text-anchor="middle" font-family="ui-monospace, monospace">50%</text>
          </g>
          <g transform="translate(790, 424)">
            <line x1="0" y1="0" x2="0" y2="6" stroke="#64748b" stroke-width="1.5" stroke-linecap="round" />
            <text x="0" y="18" font-size="11" font-weight="700" fill="#64748b" text-anchor="middle" font-family="ui-monospace, monospace">75%</text>
          </g>
          <g transform="translate(980, 424)">
            <line x1="0" y1="0" x2="0" y2="6" stroke="#10b981" stroke-width="2" stroke-linecap="round" />
            <text x="0" y="18" font-size="11" font-weight="800" fill="#10b981" text-anchor="middle" font-family="ui-monospace, monospace">100%</text>
          </g>
        </g>
      </svg>
    </div>

    <footer v-if="!barOnly && statusText" class="footer-panel" :class="{ compact }">
      <div class="activity-info">
        <span class="activity-label">Current Progress</span>
        <span class="activity-text">{{ statusText }}</span>
      </div>
      <div class="percent-display" :class="{ complete: isComplete }">{{ displayPercent }}</div>
    </footer>

    <slot v-if="!barOnly" />
  </div>
</template>

<style scoped>
.beaver-log-bar {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  width: 100%;
}

.beaver-log-bar.bar-only {
  gap: 0;
}

.loader-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  text-align: center;
}

.animation-stage {
  width: 100%;
  aspect-ratio: 12 / 5;
  min-height: 180px;
}

.animation-stage.compact {
  aspect-ratio: 12 / 4;
  min-height: 120px;
}

.main-svg {
  width: 100%;
  height: 100%;
  display: block;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 5px 14px;
  background: rgba(15, 23, 42, 0.85);
  border: 1px solid #334155;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #94a3b8;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #f59e0b;
  box-shadow: 0 0 8px #f59e0b;
  animation: pulseDot 1.4s infinite ease-in-out;
}

.status-dot.complete {
  background-color: #10b981;
  box-shadow: 0 0 10px #10b981;
}

@keyframes pulseDot {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.35); opacity: 0.6; }
}

.title-text {
  font-size: 1.5rem;
  font-weight: 800;
  color: #f8fafc;
  letter-spacing: -0.02em;
}

.title-text.compact {
  font-size: 1.125rem;
}

.source-label {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.875rem;
  color: #64748b;
}

.footer-panel {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(8px);
  border: 1px solid #334155;
  padding: 14px 20px;
  border-radius: 14px;
}

.footer-panel.compact {
  padding: 10px 14px;
  border-radius: 10px;
}

.activity-info {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.activity-label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #64748b;
}

.activity-text {
  font-size: 13px;
  font-weight: 600;
  color: #e2e8f0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.footer-panel.compact .activity-text {
  font-size: 11px;
}

.percent-display {
  font-size: 1.875rem;
  font-weight: 900;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: #f59e0b;
  letter-spacing: -0.03em;
  flex-shrink: 0;
}

.footer-panel.compact .percent-display {
  font-size: 1.5rem;
}

.percent-display.complete {
  color: #10b981;
}
</style>
