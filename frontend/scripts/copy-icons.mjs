/**
 * Copies the vscode-material-icons SVG set into public/ so they're served
 * as static assets. Runs automatically after npm install (see postinstall)
 * since the copied icons aren't committed to the repo.
 */
import { cpSync, existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = dirname(dirname(fileURLToPath(import.meta.url)))
const src = join(root, 'node_modules/vscode-material-icons/generated/icons')
const dest = join(root, 'public/icons/vscode-material-icons')

if (!existsSync(src)) {
  console.error(`copy-icons: source not found at ${src}`)
  process.exit(1)
}

cpSync(src, dest, { recursive: true })
console.log(`copy-icons: copied vscode-material-icons -> public/icons/vscode-material-icons`)
