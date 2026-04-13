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
	summaryByTitle bool
	summaryByRepo  bool
	summaryLimit   int
	summaryBranch  string
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarise open Renovate PRs for the configured organisation",
	RunE:  runSummary,
}

func init() {
	summaryCmd.Flags().BoolVar(&summaryByTitle, "title", false, "Group and count open Renovate PRs by title")
	summaryCmd.Flags().BoolVar(&summaryByRepo, "repo", false, "Group and count open Renovate PRs by repository")
	summaryCmd.Flags().IntVar(&summaryLimit, "limit", 100, "Maximum number of PRs to fetch")
	summaryCmd.Flags().StringVar(&summaryBranch, "branch", "", "List repositories with open PRs from the given branch")
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
	if !summaryByTitle && !summaryByRepo && summaryBranch == "" {
		return fmt.Errorf("specify at least one of --title, --repo, or --branch")
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

	if summaryBranch != "" {
		if err := printBranchRepos(org, summaryBranch); err != nil {
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

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Println(line)
		}
	}

	return nil
}
