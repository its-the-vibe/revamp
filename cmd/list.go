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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List open Renovate PRs for the configured organisation",
	RunE:  runList,
}

var head string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&head, "head", "", "Filter PRs by base branch")
}

func runList(cmd *cobra.Command, args []string) error {
	org := viper.GetString("org")
	if org == "" {
		return fmt.Errorf("org is not set; please configure it in config.yaml or via the ORG environment variable")
	}

	ghArgs := []string{
		"search", "prs",
		"--owner", org,
		"--author", "app/renovate",
		"--state", "open",
		"-L", "100",
	}

	if head == "" {
		ghArgs = append(ghArgs, "--json", "title,repository")
		ghArgs = append(ghArgs, "--jq", `.[] | "\(.repository.nameWithOwner) | \(.title)"`)
	} else {
		ghArgs = append(ghArgs, "--json", "url")
		ghArgs = append(ghArgs, "--head", head)
	}

	ghCmd := exec.Command("gh", ghArgs...)
	ghCmd.Stderr = os.Stderr

	output, err := ghCmd.Output()
	if err != nil {
		return fmt.Errorf("gh command failed: %w", err)
	}

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	var filtered []string
	for _, line := range lines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	sort.Strings(filtered)
	for _, line := range filtered {
		fmt.Println(line)
	}

	return nil
}
