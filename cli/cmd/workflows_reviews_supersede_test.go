package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newAckFlagCommand mirrors how approve / versions create register the flag.
func newAckFlagCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "probe"}
	cmd.Flags().StringArray("acknowledge-superseded", nil, "")
	return cmd
}

// A contested review lineage is refused with a 409 whose detail ends "re-send
// this request with acknowledged_superseded_version_ids=[...]". The CLI had no
// flag carrying that field, so it printed a remediation it could not perform.
func TestAcknowledgedSupersededFlagAcceptsRepeatedValues(t *testing.T) {
	cmd := newAckFlagCommand()
	if err := cmd.Flags().Parse([]string{
		"--acknowledge-superseded", "rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA",
		"--acknowledge-superseded", "rvr_BBBBBBBBBBBBBBBBBBBBBBBBBB",
	}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := acknowledgedSupersededFlag(cmd)
	want := []string{"rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA", "rvr_BBBBBBBBBBBBBBBBBBBBBBBBBB"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// The 409 detail prints the ids as a bracketed list, so the ids a user pastes
// arrive as one comma-separated value.
func TestAcknowledgedSupersededFlagSplitsCommaSeparatedValue(t *testing.T) {
	cmd := newAckFlagCommand()
	if err := cmd.Flags().Parse([]string{
		"--acknowledge-superseded", "rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA, rvr_BBBBBBBBBBBBBBBBBBBBBBBBBB",
	}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := acknowledgedSupersededFlag(cmd); len(got) != 2 || got[1] != "rvr_BBBBBBBBBBBBBBBBBBBBBBBBBB" {
		t.Fatalf("got %v, want two trimmed ids", got)
	}
}

func TestAcknowledgedSupersededFlagEmptyWhenUnset(t *testing.T) {
	cmd := newAckFlagCommand()
	if got := acknowledgedSupersededFlag(cmd); len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

// A typo'd id would otherwise reach the server and silently fail to cover the
// version it was meant to acknowledge, re-raising the same 409.
func TestValidateAcknowledgeSupersededFlagRejectsMalformedID(t *testing.T) {
	cmd := newAckFlagCommand()
	if err := cmd.Flags().Parse([]string{"--acknowledge-superseded", "rvr_NOPE"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validateAcknowledgeSupersededFlag(cmd); err == nil {
		t.Fatal("malformed --acknowledge-superseded must be rejected locally")
	}
}

func TestValidateAcknowledgeSupersededFlagAcceptsWellFormedIDs(t *testing.T) {
	cmd := newAckFlagCommand()
	if err := cmd.Flags().Parse([]string{"--acknowledge-superseded", "rvr_AAAAAAAAAAAAAAAAAAAAAAAAAA"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validateAcknowledgeSupersededFlag(cmd); err != nil {
		t.Fatalf("well-formed id rejected: %v", err)
	}
}

// Both decision surfaces the 409 can fire on must carry the flag.
func TestApproveAndVersionCreateBothExposeAcknowledgeSuperseded(t *testing.T) {
	for _, cmd := range []*cobra.Command{workflowsReviewsApproveCmd, workflowsReviewsVersionsCreateCmd} {
		if cmd.Flags().Lookup("acknowledge-superseded") == nil {
			t.Fatalf("%q must expose --acknowledge-superseded; the 409 it can raise names that field", cmd.Use)
		}
	}
}

// A split block's stored document shape carries `partitions` (a block with no
// partition_key emits exactly "partitions": []), so the seed version a reviewer
// reads back has the key. Advertising the item as name+pages with
// additionalProperties:false told them their own snapshot was invalid.
func TestSplitReviewSchemaAdvertisesPartitions(t *testing.T) {
	schema, err := reviewSchemaForBlockType("split")
	if err != nil {
		t.Fatalf("split schema: %v", err)
	}
	encoded, err := json.Marshal(schema.Snapshot)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded struct {
		Properties struct {
			Documents struct {
				Items struct {
					Properties map[string]any `json:"properties"`
					Required   []string       `json:"required"`
				} `json:"items"`
			} `json:"documents"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	item := decoded.Properties.Documents.Items
	if _, ok := item.Properties["partitions"]; !ok {
		t.Fatalf("split document item must allow partitions; got properties %v", item.Properties)
	}
	// Optional: a reviewer who never had partitions must not be forced to invent one.
	for _, r := range item.Required {
		if r == "partitions" {
			t.Fatal("partitions must stay optional")
		}
	}
}

// An extract schema that is merely silent on additionalProperties reads as
// "extras are fine", but the server now rejects any key the block never
// declared — an approved snapshot becomes the block's output verbatim, so a
// correction must not widen the shape downstream blocks were written against.
func TestExtractReviewSchemaStatesUndeclaredFieldsAreRejected(t *testing.T) {
	base, err := reviewSchemaForBlockType("extract")
	if err != nil {
		t.Fatalf("extract schema: %v", err)
	}
	config := map[string]any{"json_schema": map[string]any{
		"type":       "object",
		"properties": map[string]any{"booking_reference": map[string]any{"type": "string"}},
	}}
	withConfig := reviewSchemaWithBlockConfig(base, config, "workflow-version")

	var stated bool
	for _, note := range withConfig.Notes {
		if strings.Contains(note, "never declared is rejected") {
			stated = true
		}
	}
	if !stated {
		t.Fatalf("schema notes must state that undeclared fields are rejected; got %#v", withConfig.Notes)
	}
}
