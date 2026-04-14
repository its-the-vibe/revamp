package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	summaryByTitle  bool
	summaryByRepo   bool
	summaryByBranch bool
	summaryLimit    int
	summaryHead     string
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarise open Renovate PRs for the configured organisation",
	RunE:  runSummary,
}

func init() {
	summaryCmd.Flags().BoolVar(&summaryByTitle, "title", false, "Group and count open Renovate PRs by title")
	summaryCmd.Flags().BoolVar(&summaryByRepo, "repo", false, "Group and count open Renovate PRs by repository")
	summaryCmd.Flags().BoolVar(&summaryByBranch, "branch", false, "Group and count open Renovate PRs by branch name")
	summaryCmd.Flags().IntVar(&summaryLimit, "limit", 100, "Maximum number of PRs to fetch")
	summaryCmd.Flags().StringVar(&summaryHead, "head", "", "List repositories with open PRs from the given branch")
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
	if !summaryByTitle && !summaryByRepo && !summaryByBranch && summaryHead == "" {
		return fmt.Errorf("specify at least one of --title, --repo, --branch, or --head")
	}

	org := viper.GetString("org")
	if org == "" {
		return fmt.Errorf("org is not set; please configure it in config.yaml or via the ORG environment variable")
	}

	if summaryByTitle {
		if err := printSummary(org, summaryLimit, "title", `.[].title`); err != nil {
			return err
		}
	}

	if summaryByRepo {
		if err := printSummary(org, summaryLimit, "repository", `.[].repository.nameWithOwner`); err != nil {
			return err
		}
	}

	if summaryByBranch {
		if err := printBranchSummary(org, summaryLimit); err != nil {
			return err
		}
	}

	if summaryHead != "" {
		if err := printBranchRepos(org, summaryHead); err != nil {
			return err
		}
	}

	return nil
}

func printSummary(org string, limit int, jsonField, jqExpr string) error {
	ghArgs := []string{
		"search", "prs",
		"--owner", org,
		"--author", "app/renovate",
		"--state", "open",
		"-L", fmt.Sprintf("%d", limit),
		"--json", jsonField,
		"--jq", jqExpr,
	}

	ghCmd := exec.Command("gh", ghArgs...)
	ghCmd.Stderr = os.Stderr

	output, err := ghCmd.Output()
	if err != nil {
		return fmt.Errorf("gh command failed: %w", err)
	}

	counts := map[string]int{}
	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	for _, line := range lines {
		if line != "" {
			counts[line]++
		}
	}

	type entry struct {
		name  string
		count int
	}
	entries := make([]entry, 0, len(counts))
	for name, count := range counts {
		entries = append(entries, entry{name, count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})

	for _, e := range entries {
		fmt.Printf("%6d %s\n", e.count, e.name)
	}

	return nil
}

func printBranchRepos(org, branch string) error {
	ghArgs := []string{
		"search", "prs",
		"--owner", org,
		"--head", branch,
		"--state", "open",
		"--json", "repository",
		"--jq", `.[].repository.nameWithOwner`,
	}

	ghCmd := exec.Command("gh", ghArgs...)
	ghCmd.Stderr = os.Stderr

	output, err := ghCmd.Output()
	if err != nil {
		return fmt.Errorf("gh command failed: %w", err)
	}

	for _, line := range uniqueSortedLines(string(output)) {
		fmt.Println(line)
	}

	return nil
}

// printBranchSummary lists unique Renovate branch names with their PR counts.
// It first fetches all open Renovate PR titles and URLs in one gh search call,
// then for each unique title (Renovate reuses the same branch name for the same
// update across repos) it resolves the branch name with a single gh pr view call.
func printBranchSummary(org string, limit int) error {
	// Step 1: one search call — get title + URL for every open Renovate PR.
	ghArgs := []string{
		"search", "prs",
		"--owner", org,
		"--author", "app/renovate",
		"--state", "open",
		"-L", fmt.Sprintf("%d", limit),
		"--json", "title,url",
		"--jq", `.[] | [.title, .url] | @tsv`,
	}

	ghCmd := exec.Command("gh", ghArgs...)
	ghCmd.Stderr = os.Stderr

	output, err := ghCmd.Output()
	if err != nil {
		return fmt.Errorf("gh command failed: %w", err)
	}

	// Count PRs per title and keep one representative URL per title.
	titleCount, titleURL := parseTitleURLLines(string(output))
	if len(titleURL) == 0 {
		return nil
	}

	// Step 2: one gh pr view call per unique title to resolve the branch name.
	branchCount := map[string]int{}
	for title, url := range titleURL {
		viewCmd := exec.Command("gh", "pr", "view", url, "--json", "headRefName", "--jq", ".headRefName")
		viewCmd.Stderr = os.Stderr

		viewOut, err := viewCmd.Output()
		if err != nil {
			return fmt.Errorf("gh pr view failed for %s: %w", url, err)
		}

		branch := strings.TrimRight(string(viewOut), "\n")
		if branch != "" {
			branchCount[branch] += titleCount[title]
		}
	}

	// Step 3: sort by count descending, then alphabetically, and print.
	type entry struct {
		name  string
		count int
	}
	entries := make([]entry, 0, len(branchCount))
	for name, count := range branchCount {
		entries = append(entries, entry{name, count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})

	for _, e := range entries {
		fmt.Printf("%6d %s\n", e.count, e.name)
	}

	return nil
}

// parseTitleURLLines parses tab-separated (title, url) lines produced by the
// jq expression `.[] | [.title, .url] | @tsv`. It returns:
//   - titleCount: how many PRs share each title
//   - titleURL: one representative PR URL per unique title
func parseTitleURLLines(output string) (titleCount map[string]int, titleURL map[string]string) {
	titleCount = map[string]int{}
	titleURL = map[string]string{}

	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		title, url := parts[0], parts[1]
		titleCount[title]++
		if _, exists := titleURL[title]; !exists {
			titleURL[title] = url
		}
	}

	return
}

// uniqueSortedLines splits output by newlines, deduplicates, and returns the
// entries sorted alphabetically. Empty lines are ignored.
func uniqueSortedLines(output string) []string {
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		if line != "" {
			seen[line] = true
		}
	}
	result := make([]string, 0, len(seen))
	for line := range seen {
		result = append(result, line)
	}
	sort.Strings(result)
	return result
}
