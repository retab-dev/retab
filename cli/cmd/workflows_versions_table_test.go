package cmd

import (
	"strings"
	"testing"
)

// publishedVersionRows builds the response shape `workflows versions list`
// actually renders: the typed cliPaginatedList[cliPublishedWorkflowVersion] the
// command decodes into, not a hand-rolled map. rowField resolves struct fields
// by json tag, so the columns must work against the real type.
func publishedVersionRows() cliPaginatedList[cliPublishedWorkflowVersion] {
	live := "v2: + due_date"
	old := "v1: initial"
	return cliPaginatedList[cliPublishedWorkflowVersion]{
		Data: []cliPublishedWorkflowVersion{
			{
				ID:                "wph_live",
				WorkflowID:        "wrk_x",
				Version:           2,
				WorkflowVersionID: "ver_LIVE",
				Description:       &live,
				PublishedAt:       "2026-07-16T11:33:03.375Z",
				IsCurrent:         true,
			},
			{
				ID:                "wph_old",
				WorkflowID:        "wrk_x",
				Version:           1,
				WorkflowVersionID: "ver_OLD",
				Description:       &old,
				PublishedAt:       "2026-07-16T11:04:06.375Z",
				IsCurrent:         false,
			},
		},
	}
}

func renderPublishedVersions(t *testing.T, format OutputFormat) string {
	t.Helper()
	var buf strings.Builder
	if err := RenderList(&buf, format, publishedVersionRows(), publishedVersionColumns); err != nil {
		t.Fatalf("RenderList: %v", err)
	}
	return buf.String()
}

// TestVersionsListTableSurfacesTheActionableID is the regression that matters:
// the generic auto-renderer matched only `id` (these rows carry no
// name/type/created_at), collapsing the table to one column of `wph_...`
// publication-record ids. Nothing accepts a wph_ id — `versions get`, `diff` and
// `restore` all take the `ver_...` workflow_version_id — so the only identifier
// on screen 404s when pasted into the next command, and the one the user needs
// was hidden. Verified against staging: `versions get <wf> wph_...` returns
// "404 Workflow version not found" while the ver_ id resolves.
func TestVersionsListTableSurfacesTheActionableID(t *testing.T) {
	out := renderPublishedVersions(t, OutputTable)
	if !strings.Contains(out, "WORKFLOW_VERSION_ID") {
		t.Fatalf("table must expose the id that get/diff/restore accept:\n%s", out)
	}
	for _, want := range []string{"ver_LIVE", "ver_OLD"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table missing actionable id %q:\n%s", want, out)
		}
	}
	// The dead-end publication id must not be what the reader reaches for.
	for _, dead := range []string{"wph_live", "wph_old"} {
		if strings.Contains(out, dead) {
			t.Fatalf("table must not surface the non-actionable publication id %q:\n%s", dead, out)
		}
	}
}

// TestVersionsListTableMarksTheLiveVersion pins that "not current" and "field
// missing" do not render identically. is_current is a bool, and the zero value
// is false — routing it through the generic cell helper would blank the column
// for every non-live row via cellIsEmpty, leaving the reader unable to tell
// which version is actually serving.
func TestVersionsListTableMarksTheLiveVersion(t *testing.T) {
	out := renderPublishedVersions(t, OutputTable)
	if !strings.Contains(out, "CURRENT") {
		t.Fatalf("table must have a CURRENT column:\n%s", out)
	}
	var liveLine, oldLine string
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.Contains(line, "ver_LIVE"):
			liveLine = line
		case strings.Contains(line, "ver_OLD"):
			oldLine = line
		}
	}
	if liveLine == "" || oldLine == "" {
		t.Fatalf("expected a row per version:\n%s", out)
	}
	if !strings.Contains(liveLine, "yes") {
		t.Fatalf("live version row must be marked current: %q", liveLine)
	}
	if strings.Contains(oldLine, "yes") {
		t.Fatalf("superseded version row must not be marked current: %q", oldLine)
	}
}

// TestVersionsListTableShowsVersionAndDescription pins the columns that make the
// list readable at a glance — the human ordinal and what each publish changed.
func TestVersionsListTableShowsVersionAndDescription(t *testing.T) {
	out := renderPublishedVersions(t, OutputTable)
	for _, want := range []string{"VERSION", "DESCRIPTION", "v2: + due_date", "v1: initial"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table missing %q:\n%s", want, out)
		}
	}
}

// TestVersionsListTableNormalizesPublishedAt pins that published_at is
// canonicalized like every other timestamp column. The API returns millisecond
// precision; rendering it raw would print "2026-07-16T11:33:03.375Z" in a column
// that reads as second-precision RFC3339 everywhere else.
func TestVersionsListTableNormalizesPublishedAt(t *testing.T) {
	out := renderPublishedVersions(t, OutputTable)
	if !strings.Contains(out, "2026-07-16T11:33:03Z") {
		t.Fatalf("PUBLISHED_AT must be canonical second-precision RFC3339:\n%s", out)
	}
	if strings.Contains(out, ".375") {
		t.Fatalf("PUBLISHED_AT must drop fractional seconds:\n%s", out)
	}
}

// TestVersionsListCSVMatchesTheTable pins that --output csv carries the same
// columns as the table: RenderList feeds both from one spec, and a CSV that
// omitted the actionable id would defeat the scripting use case the format
// exists for.
func TestVersionsListCSVMatchesTheTable(t *testing.T) {
	out := renderPublishedVersions(t, OutputCSV)
	header, rest, found := strings.Cut(out, "\n")
	if !found {
		t.Fatalf("csv must have a header and rows:\n%s", out)
	}
	if header != "VERSION,WORKFLOW_VERSION_ID,CURRENT,DESCRIPTION,PUBLISHED_AT" {
		t.Fatalf("unexpected csv header: %q", header)
	}
	if !strings.Contains(rest, "2,ver_LIVE,yes,") {
		t.Fatalf("csv must carry the live row with its actionable id:\n%s", rest)
	}
	if !strings.Contains(rest, "1,ver_OLD,,") {
		t.Fatalf("csv must render a superseded row with an empty CURRENT cell:\n%s", rest)
	}
}
