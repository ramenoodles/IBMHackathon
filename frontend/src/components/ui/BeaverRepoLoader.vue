<script setup lang="ts">
/**
 * Beaver log-chewing loader for repo clone / archive setup.
 */
import { computed, onMounted, onUnmounted, ref, toRef, watch } from 'vue'
import logo from '@/assets/logo.png'
import { useRotatingPhrase } from '@/composables/useRotatingPhrase'

const props = withDefaults(
  defineProps<{
    active: boolean
    phrases: readonly string[]
    sourceLabel?: string
  }>(),
  { sourceLabel: '' },
)

const { currentPhrase } = useRotatingPhrase(toRef(props, 'active'), toRef(props, 'phrases'))

const uid = `beaver-${Math.random().toString(36).slice(2, 9)}`

const START_X = 220
const END_X = 980
const TOTAL_TRAVEL = END_X - START_X
const BEAVER_OFFSET_X = -50
const BEAVER_BASE_Y = 252
const MAX_SHAVINGS = 50
const CHIP_COUNT = 32
const chipColors = ['#faeed7', '#e6c49c', '#cca06f', '#8d562c', '#6a3d1c', '#ffe8cc']

const ACTIONS = [
  { threshold: 0, text: 'MEASURING LOG PROFILE...', badge: 'STAGE 1: TIMBER SCAN' },
  { threshold: 20, text: 'CHISELING OUTER BARK...', badge: 'STAGE 2: GNAWING' },
  { threshold: 50, text: 'RAPID CORE PROCESSING...', badge: 'STAGE 3: FAST CHEWING' },
  { threshold: 75, text: 'TRIMMING WOOD BEAM...', badge: 'STAGE 4: SHAPING' },
  { threshold: 100, text: 'TIMBER COMPLETE! READY!', badge: 'FINISHED' },
] as const

const progress = ref(0)
const actionText = ref<string>(ACTIONS[0].text)
const badgeLabel = ref<string>(ACTIONS[0].badge)
const isComplete = ref(false)

const beaverRig = ref<SVGGElement | null>(null)
const beaverBobbing = ref<SVGGElement | null>(null)
const intactClipRect = ref<SVGRectElement | null>(null)
const gnawedClipRect = ref<SVGRectElement | null>(null)
const activeNotch = ref<SVGGElement | null>(null)
const chipsLayer = ref<SVGGElement | null>(null)
const shavingsGroup = ref<SVGGElement | null>(null)

interface Chip {
  el: SVGPolygonElement
  active: boolean
  x: number
  y: number
  vx: number
  vy: number
  rot: number
  vRot: number
  scale: number
  life: number
  maxLife: number
}

interface Shaving {
  el: SVGEllipseElement
  x: number
  y: number
  rot: number
  revealed: boolean
}

const chipPool: Chip[] = []
const shavings: Shaving[] = []

let mainLoopId: number | null = null
let lastTime = performance.now()
let spawnCounter = 0
let finishing = false

const displayPercent = computed(() => `${Math.floor(progress.value)}%`)

function initParticles(): void {
  if (!chipsLayer.value || !shavingsGroup.value) return

  for (let i = 0; i < CHIP_COUNT; i++) {
    const p = document.createElementNS('http://www.w3.org/2000/svg', 'polygon')
    p.setAttribute('points', i % 2 === 0 ? '-4,-2 4,-4 2,3 -3,3' : '-3,-3 4,-1 1,4 -3,2')
    p.setAttribute('fill', chipColors[i % chipColors.length]!)
    p.setAttribute('opacity', '0')
    chipsLayer.value.appendChild(p)
    chipPool.push({
      el: p,
      active: false,
      x: 0,
      y: 0,
      vx: 0,
      vy: 0,
      rot: 0,
      vRot: 0,
      scale: 1,
      life: 0,
      maxLife: 1,
    })
  }

  for (let i = 0; i < MAX_SHAVINGS; i++) {
    const sh = document.createElementNS('http://www.w3.org/2000/svg', 'ellipse')
    sh.setAttribute('rx', (Math.random() * 3.5 + 2).toFixed(1))
    sh.setAttribute('ry', (Math.random() * 1.8 + 0.8).toFixed(1))
    sh.setAttribute('fill', chipColors[Math.floor(Math.random() * chipColors.length)]!)
    sh.setAttribute('opacity', '0')
    shavingsGroup.value.appendChild(sh)
    shavings.push({
      el: sh,
      x: 0,
      y: 416 + Math.random() * 8,
      rot: Math.random() * 360,
      revealed: false,
    })
  }
}

function spawnWoodChip(originX: number, originY: number): void {
  const chip = chipPool.find((c) => !c.active)
  if (!chip) return
  chip.active = true
  chip.x = originX + (Math.random() * 8 - 4)
  chip.y = originY + (Math.random() * 14 - 7)
  chip.vx = -(Math.random() * 4 + 2)
  chip.vy = Math.random() * 5.5 - 3.2
  chip.rot = Math.random() * 360
  chip.vRot = Math.random() * 24 - 12
  chip.scale = Math.random() * 0.6 + 0.65
  chip.life = 0
  chip.maxLife = Math.random() * 0.4 + 0.35
  chip.el.setAttribute('opacity', '1')
}

function resetShavings(): void {
  for (const sh of shavings) {
    sh.revealed = false
    sh.el.setAttribute('opacity', '0')
  }
}

function updateLabels(): void {
  for (let i = ACTIONS.length - 1; i >= 0; i--) {
    if (progress.value >= ACTIONS[i]!.threshold) {
      actionText.value = ACTIONS[i]!.text
      badgeLabel.value = ACTIONS[i]!.badge
      break
    }
  }
  isComplete.value = progress.value >= 100
}

function update(now: number): void {
  const dt = Math.min((now - lastTime) / 1000, 0.1)
  lastTime = now

  if (props.active && !finishing) {
    const target = 90
    const speed = 8 + Math.sin(progress.value * 0.08) * 2
    if (progress.value < target) {
      progress.value += speed * dt
      if (progress.value > target) progress.value = target
    }
  } else if (finishing || !props.active) {
    progress.value += 40 * dt
    if (progress.value > 100) progress.value = 100
  }

  const cutX = START_X + (progress.value / 100) * TOTAL_TRAVEL
  const chewing = progress.value < 100

  if (beaverRig.value && beaverBobbing.value) {
    const beaverX = cutX + BEAVER_OFFSET_X
    const chewBobY = chewing ? Math.sin(now * 0.038) * 3.2 : 0
    const chewBobRot = chewing ? Math.cos(now * 0.038) * 2.0 : 0
    const chewScaleX = chewing ? 1 + Math.sin(now * 0.076) * 0.02 : 1
    beaverRig.value.setAttribute('transform', `translate(${beaverX}, ${BEAVER_BASE_Y})`)
    beaverBobbing.value.setAttribute(
      'transform',
      `translate(0, ${chewBobY}) rotate(${chewBobRot} 65 75) scale(${chewScaleX} 1)`,
    )
  }

  if (intactClipRect.value) {
    intactClipRect.value.setAttribute('x', String(cutX))
    intactClipRect.value.setAttribute('width', String(Math.max(0, END_X - cutX + 10)))
  }
  if (gnawedClipRect.value) {
    gnawedClipRect.value.setAttribute('x', String(START_X - 5))
    gnawedClipRect.value.setAttribute('width', String(Math.max(0, cutX - START_X + 5)))
  }
  if (activeNotch.value) {
    activeNotch.value.setAttribute('transform', `translate(${cutX}, 380)`)
    activeNotch.value.setAttribute('opacity', chewing ? '1' : '0')
  }

  if (chewing) {
    spawnCounter += dt
    if (spawnCounter > 0.045) {
      spawnCounter = 0
      spawnWoodChip(cutX - 10, 380)
      spawnWoodChip(cutX - 4, 376)
    }
  }

  for (const chip of chipPool) {
    if (!chip.active) continue
    chip.life += dt
    if (chip.life >= chip.maxLife) {
      chip.active = false
      chip.el.setAttribute('opacity', '0')
      continue
    }
    chip.vy += 22 * dt
    chip.x += chip.vx
    chip.y += chip.vy
    chip.rot += chip.vRot
    const alpha = 1 - chip.life / chip.maxLife
    chip.el.setAttribute(
      'transform',
      `translate(${chip.x.toFixed(1)}, ${chip.y.toFixed(1)}) rotate(${chip.rot.toFixed(1)}) scale(${chip.scale})`,
    )
    chip.el.setAttribute('opacity', Math.max(0, alpha).toFixed(2))
  }

  const revealCount = Math.floor((progress.value / 100) * MAX_SHAVINGS)
  for (let i = 0; i < MAX_SHAVINGS; i++) {
    const sh = shavings[i]!
    if (i < revealCount && !sh.revealed) {
      sh.revealed = true
      sh.x = START_X + (i / MAX_SHAVINGS) * TOTAL_TRAVEL + (Math.random() * 20 - 10)
      sh.el.setAttribute('cx', sh.x.toFixed(1))
      sh.el.setAttribute('cy', sh.y.toFixed(1))
      sh.el.setAttribute('transform', `rotate(${sh.rot.toFixed(1)} ${sh.x.toFixed(1)} ${sh.y.toFixed(1)})`)
      sh.el.setAttribute('opacity', '0.75')
    }
  }

  updateLabels()
  mainLoopId = requestAnimationFrame(update)
}

function startLoop(): void {
  if (mainLoopId !== null) return
  lastTime = performance.now()
  mainLoopId = requestAnimationFrame(update)
}

function stopLoop(): void {
  if (mainLoopId !== null) {
    cancelAnimationFrame(mainLoopId)
    mainLoopId = null
  }
}

watch(
  () => props.active,
  (isActive) => {
    if (!isActive && progress.value < 100) {
      finishing = true
    }
    if (isActive) {
      finishing = false
      progress.value = 0
      resetShavings()
      chipPool.forEach((chip) => {
        chip.active = false
        chip.el.setAttribute('opacity', '0')
      })
      startLoop()
    }
  },
  { immediate: true },
)

onMounted(() => {
  initParticles()
  startLoop()
})

onUnmounted(() => {
  stopLoop()
})
</script>

<template>
  <div class="beaver-loader">
    <div class="ui-layer">
      <div class="header-box">
        <div class="status-badge">
          <span class="status-dot" :class="{ complete: isComplete }" />
          <span>{{ badgeLabel }}</span>
        </div>
        <h2 class="title-text">Setting up your workspace</h2>
        <p v-if="sourceLabel" class="source-label">{{ sourceLabel }}</p>
      </div>

      <div class="footer-panel">
        <div class="activity-info">
          <span class="activity-label">Current Progress</span>
          <span class="activity-text">{{ actionText }}</span>
        </div>
        <div class="percent-display" :class="{ complete: isComplete }">{{ displayPercent }}</div>
      </div>
    </div>

    <svg class="main-svg" viewBox="0 0 1200 700" xmlns="http://www.w3.org/2000/svg">
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

      <g opacity="0.6">
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

    <Transition name="phrase-fade" mode="out-in">
      <p :key="currentPhrase" class="rotating-phrase">{{ currentPhrase }}</p>
    </Transition>
  </div>
</template>

<style scoped>
.beaver-loader {
  position: relative;
  width: 100%;
  max-width: 400px;
  margin: 0 auto;
  aspect-ratio: 12 / 7;
  background: radial-gradient(circle at 50% 45%, #1e293b 0%, #0f172a 100%);
  border-radius: 12px;
  overflow: hidden;
}

.main-svg {
  width: 100%;
  height: 100%;
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.ui-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px 10px;
  z-index: 10;
}

.header-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  background: rgba(15, 23, 42, 0.85);
  border: 1px solid #334155;
  border-radius: 9999px;
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #94a3b8;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #f59e0b;
  box-shadow: 0 0 6px #f59e0b;
  animation: pulseDot 1.4s infinite ease-in-out;
}

.status-dot.complete {
  background-color: #10b981;
  box-shadow: 0 0 8px #10b981;
}

@keyframes pulseDot {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.35); opacity: 0.6; }
}

.title-text {
  font-size: 14px;
  font-weight: 700;
  color: #f1f5f9;
  letter-spacing: -0.02em;
}

.source-label {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, monospace;
  font-size: 10px;
  color: #64748b;
}

.footer-panel {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(8px);
  border: 1px solid #334155;
  padding: 8px 14px;
  border-radius: 10px;
}

.activity-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.activity-label {
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: #64748b;
}

.activity-text {
  font-size: 10px;
  font-weight: 600;
  color: #cbd5e1;
  font-family: ui-monospace, monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.percent-display {
  font-size: 20px;
  font-weight: 900;
  font-family: ui-monospace, monospace;
  color: #f59e0b;
  letter-spacing: -0.03em;
  flex-shrink: 0;
}

.percent-display.complete {
  color: #10b981;
}

.rotating-phrase {
  position: absolute;
  bottom: -28px;
  left: 0;
  right: 0;
  text-align: center;
  font-size: 13px;
  font-style: italic;
  color: #94a3b8;
}

.phrase-fade-enter-active,
.phrase-fade-leave-active {
  transition: opacity 0.3s ease;
}

.phrase-fade-enter-from,
.phrase-fade-leave-to {
  opacity: 0;
}
</style>
