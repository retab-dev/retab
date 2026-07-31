package cmd

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// A workflow block has two names: the spec alias the author wrote (all `spec get`
// ever shows them) and the runtime `block_<nanoid>` the server minted.
// `runs create --document <alias>=f.pdf` resolves the alias, so a spec author
// never sees the runtime id — and then hit a 404, or a misleading
// "422 Step data not found for this block in the given run", on every OTHER
// command that takes a block id.
//
// Resolution is LAZY by design: send what the user typed, look the block list up
// only if that call fails, retry once. That keeps a runtime id — all any script
// holds — at exactly one request, and means resolution can only turn a failure
// into a success, never perturb a call that already worked. The "costs one
// request" half of that contract is asserted explicitly below, because eager
// resolution is the obvious refactor and it is the one that breaks it.

// ---------------------------------------------------------------------------
// the pure matcher
// ---------------------------------------------------------------------------

func aliasBlock(id, declSourceID, declPath string) retab.WorkflowBlock {
	block := retab.WorkflowBlock{ID: id}
	if declSourceID != "" {
		block.DeclarativeSourceBlockID = &declSourceID
	}
	if declPath != "" {
		block.DeclarativePath = &declPath
	}
	return block
}

func TestMatchWorkflowBlockAlias(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		aliasBlock("block_runtime_a", "s31a_ex", "ex"),
		aliasBlock("block_runtime_b", "s31a_start", "start"),
	}
	cases := map[string]struct{ in, want string }{
		"declarative source id":  {"s31a_ex", "block_runtime_a"},
		"declarative path":       {"start", "block_runtime_b"},
		"already runtime id":     {"block_runtime_a", "block_runtime_a"},
		"unknown forwarded":      {"block_nope", "block_nope"},
		"empty forwarded":        {"", ""},
		"case sensitive":         {"S31A_EX", "S31A_EX"},
		"whitespace not trimmed": {" s31a_ex", " s31a_ex"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := matchWorkflowBlockAlias(tc.in, blocks); got != tc.want {
				t.Fatalf("matchWorkflowBlockAlias(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestMatchWorkflowBlockAliasOnAnEmptyBlockList(t *testing.T) {
	for _, blocks := range [][]retab.WorkflowBlock{nil, {}} {
		if got := matchWorkflowBlockAlias("s31a_ex", blocks); got != "s31a_ex" {
			t.Fatalf("empty block list resolved to %q, want the raw value", got)
		}
	}
}

// An alias that hits two blocks is not resolvable, and picking one arbitrarily
// would silently act on the WRONG block — the one failure mode worse than the
// 404 this whole feature exists to remove.
func TestMatchWorkflowBlockAliasRefusesAmbiguousHits(t *testing.T) {
	cases := map[string][]retab.WorkflowBlock{
		"source id vs path": {
			aliasBlock("block_runtime_a", "dup", ""),
			aliasBlock("block_runtime_b", "", "dup"),
		},
		"two source ids": {
			aliasBlock("block_runtime_a", "dup", ""),
			aliasBlock("block_runtime_b", "dup", ""),
		},
		"three-way": {
			aliasBlock("block_a", "dup", ""),
			aliasBlock("block_b", "", "dup"),
			aliasBlock("block_c", "dup", "dup"),
		},
	}
	for name, blocks := range cases {
		t.Run(name, func(t *testing.T) {
			if got := matchWorkflowBlockAlias("dup", blocks); got != "dup" {
				t.Fatalf("ambiguous alias resolved to %q, want the raw value", got)
			}
		})
	}
}

// One block matching on BOTH of its own alias fields is not ambiguous — it is one
// block, and it must resolve.
func TestMatchWorkflowBlockAliasResolvesASingleBlockMatchingBothFields(t *testing.T) {
	blocks := []retab.WorkflowBlock{aliasBlock("block_runtime", "same", "same")}
	if got := matchWorkflowBlockAlias("same", blocks); got != "block_runtime" {
		t.Fatalf("resolved to %q, want block_runtime", got)
	}
}

// A runtime id always wins over an alias that happens to share the string, so a
// concrete id is never redirected.
func TestMatchWorkflowBlockAliasPrefersRuntimeID(t *testing.T) {
	blocks := []retab.WorkflowBlock{
		aliasBlock("block_shared", "", ""),
		aliasBlock("block_other", "block_shared", ""),
	}
	if got := matchWorkflowBlockAlias("block_shared", blocks); got != "block_shared" {
		t.Fatalf("runtime id redirected to %q", got)
	}
}

// A for_each child carries the same alias fields as any other block; the matcher
// has no reason to treat it specially and must not.
func TestMatchWorkflowBlockAliasResolvesNestedChildren(t *testing.T) {
	parent := "block_loop_runtime"
	child := aliasBlock("block_child_runtime", "s31d_item", "loop.item")
	child.ParentID = &parent
	blocks := []retab.WorkflowBlock{aliasBlock(parent, "s31d_loop", "loop"), child}
	if got := matchWorkflowBlockAlias("s31d_item", blocks); got != "block_child_runtime" {
		t.Fatalf("nested child resolved to %q", got)
	}
	if got := matchWorkflowBlockAlias("loop.item", blocks); got != "block_child_runtime" {
		t.Fatalf("nested child path resolved to %q", got)
	}
}

// ---------------------------------------------------------------------------
// the lazy-retry wrapper
// ---------------------------------------------------------------------------

func TestCallWithBlockAliasFallback(t *testing.T) {
	boom := errors.New("422 step data not found")

	t.Run("success short-circuits: no resolve, one call", func(t *testing.T) {
		calls := 0
		resolves := 0
		got, err := callWithBlockAliasFallback("blk", func(id string) (string, error) {
			calls++
			return "ok:" + id, nil
		}, func() string { resolves++; return "other" })
		if err != nil || got != "ok:blk" {
			t.Fatalf("got %q, %v", got, err)
		}
		if calls != 1 {
			t.Fatalf("made %d calls on the happy path, want exactly 1", calls)
		}
		if resolves != 0 {
			t.Fatalf("resolved %d times on the happy path, want 0", resolves)
		}
	})

	t.Run("failure then a different id retries once", func(t *testing.T) {
		var seen []string
		got, err := callWithBlockAliasFallback("alias", func(id string) (string, error) {
			seen = append(seen, id)
			if id != "runtime" {
				return "", boom
			}
			return "ok", nil
		}, func() string { return "runtime" })
		if err != nil || got != "ok" {
			t.Fatalf("got %q, %v", got, err)
		}
		if len(seen) != 2 || seen[0] != "alias" || seen[1] != "runtime" {
			t.Fatalf("attempts = %v, want [alias runtime]", seen)
		}
	})

	t.Run("resolution to the same id does not retry", func(t *testing.T) {
		calls := 0
		_, err := callWithBlockAliasFallback("blk", func(string) (string, error) {
			calls++
			return "", boom
		}, func() string { return "blk" })
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want the original", err)
		}
		if calls != 1 {
			t.Fatalf("made %d calls, want 1 (nothing new to try)", calls)
		}
	})

	t.Run("resolution to empty does not retry", func(t *testing.T) {
		calls := 0
		_, err := callWithBlockAliasFallback("blk", func(string) (string, error) {
			calls++
			return "", boom
		}, func() string { return "" })
		if !errors.Is(err, boom) || calls != 1 {
			t.Fatalf("calls=%d err=%v", calls, err)
		}
	})

	t.Run("resolution is whitespace-trimmed", func(t *testing.T) {
		var seen []string
		_, err := callWithBlockAliasFallback("alias", func(id string) (string, error) {
			seen = append(seen, id)
			if id != "runtime" {
				return "", boom
			}
			return "ok", nil
		}, func() string { return "  runtime  " })
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(seen) != 2 || seen[1] != "runtime" {
			t.Fatalf("attempts = %v, want the trimmed id second", seen)
		}
	})

	// The user typed the alias. An error naming an id they never wrote would send
	// them looking in the wrong place, so a failed retry keeps the FIRST error.
	t.Run("a failed retry keeps the original error", func(t *testing.T) {
		retryErr := errors.New("404 block not found: block_runtime")
		_, err := callWithBlockAliasFallback("alias", func(id string) (string, error) {
			if id == "alias" {
				return "", boom
			}
			return "", retryErr
		}, func() string { return "runtime" })
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want the original error the user's id produced", err)
		}
	})

	// The result of a failed retry must not leak past the first attempt's value.
	t.Run("returns the first attempt's value when both fail", func(t *testing.T) {
		got, err := callWithBlockAliasFallback("alias", func(id string) (int, error) {
			if id == "alias" {
				return 1, boom
			}
			return 2, errors.New("also bad")
		}, func() string { return "runtime" })
		if err == nil {
			t.Fatal("expected an error")
		}
		if got != 1 {
			t.Fatalf("got %d, want the first attempt's value", got)
		}
	})
}

// ---------------------------------------------------------------------------
// the commands, end to end against a stub
// ---------------------------------------------------------------------------

// blockAliasServer stubs the lookups the resolvers need plus a catch-all that
// records the block id each real request carried.
type blockAliasServer struct {
	attempts     []string
	blocksCalls  int
	runCalls     int
	blocksStatus int
	runStatus    int
	blocks       []retab.WorkflowBlock
	workflowID   string
	// acceptOnly, when set, is the only block id the stubbed route accepts —
	// standing in for a route that indexes by runtime id alone.
	acceptOnly string
	payload    any
	// endlessPaging makes every blocks page point at another page.
	endlessPaging bool
}

func (s *blockAliasServer) lastAttempt() string {
	if len(s.attempts) == 0 {
		return ""
	}
	return s.attempts[len(s.attempts)-1]
}

func (s *blockAliasServer) handler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/v1/workflows/blocks" && r.Method == http.MethodGet:
			s.blocksCalls++
			if s.blocksStatus != 0 && s.blocksStatus != http.StatusOK {
				w.WriteHeader(s.blocksStatus)
				_, _ = w.Write([]byte(`{"detail":"workflow not found"}`))
				return
			}
			after := any(nil)
			if s.endlessPaging {
				after = "block_next"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          s.blocks,
				"list_metadata": map[string]any{"before": nil, "after": after},
			})
		case strings.HasPrefix(r.URL.Path, "/v1/workflows/runs/run_") && r.Method == http.MethodGet:
			s.runCalls++
			if s.runStatus != 0 && s.runStatus != http.StatusOK {
				w.WriteHeader(s.runStatus)
				_, _ = w.Write([]byte(`{"detail":"run not found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":          "run_123",
				"workflow_id": s.workflowID,
				"lifecycle":   map[string]any{"status": "completed"},
			})
		default:
			sent := r.URL.Query().Get("block_id")
			if r.Method == http.MethodPost {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
					if raw, ok := body["block_id"].(string); ok {
						sent = raw
					}
				}
			}
			s.attempts = append(s.attempts, sent)
			if s.acceptOnly != "" && sent != s.acceptOnly {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"detail":"Step data not found for this block in the given run"}`))
				return
			}
			payload := s.payload
			if payload == nil {
				payload = map[string]any{
					"id":        "exec_1",
					"run_id":    "run_123",
					"block_id":  sent,
					"lifecycle": map[string]any{"status": "completed"},
				}
			}
			_ = json.NewEncoder(w).Encode(payload)
		}
	})
}

func newBlockAliasEnv(t *testing.T, stub *blockAliasServer) {
	t.Helper()
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(stub.handler(t))
	t.Cleanup(server.Close)
	t.Setenv("RETAB_API_BASE_URL", server.URL)
}

func setAliasFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set --%s: %v", name, err)
	}
	t.Cleanup(func() { resetWorkflowBlocksExecutionsFlag(t, cmd, name) })
}

// runAliasCmd runs a command and returns both its error and its stderr. runE
// renders an *APIError to stderr and returns errSilent, so the server's own
// message is only visible in the captured stream.
func runAliasCmd(t *testing.T, cmd *cobra.Command, args []string) (error, string) {
	t.Helper()
	var err error
	_, stderr := captureStd(t, func() { err = cmd.RunE(cmd, args) })
	return err, stderr
}

// The whole point: the id a user copies out of `workflows spec get` must reach
// the block-executions route as the runtime id it looks the step up by.
func TestBlocksExecutionsCreateResolvesSpecAlias(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "s31g_fn", "fn")},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsBlocksExecutionsCreateCmd, "block-id", "s31g_fn")

	if err, _ := runAliasCmd(t, workflowsBlocksExecutionsCreateCmd, []string{"run_123"}); err != nil {
		t.Fatalf("executions create: %v", err)
	}
	if stub.lastAttempt() != "block_runtime" {
		t.Fatalf("block_id sent = %q, want block_runtime", stub.lastAttempt())
	}
	if len(stub.attempts) != 2 || stub.attempts[0] != "s31g_fn" {
		t.Fatalf("attempts = %v, want the alias then the runtime id", stub.attempts)
	}
}

// The declarative PATH spelling resolves too, not just the source block id.
func TestBlocksExecutionsCreateResolvesDeclarativePath(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "s31g_fn", "loop.fn")},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsBlocksExecutionsCreateCmd, "block-id", "loop.fn")

	if err, _ := runAliasCmd(t, workflowsBlocksExecutionsCreateCmd, []string{"run_123"}); err != nil {
		t.Fatalf("executions create: %v", err)
	}
	if stub.lastAttempt() != "block_runtime" {
		t.Fatalf("block_id sent = %q, want block_runtime", stub.lastAttempt())
	}
}

// Lazy resolution's contract: a runtime id — everything a script holds — must
// still cost exactly ONE request and trigger NO lookups. Eager resolution is the
// obvious refactor and it is the one that breaks this.
func TestBlocksExecutionsCreateWithARuntimeIDCostsOneRequest(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "s31g_fn", "fn")},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsBlocksExecutionsCreateCmd, "block-id", "block_runtime")

	if err, _ := runAliasCmd(t, workflowsBlocksExecutionsCreateCmd, []string{"run_123"}); err != nil {
		t.Fatalf("executions create: %v", err)
	}
	if len(stub.attempts) != 1 {
		t.Fatalf("attempts = %v, want exactly one", stub.attempts)
	}
	if stub.blocksCalls != 0 || stub.runCalls != 0 {
		t.Fatalf("happy path made %d block lookups and %d run lookups, want 0",
			stub.blocksCalls, stub.runCalls)
	}
}

// Resolution is a convenience, never a gate. Each lookup that can fail must leave
// the command sending the raw value so the SERVER owns the verdict.
func TestBlocksExecutionsCreateSurvivesLookupFailures(t *testing.T) {
	cases := map[string]*blockAliasServer{
		"blocks list 404": {workflowID: "wf_123", blocksStatus: http.StatusNotFound},
		"run get 404":     {workflowID: "wf_123", runStatus: http.StatusNotFound},
		"run get 500":     {workflowID: "wf_123", runStatus: http.StatusInternalServerError},
		"run has no workflow id": {
			workflowID: "",
			blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "block_raw", "")},
		},
		"alias matches nothing": {
			workflowID: "wf_123",
			blocks:     []retab.WorkflowBlock{aliasBlock("block_other", "something_else", "")},
		},
	}
	for name, stub := range cases {
		t.Run(name, func(t *testing.T) {
			newBlockAliasEnv(t, stub)
			setAliasFlag(t, workflowsBlocksExecutionsCreateCmd, "block-id", "block_raw")

			if err, _ := runAliasCmd(t, workflowsBlocksExecutionsCreateCmd, []string{"run_123"}); err != nil {
				t.Fatalf("executions create: %v", err)
			}
			if stub.lastAttempt() != "block_raw" {
				t.Fatalf("block_id sent = %q, want the raw block_raw forwarded", stub.lastAttempt())
			}
		})
	}
}

// When resolution cannot help, the error the user sees must be the one THEIR id
// produced — not one naming an id they never typed.
func TestBlocksExecutionsCreateKeepsTheOriginalErrorWhenResolutionCannotHelp(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_other", "unrelated", "")},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsBlocksExecutionsCreateCmd, "block-id", "typo_block")

	err, stderr := runAliasCmd(t, workflowsBlocksExecutionsCreateCmd, []string{"run_123"})
	if err == nil {
		t.Fatal("expected the server's rejection to surface")
	}
	if !strings.Contains(stderr, "Step data not found") {
		t.Fatalf("stderr = %q, want the server's own message", stderr)
	}
	if len(stub.attempts) != 1 || stub.attempts[0] != "typo_block" {
		t.Fatalf("attempts = %v, want only the id the user typed", stub.attempts)
	}
}

// A blocks endpoint that hands back a cursor on every page would send AutoPaging
// into an endless walk. Resolution is only a retry hint, so it must give up on
// its own deadline rather than hanging the command.
func TestBlockAliasLookupCannotHangForever(t *testing.T) {
	prev := workflowBlockAliasLookupTimeout
	workflowBlockAliasLookupTimeout = 250 * time.Millisecond
	t.Cleanup(func() { workflowBlockAliasLookupTimeout = prev })

	stub := &blockAliasServer{
		workflowID:    "wf_123",
		acceptOnly:    "block_runtime",
		endlessPaging: true,
		blocks:        []retab.WorkflowBlock{aliasBlock("block_runtime", "s31g_fn", "fn")},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsBlocksExecutionsCreateCmd, "block-id", "s31g_fn")

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = runAliasCmd(t, workflowsBlocksExecutionsCreateCmd, []string{"run_123"})
	}()
	select {
	case <-done:
		// Whatever the verdict, it came back — that is the assertion.
	case <-time.After(20 * time.Second):
		t.Fatalf("alias lookup hung; %d block pages fetched", stub.blocksCalls)
	}
}

func TestBlocksExecutionsListResolvesSpecAlias(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "s31g_fn", "fn")},
		payload:    map[string]any{"data": []any{}, "list_metadata": map[string]any{}},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsBlocksExecutionsListCmd, "block-id", "s31g_fn")

	if err, _ := runAliasCmd(t, workflowsBlocksExecutionsListCmd, []string{"run_123"}); err != nil {
		t.Fatalf("executions list: %v", err)
	}
	if stub.lastAttempt() != "block_runtime" {
		t.Fatalf("block_id sent = %q, want block_runtime", stub.lastAttempt())
	}
}

// `runs export --block-id <alias>` 404'd on a block the workflow plainly has, and
// on an export an unhelpful 404 is what stands between the user and their data.
func TestRunsExportResolvesSpecAlias(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "s31a_ex", "ex")},
		payload:    map[string]any{"csv_data": "filename\n", "rows": 0, "columns": 1},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsRunsExportCmd, "block-id", "s31a_ex")

	if err, _ := runAliasCmd(t, workflowsRunsExportCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("runs export: %v", err)
	}
	if stub.lastAttempt() != "block_runtime" {
		t.Fatalf("block_id sent = %q, want block_runtime", stub.lastAttempt())
	}
	if len(stub.attempts) != 2 || stub.attempts[0] != "s31a_ex" {
		t.Fatalf("attempts = %v, want the alias then the runtime id", stub.attempts)
	}
}

// Export resolves against the workflow id it already has, so it needs no run
// lookup at all.
func TestRunsExportResolvesWithoutFetchingARun(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "s31a_ex", "ex")},
		payload:    map[string]any{"csv_data": "filename\n", "rows": 0, "columns": 1},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsRunsExportCmd, "block-id", "s31a_ex")

	if err, _ := runAliasCmd(t, workflowsRunsExportCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("runs export: %v", err)
	}
	if stub.runCalls != 0 {
		t.Fatalf("export made %d run lookups, want 0", stub.runCalls)
	}
}

func TestRunsExportWithARuntimeIDCostsOneRequest(t *testing.T) {
	stub := &blockAliasServer{
		workflowID: "wf_123",
		acceptOnly: "block_runtime",
		blocks:     []retab.WorkflowBlock{aliasBlock("block_runtime", "s31a_ex", "ex")},
		payload:    map[string]any{"csv_data": "filename\n", "rows": 0, "columns": 1},
	}
	newBlockAliasEnv(t, stub)
	setAliasFlag(t, workflowsRunsExportCmd, "block-id", "block_runtime")

	if err, _ := runAliasCmd(t, workflowsRunsExportCmd, []string{"wf_123"}); err != nil {
		t.Fatalf("runs export: %v", err)
	}
	if len(stub.attempts) != 1 || stub.blocksCalls != 0 {
		t.Fatalf("attempts=%v blocksCalls=%d, want one attempt and no lookup",
			stub.attempts, stub.blocksCalls)
	}
}

// ---------------------------------------------------------------------------
// the publish-history id hint
// ---------------------------------------------------------------------------

// `workflows versions list` rows carry both a `wph_` publish-history id and the
// `ver_` the version is actually addressed by. Passing the row's own `id` — the
// obvious field — used to earn a bare "Workflow version was not found", which
// reads like the version is gone rather than like the wrong field was copied.
func TestRejectPublishHistoryIDAsVersion(t *testing.T) {
	err := rejectPublishHistoryIDAsVersion("wph_OUamppx7Jy45Cd04eMDpc")
	if err == nil {
		t.Fatal("a wph_ publish-history id was accepted as --version")
	}
	for _, want := range []string{"workflow_version_id", "ver_", "versions list"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %q", err.Error(), want)
		}
	}
}

func TestRejectPublishHistoryIDAsVersionAcceptsEverythingElse(t *testing.T) {
	valid := []string{
		"",
		"   ",
		"draft",
		"ver_jbxlBITBmdxVUhTlU1R8Slds0fFXiv5d",
		// Not a prefix match on a longer word.
		"wphony_id",
		// A version id that merely CONTAINS the prefix is fine.
		"ver_wph_something",
	}
	for _, version := range valid {
		if err := rejectPublishHistoryIDAsVersion(version); err != nil {
			t.Fatalf("rejected a valid --version %q: %v", version, err)
		}
	}
}

// Leading/trailing whitespace must not smuggle a wph_ id past the check.
func TestRejectPublishHistoryIDAsVersionTrimsWhitespace(t *testing.T) {
	if err := rejectPublishHistoryIDAsVersion("  wph_abc  "); err == nil {
		t.Fatal("a padded wph_ id slipped through")
	}
}
