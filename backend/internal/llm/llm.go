// Package llm orchestrates prompt construction and Ollama streaming for code analysis.
package llm

// UserContext carries developer profile data from the frontend onboarding flow.
type UserContext struct {
	PrimaryLanguage string
	ExperienceLevel string
	WorkspacePath   string
}

// PromptInput aggregates scan results and user context for prompt building.
type PromptInput struct {
	Symbol      string
	FilePath    string
	Matches     []MatchRef
	Snippet     string
	UserContext UserContext
}

// MatchRef is a lightweight reference to a ripgrep match for prompts.
type MatchRef struct {
	File    string
	Line    int
	Content string
}
