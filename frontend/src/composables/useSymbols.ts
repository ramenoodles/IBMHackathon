/**
 * Extract function/class symbols only (not module-level assignments).
 * @param source - Raw file contents.
 * @param filePath - File path used to infer language.
 */
export function extractFunctions(source: string, filePath: string): string[] {
  const ext = filePath.split('.').pop()?.toLowerCase() ?? ''
  const patterns: RegExp[] = []
  const seen = new Set<string>()
  const symbols: string[] = []

  const add = (name: string | undefined): void => {
    if (!name || seen.has(name) || isReserved(name)) return
    seen.add(name)
    symbols.push(name)
  }

  if (ext === 'py') {
    patterns.push(/^\s*(?:async\s+)?def\s+([A-Za-z_]\w*)/gm)
    patterns.push(/^\s*class\s+([A-Za-z_]\w*)/gm)
  } else if (ext === 'go') {
    patterns.push(/^\s*func\s+(?:\([^)]*\)\s+)?([A-Za-z_]\w*)/gm)
  } else if (ext === 'rs') {
    patterns.push(/^\s*(?:pub\s+)?fn\s+([A-Za-z_]\w*)/gm)
  } else {
    patterns.push(/^\s*(?:static\s+|inline\s+)?[\w\s\*]+\s+([A-Za-z_]\w*)\s*\(/gm)
    patterns.push(/^\s*class\s+([A-Za-z_]\w*)/gm)
  }

  for (const pattern of patterns) {
    for (const match of source.matchAll(pattern)) {
      add(match[1])
    }
  }
  return symbols
}

/**
 * Extract all clickable symbol names including module-level names.
 */
export function extractSymbols(source: string, filePath: string): string[] {
  return extractFunctions(source, filePath)
}

function isReserved(name: string): boolean {
  const reserved = new Set(['if', 'for', 'while', 'return', 'import', 'from', 'const', 'let', 'var'])
  return reserved.has(name)
}
