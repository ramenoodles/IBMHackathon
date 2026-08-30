<script setup lang="ts">
/**
 * Compact single-row team credits bar for workspace and splash footers.
 */
import logo from '@/assets/logo.png'
import WorkspaceHelpButton from '@/components/ui/WorkspaceHelpButton.vue'
import { SOURCE_REPO_LABEL, SOURCE_REPO_URL } from '@/constants/projectRepo'
import { TEAM_MEMBERS } from '@/constants/teamMembers'

withDefaults(
  defineProps<{
    variant?: 'compact' | 'splash'
  }>(),
  { variant: 'compact' },
)

function memberLink(member: (typeof TEAM_MEMBERS)[number]): string | undefined {
  return member.github ?? member.linkedin
}

function firstName(fullName: string): string {
  return fullName.split(' ')[0] ?? fullName
}
</script>

<template>
  <footer
    id="built-by"
    class="flex shrink-0 items-center gap-3 bg-slate-950/80 px-4 py-2"
    :class="variant === 'splash' ? 'mx-auto w-full max-w-6xl' : 'border-t border-slate-800/80'"
  >
    <div class="flex min-w-0 shrink-0 items-center gap-2">
      <img :src="logo" alt="" class="h-5 w-auto opacity-70" aria-hidden="true" />
      <span class="whitespace-nowrap text-xs font-medium uppercase tracking-wide text-slate-500">
        Built by
      </span>
    </div>

    <ul class="flex min-w-0 flex-1 items-center gap-0 overflow-x-auto whitespace-nowrap text-xs text-slate-400">
      <li
        v-for="(member, index) in TEAM_MEMBERS"
        :key="member.name"
        class="inline-flex items-center"
      >
        <span v-if="index > 0" class="mx-1.5 text-slate-600" aria-hidden="true">·</span>
        <a
          v-if="memberLink(member)"
          :href="memberLink(member)"
          target="_blank"
          rel="noopener noreferrer"
          class="transition hover:text-white"
          :title="member.name"
        >
          {{ firstName(member.name) }}
        </a>
        <span v-else>{{ firstName(member.name) }}</span>
      </li>
    </ul>

    <a
      :href="SOURCE_REPO_URL"
      target="_blank"
      rel="noopener noreferrer"
      class="hidden shrink-0 font-mono text-xs text-slate-500 transition hover:text-slate-300 sm:inline"
      :title="SOURCE_REPO_URL"
    >
      {{ SOURCE_REPO_LABEL }}
    </a>
    <WorkspaceHelpButton
      v-if="variant === 'splash'"
      variant="splash"
      placement="above"
    />
  </footer>
</template>
