//go:build !retab_oagen_cli_files && !retab_oagen_cli_workflows_runs

package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Every list command must report the --before/--after conflict with the same
// concise message. `addListFlags` (and four workflow list commands) used to
// register cobra's MarkFlagsMutuallyExclusive, whose validateFlagGroups stage
// runs BEFORE RunE and emits "if any flags in the group [before after] are set
// none of the others can be; [after before] were all set". That shadowed the
// handwritten message the rest of the CLI prints — and made the
// `validateBeforeAfterMutex` call already present in `edits list`'s RunE dead
// code. Pin both halves: no cobra annotation anywhere, and the RunE check
// present on every command that has the flag pair.
func TestListCommandsShareOneBeforeAfterMessage(t *testing.T) {
	var walk func(cmd *cobra.Command)
	checked := 0
	walk = func(cmd *cobra.Command) {
		for _, child := range cmd.Commands() {
			walk(child)
		}
		before := cmd.Flags().Lookup("before")
		after := cmd.Flags().Lookup("after")
		if before == nil || after == nil {
			return
		}
		checked++
		path := cmd.CommandPath()
		// Cobra's group validation must not be registered: it fires first and
		// replaces the concise message.
		for _, f := range []*pflag.Flag{before, after} {
			if _, ok := f.Annotations["cobra_annotation_mutually_exclusive"]; ok {
				t.Errorf("%s: --%s carries cobra's mutex annotation; that shadows the concise RunE message", path, f.Name)
			}
		}
		// The RunE-level check must actually fire.
		if err := before.Value.Set("AAA"); err != nil {
			t.Fatalf("%s: set --before: %v", path, err)
		}
		if err := after.Value.Set("BBB"); err != nil {
			t.Fatalf("%s: set --after: %v", path, err)
		}
		err := validateBeforeAfterMutex(cmd)
		if err == nil {
			t.Errorf("%s: validateBeforeAfterMutex accepted both flags", path)
		} else if err.Error() != "--before and --after are mutually exclusive" {
			t.Errorf("%s: unexpected message %q", path, err)
		}
		_ = before.Value.Set("")
		_ = after.Value.Set("")
	}
	walk(rootCmd)
	if checked == 0 {
		t.Fatal("walked the command tree but found no command with --before/--after")
	}
	t.Logf("checked %d list commands", checked)
}

// A workflow run that settles at "failed" must end the wait. The terminal set
// omitted it while the CLI's own --status allowlist included it, so
// `runs create --wait` / `runs wait` polled a failed run for the full
// --timeout-seconds and then reported a timeout instead of the failure.
func TestWorkflowRunWaitTreatsFailedAsTerminal(t *testing.T) {
	for _, status := range []string{"completed", "error", "failed", "cancelled", "awaiting_review"} {
		if !workflowRunWaitTerminalStatuses[status] {
			t.Errorf("status %q is not treated as terminal by the run wait loop", status)
		}
	}
	// pending -> queued -> running are the in-flight statuses; the wait must
	// keep polling through them.
	inFlight := map[string]bool{"pending": true, "queued": true, "running": true}
	for status := range inFlight {
		if workflowRunWaitTerminalStatuses[status] {
			t.Errorf("in-flight status %q must not end the wait", status)
		}
	}
	// Every settled status the --status filter accepts must also end a wait, so
	// the two tables cannot drift apart again — that drift is what hid "failed".
	for status := range allowedWorkflowRunStatuses {
		if inFlight[status] {
			continue
		}
		if !workflowRunWaitTerminalStatuses[status] {
			t.Errorf("--status accepts settled status %q but the wait loop never settles on it", status)
		}
	}
}

// validateDateRange only parsed "2006-01-02", but --from-date/--to-date are
// registered as rfc3339FlagValue on the run/experiment list commands, which
// deliberately also accept a full RFC3339 timestamp. The reversed-range guard
// therefore silently no-opped for exactly the inputs it was written to catch.
func TestValidateDateRangeCoversRFC3339Bounds(t *testing.T) {
	reversed := [][2]string{
		{"2026-06-01", "2026-01-01"},
		{"2026-06-01T00:00:00Z", "2026-01-01T00:00:00Z"},
		{"2026-06-01", "2026-01-01T00:00:00Z"},
		{"2026-06-01T00:00:00Z", "2026-01-01"},
		{"2026-06-01T00:00:00", "2026-01-01T00:00:00"},
	}
	for _, tc := range reversed {
		if err := validateDateRange("from-date", "to-date", tc[0], tc[1]); err == nil {
			t.Errorf("reversed range %s..%s was accepted", tc[0], tc[1])
		}
	}
	ordered := [][2]string{
		{"2026-01-01", "2026-06-01"},
		{"2026-01-01T00:00:00Z", "2026-06-01T00:00:00Z"},
		{"2026-01-01", "2026-06-01T00:00:00Z"},
		// Same-day mixed pair: a bare upper bound means END of that UTC day,
		// matching the backend's parseISO, so this is a valid range.
		{"2026-06-01T12:00:00Z", "2026-06-01"},
		{"2026-06-01", "2026-06-01"},
		// Unparseable bounds are deferred to the server, not rejected here.
		{"not-a-date", "2026-01-01"},
	}
	for _, tc := range ordered {
		if err := validateDateRange("from-date", "to-date", tc[0], tc[1]); err != nil {
			t.Errorf("valid range %s..%s was rejected: %v", tc[0], tc[1], err)
		}
	}
}

// `files download` registered its destination flag as --output, which shadows
// the root's persistent --output (json|table|csv) because cobra's AddFlagSet
// skips names already in the local set. `retab files download <id> --output
// json` then wrote the file to a local file named "json". The destination flag
// is --out, matching every sibling that writes bytes.
func TestFilesDownloadDoesNotShadowGlobalOutputFlag(t *testing.T) {
	// Assert on LOCAL flags: cobra merges a parent's persistent flags into
	// cmd.Flags() once the command has been parsed, so by the time the whole
	// package's tests have run, Flags().Lookup("output") legitimately finds the
	// inherited global — which is the very thing this fix restored.
	if local := filesDownloadCmd.LocalFlags().Lookup("output"); local != nil {
		if global := rootCmd.PersistentFlags().Lookup("output"); global == nil || local != global {
			t.Fatal("files download declares its own --output, shadowing the global format flag")
		}
	}
	out := filesDownloadCmd.Flags().Lookup("out")
	if out == nil {
		t.Fatal("files download is missing its --out destination flag")
	}
	if out.Shorthand != "o" {
		t.Errorf("--out shorthand = %q, want \"o\" (the form the help text documents)", out.Shorthand)
	}
	// The root's persistent --output must exist and be the format flag, so
	// nothing local can be mistaken for it. (Deliberately read-only: the flag is
	// process-global state shared with every other test in this package.)
	global := rootCmd.PersistentFlags().Lookup("output")
	if global == nil {
		t.Fatal("root is missing its persistent --output format flag")
	}
	if !strings.Contains(global.Usage, "json") {
		t.Errorf("root --output usage = %q, want the json|table|csv format flag", global.Usage)
	}
}

// selectSheet resolved an all-digit --sheet as an index and hard-errored when
// out of range, so a sheet literally named "2024" (year tabs are everywhere in
// real workbooks) was unreachable by any selector.
func TestSelectSheetFallsBackToDigitSheetName(t *testing.T) {
	result := &ParseResult{
		Filename: "book.xlsx",
		Sheets: []SheetData{
			{Name: "Summary"},
			{Name: "2024"},
		},
	}
	newCmd := func(sheet string) *cobra.Command {
		c := &cobra.Command{}
		c.Flags().String("sheet", "", "")
		if sheet != "" {
			_ = c.Flags().Set("sheet", sheet)
		}
		return c
	}
	got, err := selectSheet(newCmd("2024"), result, kindSpreadsheet)
	if err != nil {
		t.Fatalf("--sheet 2024: %v", err)
	}
	if got.Name != "2024" {
		t.Errorf("--sheet 2024 selected %q, want the sheet named 2024", got.Name)
	}
	// A usable index still wins, so no currently-working invocation changes.
	if got, err = selectSheet(newCmd("2"), result, kindSpreadsheet); err != nil || got.Name != "2024" {
		t.Errorf("--sheet 2 = (%v, %v), want the second sheet", got.Name, err)
	}
	if got, err = selectSheet(newCmd("1"), result, kindSpreadsheet); err != nil || got.Name != "Summary" {
		t.Errorf("--sheet 1 = (%v, %v), want the first sheet", got.Name, err)
	}
	// A number that is neither a valid index nor a sheet name still errors.
	if _, err = selectSheet(newCmd("99"), result, kindSpreadsheet); err == nil {
		t.Error("--sheet 99 was accepted")
	}
}

// Every extension detectKind accepts as text must map to a real content type.
// Only .txt/.md/.json were listed, so the other twelve reported
// "application/octet-stream" for content the CLI had just decoded as text.
func TestMIMEForExtCoversEveryTextExtension(t *testing.T) {
	for ext := range textExtensions {
		got := mimeForExt(ext)
		if got == "application/octet-stream" || got == "" {
			t.Errorf("mimeForExt(%q) = %q, want a real text content type", ext, got)
		}
	}
	for ext, want := range map[string]string{
		".html": "text/html",
		".htm":  "text/html",
		".yaml": "application/yaml",
		".yml":  "application/yaml",
		".json": "application/json",
		".txt":  "text/plain",
		".log":  "text/plain",
		".toml": "application/toml",
	} {
		if got := mimeForExt(ext); got != want {
			t.Errorf("mimeForExt(%q) = %q, want %q", ext, got, want)
		}
	}
	// Case-insensitive, and genuinely unknown extensions still fall back.
	if got := mimeForExt(".HTML"); got != "text/html" {
		t.Errorf("mimeForExt(.HTML) = %q, want text/html", got)
	}
	if got := mimeForExt(".zzzznotathing"); got != "application/octet-stream" {
		t.Errorf("mimeForExt(unknown) = %q, want application/octet-stream", got)
	}
}

// Guard against the help text drifting away from the flag: `files download`'s
// Long/Example text documents the -o form, which must keep working.
func TestFilesDownloadHelpMatchesFlags(t *testing.T) {
	if !strings.Contains(filesDownloadCmd.Long, "-o") {
		t.Error("files download Long text no longer mentions -o")
	}
	if !strings.Contains(filesDownloadCmd.Example, "-o ") {
		t.Error("files download Example no longer demonstrates -o")
	}
}

// `tables query --output csv` wrote cells and column names straight into the
// csv.Writer, while every other CSV path in the CLI goes through writeCSV ->
// sanitizeCSVCell. A cell (or a user-uploaded column name) beginning =/+/-/@
// was therefore emitted verbatim for a spreadsheet to execute on open, in the
// same command family whose sibling neutralizes it.
func TestSanitizeCSVCellNeutralizesFormulaTriggers(t *testing.T) {
	for _, raw := range []string{
		`=HYPERLINK("http://evil","x")`,
		`@SUM(A1)`,
		`+1+1`,
		`-1-1`,
		"\tlead-tab",
	} {
		if got := sanitizeCSVCell(raw); got != "'"+raw {
			t.Errorf("sanitizeCSVCell(%q) = %q, want it prefixed with an apostrophe", raw, got)
		}
	}
	// Real negative numbers must stay numeric, not become text.
	for _, raw := range []string{"-12.5", "+3", "0", "12"} {
		if got := sanitizeCSVCell(raw); got != raw {
			t.Errorf("sanitizeCSVCell(%q) = %q, want it unchanged", raw, got)
		}
	}
}

// An explicit field_mappings entry must beat a payload key that happens to be
// named like the mapping's TARGET. The passthrough loop overwrote
// unconditionally, so {"order_id":"id"} against {"order_id":..,"id":..} posted
// the stale `id` — silently, and to the live endpoint under --execute.
func TestAPICallFieldMappingBeatsCollidingPayloadKey(t *testing.T) {
	got := mapAndFilterLocalAPICallInput(
		map[string]any{"order_id": "ord_123", "id": "legacy", "other": "kept"},
		nil,
		map[string]string{"order_id": "id"},
	)
	if got["id"] != "ord_123" {
		t.Errorf("mapped id = %v, want ord_123 (the explicit mapping)", got["id"])
	}
	if _, ok := got["order_id"]; ok {
		t.Error("mapping source order_id leaked into the request body")
	}
	if got["other"] != "kept" {
		t.Errorf("unmapped key dropped: %v", got["other"])
	}
	// No mappings at all: everything passes through untouched.
	plain := mapAndFilterLocalAPICallInput(map[string]any{"a": 1, "b": 2}, nil, nil)
	if plain["a"] != 1 || plain["b"] != 2 || len(plain) != 2 {
		t.Errorf("unmapped passthrough changed: %v", plain)
	}
}

// `auth login --api-key rt_test_...` stores the key in cfg.APIKey, and the
// stored-key branch hardcoded production — so a test key demanded the
// production confirmation (and hard-failed CI) purely because it had been
// saved rather than passed via --api-key, which classifies it correctly.
func TestStoredAPIKeyEnvironmentMatchesItsPrefix(t *testing.T) {
	for key, want := range map[string]string{
		"rt_test_abc":       slugTest,
		"sk_retab_test_abc": slugTest,
		"rt_live_abc":       slugProduction,
		"sk_retab_abc":      slugProduction,
	} {
		if got := environmentFromKeyPrefix(key); got != want {
			t.Errorf("environmentFromKeyPrefix(%q) = %q, want %q", key, got, want)
		}
	}
	// An unplaceable prefix must still fall back to production in the stored
	// branch — the safe default for a key the CLI cannot classify.
	if got := environmentFromKeyPrefix("weird_prefix_abc"); got != "" {
		t.Errorf("environmentFromKeyPrefix(unknown) = %q, want empty", got)
	}
}
