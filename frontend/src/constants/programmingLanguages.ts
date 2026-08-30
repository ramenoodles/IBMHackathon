/**
 * Programming language options for onboarding and workspace settings.
 * `value` must match a backend language profile name or alias.
 */
export interface ProgrammingLanguageOption {
  label: string
  value: string
}

export const PROGRAMMING_LANGUAGES: readonly ProgrammingLanguageOption[] = [
  { label: 'Python', value: 'python' },
  { label: 'JavaScript', value: 'javascript' },
  { label: 'TypeScript', value: 'typescript' },
  { label: 'Go', value: 'go' },
  { label: 'Rust', value: 'rust' },
  { label: 'Java', value: 'java' },
  { label: 'C', value: 'c' },
  { label: 'C++', value: 'cpp' },
  { label: 'C#', value: 'csharp' },
  { label: 'Ruby', value: 'auto' },
  { label: 'PHP', value: 'auto' },
  { label: 'Swift', value: 'auto' },
  { label: 'Kotlin', value: 'auto' },
  { label: 'Scala', value: 'auto' },
  { label: 'Haskell', value: 'auto' },
  { label: 'Lua', value: 'auto' },
  { label: 'Elixir', value: 'auto' },
  { label: 'Clojure', value: 'auto' },
  { label: 'Dart', value: 'auto' },
  { label: 'R', value: 'auto' },
  { label: 'MATLAB', value: 'auto' },
  { label: 'Shell/Bash', value: 'auto' },
  { label: 'Other', value: 'auto' },
] as const

/** Map stored primaryLanguage (comma-separated values) to display labels. */
export function languageLabelsFromStored(stored: string): Set<string> {
  return new Set(
    stored
      .split(',')
      .map((s) => s.trim())
      .map((s) => PROGRAMMING_LANGUAGES.find((l) => l.value === s || l.label === s)?.label ?? '')
      .filter(Boolean),
  )
}

/** Serialize selected labels to comma-separated backend-safe values. */
export function storedLanguagesFromLabels(labels: Iterable<string>): string {
  const values = [...labels].map(
    (label) => PROGRAMMING_LANGUAGES.find((l) => l.label === label)?.value ?? 'auto',
  )
  return [...new Set(values)].join(',')
}
