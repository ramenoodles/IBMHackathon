import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'
import { dirname, resolve } from 'node:path'

const dir = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  resolve: {
    alias: {
      '@': resolve(dir, './src'),
    },
  },
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts'],
  },
})
