package flow

// ExtractFlow merges Tree-sitter and regex extractions (TS wins per line).
func ExtractFlow(content, filePath, symbol, lang string) []Step {
	src, resolvedLang := PrepareSource(content, filePath, lang)
	tsSteps := ExtractTreeSitter(src, symbol, resolvedLang)
	rxSteps := ExtractRegex(src, symbol, resolvedLang)
	steps := mergeSteps(tsSteps, rxSteps)
	return EnrichSteps(steps)
}

func mergeSteps(primary, fallback []Step) []Step {
	if len(primary) == 0 {
		return fallback
	}
	if len(fallback) == 0 {
		return primary
	}

	byLine := map[int]Step{}
	for _, s := range fallback {
		byLine[s.Line] = s
	}
	for _, s := range primary {
		merged := s
		merged.Source = "merged"
		if fb, ok := byLine[s.Line]; ok {
			if merged.CalleeSymbol == "" {
				merged.CalleeSymbol = fb.CalleeSymbol
			}
			if merged.CalleeQualified == "" {
				merged.CalleeQualified = fb.CalleeQualified
			}
			if merged.Code == "" {
				merged.Code = fb.Code
			}
		}
		byLine[s.Line] = merged
	}

	lines := sortedLines(byLine)
	out := make([]Step, 0, len(lines))
	for _, line := range lines {
		out = append(out, byLine[line])
	}
	return out
}

func sortedLines(m map[int]Step) []int {
	lines := make([]int, 0, len(m))
	for k := range m {
		lines = append(lines, k)
	}
	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			if lines[j] < lines[i] {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}
	}
	return lines
}
