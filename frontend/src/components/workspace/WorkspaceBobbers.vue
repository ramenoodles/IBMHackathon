<script setup lang="ts">
/**
 * Subtle decorative bobers tucked into workspace corners — pointer-events-none.
 */
import logo from '@/assets/logo.png'
</script>

<template>
  <div class="pointer-events-none absolute inset-0 z-0 overflow-hidden" aria-hidden="true">
    <img
      :src="logo"
      alt=""
      class="bobber-peek absolute -bottom-2 left-3 h-9 w-auto opacity-[0.08]"
    />
    <img
      :src="logo"
      alt=""
      class="bobber-float absolute bottom-[18%] left-[12%] h-7 w-auto opacity-[0.05] sm:left-[10%]"
    />
    <img
      :src="logo"
      alt=""
      class="bobber-bob absolute right-[22%] top-14 h-6 w-auto opacity-[0.06] sm:top-12"
    />
    <img
      :src="logo"
      alt=""
      class="bobber-float-delayed absolute bottom-[22%] right-[8%] h-10 w-auto opacity-[0.05]"
    />
    <img
      :src="logo"
      alt=""
      class="bobber-peek absolute -right-1 top-[38%] h-8 w-auto rotate-12 opacity-[0.04]"
    />
    <img
      :src="logo"
      alt=""
      class="bobber-bob absolute bottom-6 left-[42%] h-5 w-auto opacity-[0.04] max-sm:hidden"
    />
  </div>
</template>

<style scoped>
.bobber-float {
  animation: bobber-drift 10s ease-in-out infinite;
}

.bobber-float-delayed {
  animation: bobber-drift-reverse 12s ease-in-out infinite;
}

.bobber-bob {
  animation: bobber-bob 3.5s ease-in-out infinite;
}

.bobber-peek {
  animation: bobber-peek 4.5s ease-in-out infinite;
}

@keyframes bobber-drift {
  0%, 100% { transform: translateY(0) rotate(-10deg); }
  50% { transform: translateY(-8px) rotate(-6deg); }
}

@keyframes bobber-drift-reverse {
  0%, 100% { transform: translateY(0) rotate(10deg); }
  50% { transform: translateY(-6px) rotate(6deg); }
}

@keyframes bobber-bob {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

@keyframes bobber-peek {
  0%, 100% { transform: translateY(0) rotate(-10deg); }
  50% { transform: translateY(2px) rotate(-6deg); }
}
</style>
