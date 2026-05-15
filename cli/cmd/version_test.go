package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVersionHonorsOutputJSON(t *testing.T) {
	if err := rootCmd.PersistentFlags().Set("output", "json"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := versionCmd.RunE(versionCmd, nil); err != nil {
			t.Fatal(err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("expected JSON version output, got %v for:\n%s", err, stdout)
	}
	if got["version"] != version {
		t.Fatalf("version = %q, want %q", got["version"], version)
	}
	if got["commit"] != commit {
		t.Fatalf("commit = %q, want %q", got["commit"], commit)
	}
	if got["built"] != date {
		t.Fatalf("built = %q, want %q", got["built"], date)
	}
}

func TestVersionHonorsOutputTable(t *testing.T) {
	if err := rootCmd.PersistentFlags().Set("output", "table"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rootCmd.PersistentFlags().Set("output", "") })

	stdout, stderr := captureStd(t, func() {
		if err := versionCmd.RunE(versionCmd, nil); err != nil {
			t.Fatal(err)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	for _, want := range []string{"VERSION", "COMMIT", "BUILT", version, commit, date} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected %q in table output:\n%s", want, stdout)
		}
	}
}
