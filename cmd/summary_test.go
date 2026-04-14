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

func TestParseTitleURLLines(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantCount      map[string]int
		wantURLPresent map[string]bool // just check a URL was recorded, not its exact value
	}{
		{
			name:           "empty input",
			input:          "",
			wantCount:      map[string]int{},
			wantURLPresent: map[string]bool{},
		},
		{
			name:  "single PR",
			input: "Update lodash\thttps://github.com/org/repo/pull/1\n",
			wantCount: map[string]int{
				"Update lodash": 1,
			},
			wantURLPresent: map[string]bool{
				"Update lodash": true,
			},
		},
		{
			name: "two different titles",
			input: "Update lodash\thttps://github.com/org/repo-a/pull/1\n" +
				"Update axios\thttps://github.com/org/repo-b/pull/2\n",
			wantCount: map[string]int{
				"Update lodash": 1,
				"Update axios":  1,
			},
			wantURLPresent: map[string]bool{
				"Update lodash": true,
				"Update axios":  true,
			},
		},
		{
			name: "duplicate titles are counted",
			input: "Update lodash\thttps://github.com/org/repo-a/pull/1\n" +
				"Update lodash\thttps://github.com/org/repo-b/pull/2\n" +
				"Update lodash\thttps://github.com/org/repo-c/pull/3\n",
			wantCount: map[string]int{
				"Update lodash": 3,
			},
			wantURLPresent: map[string]bool{
				"Update lodash": true,
			},
		},
		{
			name:           "malformed lines without tab are skipped",
			input:          "no-tab-here\n",
			wantCount:      map[string]int{},
			wantURLPresent: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCount, gotURL := parseTitleURLLines(tt.input)

			if !reflect.DeepEqual(gotCount, tt.wantCount) {
				t.Errorf("parseTitleURLLines(%q) titleCount = %v, want %v", tt.input, gotCount, tt.wantCount)
			}

			for title, wantPresent := range tt.wantURLPresent {
				_, present := gotURL[title]
				if present != wantPresent {
					t.Errorf("parseTitleURLLines(%q) titleURL[%q] present=%v, want %v", tt.input, title, present, wantPresent)
				}
			}
		})
	}
}

func TestRunSummaryValidation(t *testing.T) {
	// Save and restore flag state around each sub-test.
	origTitle := summaryByTitle
	origRepo := summaryByRepo
	origBranch := summaryByBranch
	origHead := summaryHead
	t.Cleanup(func() {
		summaryByTitle = origTitle
		summaryByRepo = origRepo
		summaryByBranch = origBranch
		summaryHead = origHead
	})

	t.Run("no flags returns error", func(t *testing.T) {
		summaryByTitle = false
		summaryByRepo = false
		summaryByBranch = false
		summaryHead = ""

		err := runSummary(summaryCmd, nil)
		if err == nil {
			t.Fatal("expected an error when no flags are provided, got nil")
		}
		want := "specify at least one of --title, --repo, --branch, or --head"
		if err.Error() != want {
			t.Errorf("got error %q, want %q", err.Error(), want)
		}
	})

	t.Run("branch flag set with no org returns org error", func(t *testing.T) {
		summaryByTitle = false
		summaryByRepo = false
		summaryByBranch = true
		summaryHead = ""

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

	t.Run("head flag set with no org returns org error", func(t *testing.T) {
		summaryByTitle = false
		summaryByRepo = false
		summaryByBranch = false
		summaryHead = "renovate/foo"

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
