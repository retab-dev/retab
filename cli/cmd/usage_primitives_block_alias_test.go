package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// usageAliasBlock builds one WorkflowBlock with the optional declarative alias
// fields populated. Empty strings leave the corresponding pointer nil, which is
// how a hand-built (non-declarative) block comes back from the API.
func usageAliasBlock(id, declSourceID, declPath, parentID string) retab.WorkflowBlock {
	block := retab.WorkflowBlock{ID: id}
	if declSourceID != "" {
		block.DeclarativeSourceBlockID = &declSourceID
	}
	if declPath != "" {
		block.DeclarativePath = &declPath
	}
	if parentID != "" {
		block.ParentID = &parentID
	}
	return block
}

// A declarative spec block id is what `workflows spec get` prints and what
// `runs create` accepts, but the usage ledger indexes the generated runtime id.
// Filtering by the spec id used to return an empty 200, which on a usage export
// reads as "this block cost nothing".
func TestMatchUsageBlockIDFilterResolvesDeclarativeAliases(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		usageAliasBlock("block_s99qo983JCgNaEvZ3nycz", "block_start", "start", ""),
		usageAliasBlock("block_lte_nNvJ6RGq86aH3LEXD", "block_ex", "ex", ""),
	}
	for _, tc := range []struct {
		name      string
		requested string
		want      string
	}{
		{"declarative source block id", "block_ex", "block_lte_nNvJ6RGq86aH3LEXD"},
		{"declarative path", "ex", "block_lte_nNvJ6RGq86aH3LEXD"},
		{"runtime id passes through", "block_lte_nNvJ6RGq86aH3LEXD", "block_lte_nNvJ6RGq86aH3LEXD"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := matchUsageBlockIDFilter(tc.requested, blocks)
			if got.blockID != tc.want {
				t.Fatalf("block_id = %q, want %q", got.blockID, tc.want)
			}
			if !got.known {
				t.Fatalf("expected %q to resolve to a known block", tc.requested)
			}
			if got.requested != tc.requested {
				t.Fatalf("requested = %q, want %q", got.requested, tc.requested)
			}
		})
	}
}

// The ledger legitimately carries block ids that appear in no block list — the
// per-iteration ids a for_each expands to, and its sentinel row. Those must be
// forwarded untouched rather than rejected or rewritten.
func TestMatchUsageBlockIDFilterForwardsUnknownIDs(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		usageAliasBlock("block_loop", "block_loop", "loop", ""),
		usageAliasBlock("block_loop_item_extract", "block_loop_item_extract", "loop.item_extract", "block_loop"),
	}
	const iterationID = "block_loop_iter_0_block_loop_item_extract"
	got := matchUsageBlockIDFilter(iterationID, blocks)
	if got.blockID != iterationID {
		t.Fatalf("block_id = %q, want the value forwarded unchanged", got.blockID)
	}
	if got.known {
		t.Fatal("an iteration id is not in the block list; known should be false")
	}
	if got.loopParentID != "" {
		t.Fatalf("loopParentID = %q, want empty for an unresolved id", got.loopParentID)
	}
}

// An ambiguous alias must not be resolved to an arbitrary block: forward the raw
// value so the caller sees their own input in the query rather than a silent
// substitution.
func TestMatchUsageBlockIDFilterLeavesAmbiguousAliasAlone(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		usageAliasBlock("block_one", "dup", "", ""),
		usageAliasBlock("block_two", "", "dup", ""),
	}
	got := matchUsageBlockIDFilter("dup", blocks)
	if got.blockID != "dup" {
		t.Fatalf("block_id = %q, want the ambiguous value forwarded unchanged", got.blockID)
	}
	if got.known {
		t.Fatal("an ambiguous alias must not be reported as a known block")
	}
}

// A block nested in a for_each / while_loop container is recorded once per
// iteration, never under its own id. Both spellings of that block must carry the
// container id so the empty page can be explained.
func TestMatchUsageBlockIDFilterFlagsContainerChildren(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		usageAliasBlock("block_loop", "block_loop", "loop", ""),
		usageAliasBlock("block_loop_item_extract", "block_loop_item_extract", "loop.item_extract", "block_loop"),
	}
	for _, requested := range []string{"block_loop_item_extract", "loop.item_extract"} {
		got := matchUsageBlockIDFilter(requested, blocks)
		if got.loopParentID != "block_loop" {
			t.Fatalf("%q: loopParentID = %q, want block_loop", requested, got.loopParentID)
		}
	}
	// The container itself is top-level and must not be flagged.
	if got := matchUsageBlockIDFilter("block_loop", blocks); got.loopParentID != "" {
		t.Fatalf("container loopParentID = %q, want empty", got.loopParentID)
	}
}

func newUsageBlockWarnCommand(workflowID string) (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{Use: "primitives"}
	cmd.Flags().String("workflow-id", "", "")
	if workflowID != "" {
		_ = cmd.Flags().Set("workflow-id", workflowID)
	}
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	return cmd, &stderr
}

// The zero-row page for a container child is the confusing case: say why, and
// point at the command that lists the ledger's real block ids.
func TestWarnEmptyUsageBlockFilterExplainsContainerChildren(t *testing.T) {
	cmd, stderr := newUsageBlockWarnCommand("wrk_abc")
	filter := usageBlockIDFilter{
		requested:    "block_loop_item_extract",
		blockID:      "block_loop_item_extract",
		loopParentID: "block_loop",
		known:        true,
	}
	warnEmptyUsageBlockFilter(cmd, filter, 0)
	got := stderr.String()
	for _, want := range []string{
		"block_loop_item_extract",
		"block_loop_iter_<n>_block_loop_item_extract",
		"retab usage blocks --workflow-id wrk_abc",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("warning %q does not mention %q", got, want)
		}
	}
}

// The warning is diagnostic only: it must stay silent when rows came back, when
// no --block-id was given, and for blocks that are not container children.
func TestWarnEmptyUsageBlockFilterStaysSilent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		filter usageBlockIDFilter
		rows   int
	}{
		{"rows returned", usageBlockIDFilter{requested: "b", loopParentID: "block_loop"}, 3},
		{"no block filter", usageBlockIDFilter{}, 0},
		{"top-level block", usageBlockIDFilter{requested: "b", blockID: "b", known: true}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd, stderr := newUsageBlockWarnCommand("wrk_abc")
			warnEmptyUsageBlockFilter(cmd, tc.filter, tc.rows)
			if stderr.Len() != 0 {
				t.Fatalf("expected no warning, got %q", stderr.String())
			}
		})
	}
}

// runUsagePrimitivesAgainst points the CLI at a stub API, runs
// `usage primitives` with the given flags, and returns the block_id query
// argument the usage route actually received.
func runUsagePrimitivesAgainst(t *testing.T, handler http.Handler, flags map[string]string) string {
	t.Helper()
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(handler)
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for k, v := range flags {
		if err := usagePrimitivesCmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set --%s: %v", k, err)
		}
	}
	t.Cleanup(func() {
		for k := range flags {
			_ = usagePrimitivesCmd.Flags().Set(k, "")
		}
	})
	captureStd(t, func() {
		if err := usagePrimitivesCmd.RunE(usagePrimitivesCmd, nil); err != nil {
			t.Fatalf("usage primitives: %v", err)
		}
	})
	return usagePrimitivesLastBlockID
}

// usagePrimitivesLastBlockID is set by the stub handlers below.
var usagePrimitivesLastBlockID string

func usageBlockAliasHandler(t *testing.T, blocksStatus int, blocks []retab.WorkflowBlock) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/workflows/blocks":
			if blocksStatus != http.StatusOK {
				w.WriteHeader(blocksStatus)
				_, _ = w.Write([]byte(`{"detail":"workflow not found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          blocks,
				"list_metadata": map[string]any{"before": nil, "after": nil},
			})
		default:
			usagePrimitivesLastBlockID = r.URL.Query().Get("block_id")
			_ = json.NewEncoder(w).Encode(usagePrimitivesFixture())
		}
	})
}

// End-to-end: the declarative id a user copies out of `workflows spec get` must
// reach the usage route as the runtime id the ledger indexes by.
func TestUsagePrimitivesSendsResolvedBlockID(t *testing.T) {
	usagePrimitivesLastBlockID = ""
	handler := usageBlockAliasHandler(t, http.StatusOK, []retab.WorkflowBlock{
		usageAliasBlock("block_lte_runtime", "block_ex", "ex", ""),
	})
	got := runUsagePrimitivesAgainst(t, handler, map[string]string{
		"workflow-id": "wf_123",
		"block-id":    "block_ex",
	})
	if got != "block_lte_runtime" {
		t.Fatalf("block_id sent = %q, want block_lte_runtime", got)
	}
}

// The usage ledger outlives the workflows it bills for, so a block lookup that
// 404s (deleted workflow) must degrade to the raw value — never fail the read.
func TestUsagePrimitivesSurvivesBlockLookupFailure(t *testing.T) {
	usagePrimitivesLastBlockID = ""
	handler := usageBlockAliasHandler(t, http.StatusNotFound, nil)
	got := runUsagePrimitivesAgainst(t, handler, map[string]string{
		"workflow-id": "wf_deleted",
		"block-id":    "block_123",
	})
	if got != "block_123" {
		t.Fatalf("block_id sent = %q, want the raw block_123 forwarded", got)
	}
}

// Without --workflow-id there is nothing to resolve against; the value must go
// out untouched rather than triggering a lookup.
func TestUsagePrimitivesForwardsBlockIDWithoutWorkflow(t *testing.T) {
	usagePrimitivesLastBlockID = ""
	handler := usageBlockAliasHandler(t, http.StatusOK, []retab.WorkflowBlock{
		usageAliasBlock("block_lte_runtime", "block_ex", "ex", ""),
	})
	got := runUsagePrimitivesAgainst(t, handler, map[string]string{"block-id": "block_ex"})
	if got != "block_ex" {
		t.Fatalf("block_id sent = %q, want block_ex forwarded unresolved", got)
	}
}

// A blocks endpoint that hands back a cursor on every page would send
// AutoPaging into an endless walk. Alias resolution is a convenience, so it must
// give up on its own deadline and let the usage read proceed with the raw value.
func TestUsagePrimitivesBlockLookupCannotHangForever(t *testing.T) {
	usagePrimitivesLastBlockID = ""
	prev := usageBlockLookupTimeout
	usageBlockLookupTimeout = 250 * time.Millisecond
	t.Cleanup(func() { usageBlockLookupTimeout = prev })

	var pages atomic.Int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/workflows/blocks" {
			pages.Add(1)
			// Never terminates: every page points at another page.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []retab.WorkflowBlock{usageAliasBlock("block_runtime", "block_ex", "ex", "")},
				"list_metadata": map[string]any{"before": nil, "after": "block_next"},
			})
			return
		}
		usagePrimitivesLastBlockID = r.URL.Query().Get("block_id")
		_ = json.NewEncoder(w).Encode(usagePrimitivesFixture())
	})

	done := make(chan string, 1)
	go func() {
		done <- runUsagePrimitivesAgainst(t, handler, map[string]string{
			"workflow-id": "wf_123",
			"block-id":    "block_ex",
		})
	}()
	select {
	case got := <-done:
		if got != "block_ex" {
			t.Fatalf("block_id sent = %q, want the raw block_ex after the lookup timed out", got)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("usage primitives hung on block-id resolution after %d block pages", pages.Load())
	}
}
