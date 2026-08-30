/** Default public demo repository for onboarding. */
export const DEMO_REPO_URL = 'https://github.com/IBM/sarama'

export const DEMO_REPO_LABEL = 'IBM/sarama'

/** Shorthand owner/repo for loading UI and labels. */
export function formatRepoDisplayLabel(input: string): string {
  const trimmed = input.trim()
  if (!trimmed) return ''
  const urlMatch = trimmed.match(/github\.com\/([^/?#]+\/[^/?#]+)/i)
  if (urlMatch) return urlMatch[1]!
  return trimmed.replace(/^https?:\/\//, '')
}
