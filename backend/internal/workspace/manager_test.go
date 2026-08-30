package workspace

import "testing"

func TestNormalizeGitHubURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"torvalds/linux", "https://github.com/torvalds/linux"},
		{"IBM/sarama", "https://github.com/IBM/sarama"},
		{"https://github.com/torvalds/linux", "https://github.com/torvalds/linux"},
		{"https://github.com/torvalds/linux/", "https://github.com/torvalds/linux"},
		{"https://github.com/torvalds/linux.git", "https://github.com/torvalds/linux"},
		{"github.com/torvalds/linux", "https://github.com/torvalds/linux"},
	}

	for _, tc := range tests {
		got, err := normalizeGitHubURL(tc.in)
		if err != nil {
			t.Fatalf("normalizeGitHubURL(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("normalizeGitHubURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeGitHubURLRejectsInvalid(t *testing.T) {
	_, err := normalizeGitHubURL("not-a-repo")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestGitHubRepoName(t *testing.T) {
	if got := githubRepoName("https://github.com/torvalds/linux"); got != "linux" {
		t.Fatalf("githubRepoName = %q", got)
	}
}
