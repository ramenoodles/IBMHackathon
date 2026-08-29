import { createRouter, createWebHistory } from 'vue-router'
import { isUserContextComplete } from '@/store/userContext'

/**
 * OnBober application router.
 * Defines routes for splash, onboarding, and workspace views.
 */
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'splash',
      component: () => import('@/views/SplashView.vue'),
      meta: { title: 'OnBober' },
    },
    {
      path: '/onboarding',
      name: 'onboarding',
      component: () => import('@/views/OnboardingView.vue'),
      meta: { title: 'Onboarding' },
    },
    {
      path: '/workspace',
      name: 'workspace',
      component: () => import('@/views/WorkspaceView.vue'),
      meta: { title: 'Workspace', requiresContext: true },
    },
  ],
})

/**
 * Navigation guard ensuring workspace access only after onboarding is complete.
 */
router.beforeEach((to) => {
  if (to.meta.requiresContext && !isUserContextComplete()) {
    return { name: 'onboarding' }
  }
  return true
})

export default router
