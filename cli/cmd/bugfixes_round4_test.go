//go:build !retab_oagen_cli_files && !retab_oagen_cli_workflows_runs

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	retab "github.com/retab-dev/retab/clients/go"
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
// stored-key branch of resolveCredential hardcoded production — so a test key
// demanded the production confirmation (and hard-failed CI with "production
// write requires --confirm") purely because it had been saved rather than
// passed via --api-key, which classifies the identical key correctly.
//
// This drives resolveCredential itself: asserting on environmentFromKeyPrefix
// would pass against the unfixed code, since that helper was already correct —
// the bug was that the stored-key branch never consulted it.
func TestResolveCredentialStoredKeyEnvironmentMatchesPrefix(t *testing.T) {
	for _, tc := range []struct {
		key  string
		want string
	}{
		{"rt_test_abc123", slugTest},
		{"sk_retab_test_abc123", slugTest},
		{"rt_live_abc123", slugProduction},
		{"sk_retab_abc123", slugProduction},
		// An unplaceable prefix must still fail safe to production.
		{"weird_prefix_abc123", slugProduction},
	} {
		t.Run(tc.key, func(t *testing.T) {
			isolateHome(t)
			if err := saveConfig(retabConfig{APIKey: tc.key}); err != nil {
				t.Fatal(err)
			}
			cred, err := resolveCredential(newTestRootCmd())
			if err != nil {
				t.Fatal(err)
			}
			if cred.Source != sourceLegacyKey {
				t.Fatalf("source = %q, want %q", cred.Source, sourceLegacyKey)
			}
			if cred.ExpectedEnvironment != tc.want {
				t.Errorf("stored key %q resolved to environment %q, want %q",
					tc.key, cred.ExpectedEnvironment, tc.want)
			}
		})
	}
}

// `files inspect --render` created its temp output directory before resolving
// the liteparse binary, parsing the document, clipping pages and enforcing the
// 3-page cap — and returned from every one of those failures without removing
// it, orphaning an empty retab-inspect-* directory per failed invocation. Only
// a successful render, whose JSON reports the path, may leave it behind.
func TestInspectRenderCleansUpTempDirOnFailure(t *testing.T) {
	before, err := filepath.Glob(filepath.Join(os.TempDir(), "retab-inspect-*"))
	if err != nil {
		t.Fatalf("glob temp dirs: %v", err)
	}

	// A .txt file is not a pdf/image, so this fails at the kind check; point
	// --liteparse-bin at a nonexistent path too so the resolver would fail as
	// well. Either way nothing may be left on disk.
	dir := t.TempDir()
	doc := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(doc, []byte("hello\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	c := &cobra.Command{}
	c.Flags().String("render", "1", "")
	c.Flags().String("out", "", "")
	c.Flags().String("liteparse-bin", filepath.Join(dir, "definitely-not-here"), "")
	_ = c.Flags().Set("render", "1")

	// The command is expected to fail; we only care that it litters nothing.
	_ = inspectRender(context.Background(), c, doc, kindPDF, "1")

	after, err := filepath.Glob(filepath.Join(os.TempDir(), "retab-inspect-*"))
	if err != nil {
		t.Fatalf("glob temp dirs: %v", err)
	}
	if len(after) > len(before) {
		t.Errorf("failed --render leaked %d temp dir(s): before=%d after=%d",
			len(after)-len(before), len(before), len(after))
	}
}

// Two mapping sources pointing at one target is a config error, but it must
// not resolve differently on every invocation: Go randomizes map iteration and
// under --execute this is the body POSTed to a live endpoint. Sorting the
// sources makes the winner stable.
func TestAPICallFieldMappingIsDeterministic(t *testing.T) {
	seen := map[string]int{}
	for i := 0; i < 200; i++ {
		got := mapAndFilterLocalAPICallInput(
			map[string]any{"a": "from_a", "b": "from_b", "x": "from_payload_x"},
			nil,
			map[string]string{"a": "x", "b": "x"},
		)
		value, _ := got["x"].(string)
		seen[value]++
	}
	if len(seen) != 1 {
		t.Fatalf("duplicate mapping targets resolved nondeterministically across runs: %v", seen)
	}
}

// --filename must be honored on a path upload and must stay a bare basename.
// The flag was previously read only on the stdin branch.
func TestUploadFilenameOverride(t *testing.T) {
	newCmd := func(value string, set bool) *cobra.Command {
		c := &cobra.Command{}
		c.Flags().String("filename", "", "")
		if set {
			_ = c.Flags().Set("filename", value)
		}
		return c
	}
	got, err := uploadFilenameOverride(newCmd("Q3 Invoice.pdf", true))
	if err != nil {
		t.Fatalf("valid override rejected: %v", err)
	}
	if got != "Q3 Invoice.pdf" {
		t.Errorf("override = %q, want %q", got, "Q3 Invoice.pdf")
	}
	if got, err = uploadFilenameOverride(newCmd("", false)); err != nil || got != "" {
		t.Errorf("unset --filename = (%q, %v), want empty and no error", got, err)
	}
	// Anything that is not a plain basename is rejected — "." and ".." matter
	// because stageStdinUpload filepath.Joins this into a temp dir.
	for _, bad := range []string{"../../etc/passwd", "a/b.pdf", `a\b.pdf`, "..", "."} {
		if _, err := uploadFilenameOverride(newCmd(bad, true)); err == nil {
			t.Errorf("--filename %q was accepted", bad)
		}
	}
}

// The content type must be derived from the name actually recorded, so the
// stored filename and declared mime agree, and so the path branch matches the
// stdin branch (where the staged file is already named after --filename).
func TestUploadContentTypeFollowsEffectiveFilename(t *testing.T) {
	pdf := []byte("%PDF-1.4\n")
	if got := detectUploadContentType("report.txt", pdf); got != "text/plain; charset=utf-8" && got != "text/plain" {
		t.Errorf("detectUploadContentType(report.txt) = %q, want a text/plain type", got)
	}
	if got := detectUploadContentType("report.pdf", pdf); got != "application/pdf" {
		t.Errorf("detectUploadContentType(report.pdf) = %q, want application/pdf", got)
	}
	// A bare basename and a full path with the same extension must agree, so
	// switching to the effective filename cannot change the no-override case.
	if a, b := detectUploadContentType("x.pdf", pdf), detectUploadContentType("/tmp/deep/x.pdf", pdf); a != b {
		t.Errorf("basename vs path disagree: %q vs %q", a, b)
	}
}

// usage primitives' timestamp columns must canonicalize like every other
// timestamp column in the CLI; normalizeTimestampCell must leave non-timestamps
// untouched rather than mangling them.
func TestUsagePrimitiveTimestampColumnsAreNormalized(t *testing.T) {
	row := map[string]any{
		"created_at":   "2026-07-16T11:04:36.389000Z",
		"completed_at": "2026-07-16T11:05:01.120000Z",
	}
	for _, header := range []string{"CREATED_AT", "COMPLETED_AT"} {
		var column *TableColumn
		for i := range usagePrimitiveColumns {
			if usagePrimitiveColumns[i].Header == header {
				column = &usagePrimitiveColumns[i]
			}
		}
		if column == nil {
			t.Fatalf("usage primitives has no %s column", header)
		}
		got := column.Extract(row)
		if strings.Contains(got, ".") {
			t.Errorf("%s = %q, want a canonicalized second-precision timestamp", header, got)
		}
	}
	for _, passthrough := range []string{"", "-", "not a time"} {
		if got := normalizeTimestampCell(passthrough); got != passthrough {
			t.Errorf("normalizeTimestampCell(%q) = %q, want it unchanged", passthrough, got)
		}
	}
}

// The mime_type printed to the user must be derived from the same name that
// was declared to the server. shapeUploadResponse used to resolve its extension
// fallback against the LOCAL path, so `files upload ./scan.pdf --filename
// report.txt` declared text/plain on the wire and printed application/pdf — the
// exact self-contradiction the --filename fix set out to remove. The fallback
// now prefers the filename the server echoed (which is the name the upload
// declared), so the two agree by construction.
//
// This drives shapeUploadResponse itself rather than the helpers, so it fails
// if the call site is rewired to pass the local path again.
// pdfMagicBytes is enough for http.DetectContentType to sniff application/pdf.
var pdfMagicBytes = []byte("%PDF-1.4\n")

func TestUploadResponseMIMETypeFollowsEffectiveFilename(t *testing.T) {
	mimeTypeOf := func(out uploadResponse) string {
		for _, p := range out.pairs {
			if p.Key == "mime_type" {
				value, _ := p.Value.(string)
				return value
			}
		}
		return ""
	}

	// Server recorded the overridden name and returned no mime_type (the case
	// the upload code notes happens often).
	declared := detectUploadContentType(effectiveUploadFilename("/tmp/scan.pdf", "report.txt"), pdfMagicBytes)
	out, err := shapeUploadResponse(
		&retab.MIMEData{Filename: "report.txt", URL: "https://x/v1/files/file_abc123.pdf"},
		"/tmp/scan.pdf", declared)
	if err != nil {
		t.Fatalf("shapeUploadResponse: %v", err)
	}
	if got := mimeTypeOf(out); got != declared {
		t.Errorf("printed mime_type %q contradicts the declared content type %q", got, declared)
	}
	if got := mimeTypeOf(out); !strings.HasPrefix(got, "text/plain") {
		t.Errorf("printed mime_type = %q, want it to follow the .txt override", got)
	}

	// No override: unchanged behaviour.
	declared = detectUploadContentType(effectiveUploadFilename("/tmp/scan.pdf", ""), pdfMagicBytes)
	out, err = shapeUploadResponse(
		&retab.MIMEData{Filename: "scan.pdf", URL: "https://x/v1/files/file_abc123.pdf"},
		"/tmp/scan.pdf", declared)
	if err != nil {
		t.Fatalf("shapeUploadResponse: %v", err)
	}
	if got := mimeTypeOf(out); got != "application/pdf" {
		t.Errorf("no-override case changed: printed %q, want application/pdf", got)
	}

	// Server returned no filename at all: fall back to the local path.
	out, err = shapeUploadResponse(
		&retab.MIMEData{URL: "https://x/v1/files/file_abc123.pdf"}, "/tmp/scan.pdf", "")
	if err != nil {
		t.Fatalf("shapeUploadResponse: %v", err)
	}
	if got := mimeTypeOf(out); got != "application/pdf" {
		t.Errorf("empty server filename should fall back to the local path, got %q", got)
	}

	// An explicit server mime_type still wins over both.
	if got := resolveUploadMIMEType("image/png", "report.txt", "text/plain"); got != "image/png" {
		t.Errorf("server mime_type should win, got %q", got)
	}
	// The shared effective-name helper.
	if got := effectiveUploadFilename("/tmp/scan.pdf", "report.txt"); got != "report.txt" {
		t.Errorf("effectiveUploadFilename with override = %q, want report.txt", got)
	}
	if got := effectiveUploadFilename("/tmp/deep/scan.pdf", ""); got != "scan.pdf" {
		t.Errorf("effectiveUploadFilename without override = %q, want scan.pdf", got)
	}
}

// --filename is trimmed and validated identically on the path and stdin
// branches. `".. "` used to pass the guard and then resolve to the parent
// directory, because Windows strips trailing spaces when resolving a path.
func TestUploadFilenameGuardRejectsPaddedDotNames(t *testing.T) {
	for _, bad := range []string{"..", ".", ".. ", " ..", ". ", "  ..  ", "a/b", `a\b`} {
		if err := validateBareUploadFilename(bad); err == nil {
			t.Errorf("validateBareUploadFilename(%q) accepted it", bad)
		}
	}
	for _, ok := range []string{"a.pdf", "..a.pdf", "a..pdf", "...", " report.pdf "} {
		if err := validateBareUploadFilename(ok); err != nil {
			t.Errorf("validateBareUploadFilename(%q) rejected a legitimate name: %v", ok, err)
		}
	}
	if got := trimmedUploadFilename("  report.pdf  "); got != "report.pdf" {
		t.Errorf("trimmedUploadFilename = %q, want report.pdf", got)
	}
}
