import { describe, expect, it } from 'vitest'
import { scoreFileMatch, searchFiles } from '@/composables/useFileIndex'
import type { TreeEntry } from '@/composables/useFileTree'

const files: TreeEntry[] = [
  { name: 'main.go', path: 'cmd/main.go', isDir: false },
  { name: 'handler.go', path: 'internal/httpapi/handler.go', isDir: false },
  { name: 'README.md', path: 'README.md', isDir: false },
  { name: 'main.ts', path: 'frontend/src/main.ts', isDir: false },
  { name: 'Sidebar.vue', path: 'frontend/src/components/workspace/Sidebar.vue', isDir: false },
]

describe('scoreFileMatch', () => {
  it('prefers exact name match', () => {
    expect(scoreFileMatch(files[0]!, 'main.go')).toBe(0)
  })

  it('matches starts-with on name', () => {
    expect(scoreFileMatch(files[0]!, 'main')).toBe(1)
  })

  it('matches contains in name', () => {
    expect(scoreFileMatch(files[1]!, 'dler')).toBe(2)
  })

  it('matches contains in path', () => {
    expect(scoreFileMatch(files[4]!, 'workspace')).toBe(3)
  })

  it('returns null when no match', () => {
    expect(scoreFileMatch(files[0]!, 'zzz')).toBeNull()
  })
})

describe('searchFiles', () => {
  it('ranks exact name above partial matches', () => {
    const results = searchFiles(files, 'main')
    expect(results[0]?.name).toBe('main.go')
    expect(results.some((r) => r.name === 'main.ts')).toBe(true)
  })

  it('returns empty for blank query', () => {
    expect(searchFiles(files, '')).toEqual([])
    expect(searchFiles(files, '   ')).toEqual([])
  })

  it('respects limit', () => {
    expect(searchFiles(files, 'main', 1)).toHaveLength(1)
  })

  it('skips directories', () => {
    const withDir: TreeEntry[] = [
      ...files,
      { name: 'internal', path: 'internal', isDir: true },
    ]
    const results = searchFiles(withDir, 'internal')
    expect(results.every((r) => !r.isDir)).toBe(true)
  })
})
