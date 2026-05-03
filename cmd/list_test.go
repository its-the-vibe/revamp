package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

func TestRunListNoOrg(t *testing.T) {
	// Save and restore viper config around the test.
	origOrg := viper.GetString("org")
	t.Cleanup(func() {
		viper.Set("org", origOrg)
	})

	// Ensure org is not set so we hit the org-missing error before any gh call.
	viper.Set("org", "")

	err := runList(listCmd, nil)
	if err == nil {
		t.Fatal("expected an error when org is not set, got nil")
	}
	want := "org is not set; please configure it in config.yaml or via the ORG environment variable"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err.Error(), want)
	}
}
