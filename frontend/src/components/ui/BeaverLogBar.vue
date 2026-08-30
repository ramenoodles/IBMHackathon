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
            <stop offset="0%" stop-color="#784421" />
            <stop offset="15%" stop-color="#8f5228" />
            <stop offset="40%" stop-color="#5a3015" />
            <stop offset="78%" stop-color="#3d1d0a" />
            <stop offset="100%" stop-color="#241004" />
          </linearGradient>
          <linearGradient :id="`${uid}-gnawedGrad`" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#fdf3e2" />
            <stop offset="28%" stop-color="#ebd0a9" />
            <stop offset="70%" stop-color="#d4aa79" />
            <stop offset="100%" stop-color="#ad7c48" />
          </linearGradient>
          <linearGradient :id="`${uid}-biteWoodGrad`" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stop-color="#fff0dc" />
            <stop offset="100%" stop-color="#d8a873" />
          </linearGradient>
          <filter :id="`${uid}-softShadow`" x="-10%" y="-40%" width="120%" height="200%">
            <feGaussianBlur stdDeviation="8" result="blur" />
            <feColorMatrix type="matrix" values="0 0 0 0 0.18  0 0 0 0 0.1  0 0 0 0 0.05  0 0 0 0.16 0" />
          </filter>
          <clipPath :id="`${uid}-intactClip`">
            <rect ref="intactClipRect" x="180" y="280" width="840" height="140" />
          </clipPath>
          <clipPath :id="`${uid}-gnawedClip`">
            <rect ref="gnawedClipRect" x="175" y="280" width="0" height="140" />
          </clipPath>
          <pattern :id="`${uid}-barkTexture`" width="90" height="60" patternUnits="userSpaceOnUse">
            <path d="M 0 12 Q 22 15 45 10 T 90 13" stroke="#3d1d0a" stroke-width="1.6" fill="none" opacity="0.4" stroke-linecap="round" />
            <path d="M 12 30 Q 35 25 58 32 T 102 28" stroke="#311403" stroke-width="1.8" fill="none" opacity="0.45" stroke-linecap="round" />
            <path d="M 0 48 Q 28 52 52 44 T 90 49" stroke="#220b01" stroke-width="1.6" fill="none" opacity="0.5" stroke-linecap="round" />
          </pattern>
        </defs>

        <ellipse cx="600" cy="402" rx="440" ry="14" fill="#000000" opacity="0.06" :filter="`url(#${uid}-softShadow)`" />

        <g ref="shavingsGroup" />

        <g>
          <g :clip-path="`url(#${uid}-gnawedClip)`">
            <path
              d="M 180 339 L 205 336 L 230 341 L 255 336 L 280 340 L 305 335 L 330 341 L 355 336 L 380 340 L 405 335 L 430 341 L 455 336 L 480 340 L 505 335 L 530 341 L 555 336 L 580 340 L 605 335 L 630 341 L 655 336 L 680 340 L 705 335 L 730 341 L 755 336 L 780 340 L 805 335 L 830 341 L 855 336 L 880 340 L 905 335 L 930 341 L 955 336 L 980 340 L 1005 336 L 1020 339 L 1020 361 L 1005 364 L 980 359 L 955 365 L 930 360 L 905 365 L 880 359 L 855 364 L 830 359 L 805 365 L 780 360 L 755 365 L 730 359 L 705 364 L 680 359 L 655 365 L 630 360 L 605 365 L 580 359 L 555 364 L 530 359 L 505 365 L 480 360 L 455 365 L 430 359 L 405 364 L 380 359 L 355 365 L 330 360 L 305 365 L 280 359 L 255 364 L 230 359 L 205 364 L 180 361 Z"
              :fill="`url(#${uid}-gnawedGrad)`"
              stroke="#875127"
              stroke-width="1.5"
            />
            <g stroke="#9c6334" stroke-width="1.4" opacity="0.6" stroke-linecap="round">
              <path d="M 195 338 L 198 360 M 240 337 L 243 359 M 285 337 L 288 360 M 330 338 L 333 359 M 375 337 L 378 360 M 420 337 L 423 359 M 465 337 L 468 360 M 510 337 L 513 359 M 555 337 L 558 360 M 600 337 L 603 359 M 645 338 L 648 360 M 690 337 L 693 359 M 735 337 L 738 360 M 780 337 L 783 360 M 825 337 L 828 360 M 870 337 L 873 360 M 915 337 L 918 360 M 960 337 L 963 360 M 1005 337 L 1008 360" />
            </g>
            <line x1="180" y1="344" x2="1020" y2="344" stroke="#ffffff" stroke-width="1.2" opacity="0.45" stroke-dasharray="10 8" />
          </g>

          <g :clip-path="`url(#${uid}-intactClip)`">
            <rect x="180" y="320" width="840" height="60" rx="5" :fill="`url(#${uid}-barkGrad)`" />
            <rect x="180" y="320" width="840" height="60" rx="5" :fill="`url(#${uid}-barkTexture)`" opacity="0.85" />
            <line x1="182" y1="322" x2="1018" y2="322" stroke="#b06437" stroke-width="2.5" stroke-linecap="round" opacity="0.9" />
            <line x1="182" y1="378" x2="1018" y2="378" stroke="#170902" stroke-width="3" stroke-linecap="round" />
            <g opacity="0.7">
              <ellipse cx="380" cy="350" rx="5" ry="8" fill="#2b1104" />
              <ellipse cx="680" cy="344" rx="4" ry="7" fill="#2b1104" />
              <ellipse cx="890" cy="356" rx="5" ry="9" fill="#2b1104" />
            </g>
          </g>

          <g>
            <rect x="177" y="320" width="5" height="60" rx="2" fill="#875127" />
            <line x1="180" y1="320" x2="180" y2="380" stroke="#ebd0a9" stroke-width="2.2" stroke-linecap="round" opacity="0.85" />
          </g>
          <g>
            <rect x="1018" y="320" width="5" height="60" rx="2" fill="#875127" />
            <line x1="1020" y1="320" x2="1020" y2="380" stroke="#ebd0a9" stroke-width="2.2" stroke-linecap="round" opacity="0.85" />
          </g>

          <g ref="activeNotch" transform="translate(180, 350)">
            <path d="M -22 -30 L 2 -11 L 10 0 L 2 11 L -22 30 Z" :fill="`url(#${uid}-biteWoodGrad)`" stroke="#875127" stroke-width="1.5" />
            <polygon points="-12,-18 -1,-14 -7,-10" fill="#ffffff" />
            <polygon points="-3,-5 8,0 -3,5" fill="#ffffff" />
            <polygon points="-12,10 -1,14 -7,18" fill="#ffffff" />
          </g>
        </g>

        <g ref="chipsLayer" />

        <g ref="beaverRig" transform="translate(126, 212)">
          <g ref="beaverBobbing">
            <image :href="logo" x="0" y="0" width="145" height="145" preserveAspectRatio="xMidYMid meet" />
            <g transform="translate(46, 116)">
              <circle cx="0" cy="0" r="2.5" fill="#f59e0b" opacity="0.9" />
              <circle cx="14" cy="4" r="1.8" fill="#fbbf24" opacity="0.8" />
            </g>
          </g>
        </g>

        <g v-if="!compact && !barOnly" opacity="0.6">
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
