package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

func TestRunMergeNoOrg(t *testing.T) {
	// Save and restore flag state and viper config around the test.
	origBranch := mergeBranch
	origDryRun := mergeDryRun
	origLimit := mergeLimit
	origMax := mergeMax
	origOrg := viper.GetString("org")
	t.Cleanup(func() {
		mergeBranch = origBranch
		mergeDryRun = origDryRun
		mergeLimit = origLimit
		mergeMax = origMax
		viper.Set("org", origOrg)
	})

	// Ensure org is not set so we hit the org-missing error before any gh call.
	viper.Set("org", "")
	mergeBranch = ""
	mergeDryRun = false
	mergeLimit = 100
	mergeMax = 0

	err := runMerge(mergeCmd, nil)
	if err == nil {
		t.Fatal("expected an error when org is not set, got nil")
	}
	want := "org is not set; please configure it in config.yaml or via the ORG environment variable"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}
