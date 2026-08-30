import type { ExperienceLevel } from '@/store/userContext'

export interface ExperienceLevelOption {
  value: ExperienceLevel
  label: string
  shortLabel: string
  description: string
}

export const EXPERIENCE_LEVELS: readonly ExperienceLevelOption[] = [
  {
    value: 'junior',
    label: 'Junior SWE',
    shortLabel: 'Junior',
    description: 'Plain language with jargon defined — longer, guided step explanations.',
  },
  {
    value: 'mid',
    label: 'Mid-level SWE',
    shortLabel: 'Mid',
    description: 'Balanced tone — assumes fundamentals, focuses on this codebase’s intent.',
  },
  {
    value: 'senior',
    label: 'Senior SWE',
    shortLabel: 'Senior',
    description: 'Concise and direct — intent, edge cases, minimal hand-holding.',
  },
] as const

export const EXPERIENCE_EFFECT_SUMMARY =
  'Your level adjusts how AI labels flow steps and how “Explain this step” answers are written.'

export const EXPERIENCE_EFFECT_BULLETS = [
  'Step titles and summaries in the flow map',
  'Detail panel explanations when you click “Explain this step”',
] as const

export function experienceLevelLabel(level: ExperienceLevel): string {
  return EXPERIENCE_LEVELS.find((o) => o.value === level)?.label ?? level
}

export function experienceLevelShortLabel(level: ExperienceLevel): string {
  return EXPERIENCE_LEVELS.find((o) => o.value === level)?.shortLabel ?? level
}
