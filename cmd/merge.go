package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	mergeBranch  string
	mergeDryRun  bool
	mergeVerbose bool
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge open Renovate PRs for the configured organisation",
	RunE:  runMerge,
}

func init() {
	mergeCmd.Flags().StringVar(&mergeBranch, "branch", "", "Only merge PRs from the given branch")
	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "Print what would be merged without actually merging")
	mergeCmd.Flags().BoolVar(&mergeVerbose, "verbose", false, "Print additional details during merging")
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) error {
	org := viper.GetString("org")
	if org == "" {
		return fmt.Errorf("org is not set; please configure it in config.yaml or via the ORG environment variable")
	}

	var ghArgs []string
	if mergeBranch != "" {
		ghArgs = []string{
			"search", "prs",
			"--owner", org,
			"--head", mergeBranch,
			"--state", "open",
			"--json", "url,title",
			"--jq", `.[] | .url + "\t" + .title`,
		}
	} else {
		ghArgs = []string{
			"search", "prs",
			"--owner", org,
			"--author", "app/renovate",
			"--state", "open",
			"-L", "100",
			"--json", "url,title",
			"--jq", `.[] | .url + "\t" + .title`,
		}
	}

	ghCmd := exec.Command("gh", ghArgs...)
	ghCmd.Stderr = os.Stderr

	output, err := ghCmd.Output()
	if err != nil {
		return fmt.Errorf("gh command failed: %w", err)
	}

	type prEntry struct {
		url   string
		title string
	}

	var prs []prEntry
	for _, line := range strings.Split(strings.TrimRight(string(output), "\n"), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		entry := prEntry{url: parts[0]}
		if len(parts) == 2 {
			entry.title = parts[1]
		}
		prs = append(prs, entry)
	}

	if len(prs) == 0 {
		fmt.Println("No open Renovate PRs found.")
		return nil
	}

	if mergeDryRun {
		fmt.Printf("Dry run: would merge %d PR(s):\n", len(prs))
		for _, pr := range prs {
			fmt.Printf("  %s  %s\n", pr.url, pr.title)
		}
		return nil
	}

	merged := 0
	failed := 0
	var failures []string

	for _, pr := range prs {
		if mergeVerbose {
			fmt.Printf("Merging: %s  %s\n", pr.url, pr.title)
		}

		mergeArgs := exec.Command("gh", "pr", "merge", pr.url, "--squash")
		mergeArgs.Stderr = os.Stderr

		if err := mergeArgs.Run(); err != nil {
			msg := fmt.Sprintf("%s  %s", pr.url, pr.title)
			fmt.Fprintf(os.Stderr, "Failed to merge %s: %v\n", msg, err)
			failures = append(failures, msg)
			failed++
		} else {
			if mergeVerbose {
				fmt.Printf("Merged:  %s  %s\n", pr.url, pr.title)
			}
			merged++
		}
	}

	fmt.Printf("\nSummary: %d merged, %d failed\n", merged, failed)
	if len(failures) > 0 {
		fmt.Fprintln(os.Stderr, "\nFailed PRs:")
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
	}

	return nil
}
