<script setup lang="ts">
/**
 * Landing page for OnBober with branding and workspace initialization CTA.
 */
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import logo from '@/assets/logo.svg'

const router = useRouter()

/** Navigate to the onboarding flow. */
function initializeWorkspace(): void {
  router.push('/onboarding')
}
</script>

<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-slate-950 px-6">
    <img :src="logo" alt="OnBober logo" class="mb-8 h-24 w-24" />
    <h1 class="mb-3 text-5xl font-bold tracking-tight text-white">
      On<span class="text-onbober-primary">Bober</span>
    </h1>
    <p class="mb-10 max-w-md text-center text-lg text-slate-400">
      Walk the execution path of any function. Expand branches on demand and onboard into complex codebases fast.
    </p>
    <Button variant="primary" @click="initializeWorkspace">Initialize Workspace</Button>
  </div>
</template>
