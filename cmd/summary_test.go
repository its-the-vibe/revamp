package cmd

import (
	"reflect"
	"testing"
)

func TestUniqueSortedLines(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   []string
	}{
		{
			name:  "empty string returns empty slice",
			input: "",
			want:  []string{},
		},
		{
			name:  "single line",
			input: "its-the-vibe/repo-one\n",
			want:  []string{"its-the-vibe/repo-one"},
		},
		{
			name:  "multiple unique lines sorted",
			input: "its-the-vibe/repo-two\nits-the-vibe/repo-one\n",
			want:  []string{"its-the-vibe/repo-one", "its-the-vibe/repo-two"},
		},
		{
			name:  "duplicate lines are deduplicated",
			input: "its-the-vibe/repo-one\nits-the-vibe/repo-two\nits-the-vibe/repo-one\n",
			want:  []string{"its-the-vibe/repo-one", "its-the-vibe/repo-two"},
		},
		{
			name:  "trailing newline without blank line",
			input: "its-the-vibe/repo-one\nits-the-vibe/repo-two",
			want:  []string{"its-the-vibe/repo-one", "its-the-vibe/repo-two"},
		},
		{
			name:  "output with all duplicates returns one entry",
			input: "its-the-vibe/repo-one\nits-the-vibe/repo-one\nits-the-vibe/repo-one\n",
			want:  []string{"its-the-vibe/repo-one"},
		},
		{
			name:  "alphabetical sorting",
			input: "org/c-repo\norg/a-repo\norg/b-repo\n",
			want:  []string{"org/a-repo", "org/b-repo", "org/c-repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniqueSortedLines(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uniqueSortedLines(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunSummaryValidation(t *testing.T) {
	// Save and restore flag state around each sub-test.
	origTitle := summaryByTitle
	origRepo := summaryByRepo
	origBranch := summaryBranch
	t.Cleanup(func() {
		summaryByTitle = origTitle
		summaryByRepo = origRepo
		summaryBranch = origBranch
	})

	t.Run("no flags returns error", func(t *testing.T) {
		summaryByTitle = false
		summaryByRepo = false
		summaryBranch = ""

		err := runSummary(summaryCmd, nil)
		if err == nil {
			t.Fatal("expected an error when no flags are provided, got nil")
		}
		want := "specify at least one of --title, --repo, or --branch"
		if err.Error() != want {
			t.Errorf("got error %q, want %q", err.Error(), want)
		}
	})

	t.Run("branch flag set with no org returns org error", func(t *testing.T) {
		summaryByTitle = false
		summaryByRepo = false
		summaryBranch = "renovate/foo"

		// Ensure org is not set so we hit the org-missing error before any gh call.
		err := runSummary(summaryCmd, nil)
		if err == nil {
			t.Fatal("expected an error when org is not set, got nil")
		}
		want := "org is not set; please configure it in config.yaml or via the ORG environment variable"
		if err.Error() != want {
			t.Errorf("got error %q, want %q", err.Error(), want)
		}
	})
}
