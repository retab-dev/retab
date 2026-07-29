package cmd

import "testing"

// The TRIGGERED_BY column on `usage runs` has to compress a whole principal into
// one short cell. It mirrors the primitives table's rule — prefer the most
// human-readable handle available — so the two usage tables read the same way.

func strPtr(s string) *string { return &s }

// A manual run shows the person's EMAIL, not their opaque user id. This is the
// whole point of the column: `user_01JHR5078RCMNTY0GN2XQ94SQS` tells an operator
// nothing, `ada@retab.com` tells them who to talk to.
func TestUsageRunTriggeredByCellPrefersEmail(t *testing.T) {
	row := usageRunRecord{
		RunID: "run_1",
		TriggeredBy: &usageRunTriggeredBy{
			UserID:    strPtr("user_01JHR5078RCMNTY0GN2XQ94SQS"),
			UserEmail: strPtr("ada@retab.com"),
		},
	}
	if got := usageRunTriggeredByCell(row); got != "ada@retab.com" {
		t.Fatalf("cell = %q, want ada@retab.com", got)
	}
}

// When WorkOS could not resolve an email (an outage, or a deleted user), the raw
// user id is still better than an empty cell — the operator can at least act on
// it.
func TestUsageRunTriggeredByCellFallsBackToUserID(t *testing.T) {
	row := usageRunRecord{
		RunID:       "run_1",
		TriggeredBy: &usageRunTriggeredBy{UserID: strPtr("user_ada")},
	}
	if got := usageRunTriggeredByCell(row); got != "user_ada" {
		t.Fatalf("cell = %q, want user_ada", got)
	}
}

// An api-triggered run has no person behind it; the key id is the handle.
func TestUsageRunTriggeredByCellUsesAPIKeyID(t *testing.T) {
	row := usageRunRecord{
		RunID:       "run_1",
		TriggeredBy: &usageRunTriggeredBy{APIKeyID: strPtr("665f00aa11bb22cc33dd44ee")},
	}
	if got := usageRunTriggeredByCell(row); got != "665f00aa11bb22cc33dd44ee" {
		t.Fatalf("cell = %q, want the api key id", got)
	}
}

// The system trigger types (schedule, webhook, email, restart) carry a null
// triggered_by. The cell must render empty rather than panic on the nil pointer
// or print "<nil>".
func TestUsageRunTriggeredByCellEmptyWhenUnattributed(t *testing.T) {
	if got := usageRunTriggeredByCell(usageRunRecord{RunID: "run_sched"}); got != "" {
		t.Fatalf("cell = %q, want empty for an unattributed run", got)
	}
	// A present-but-empty object must also render empty, not "<nil>".
	row := usageRunRecord{RunID: "run_x", TriggeredBy: &usageRunTriggeredBy{}}
	if got := usageRunTriggeredByCell(row); got != "" {
		t.Fatalf("cell = %q, want empty for an empty principal", got)
	}
}

// A non-record row (the table renderer is reflection-driven and can be handed
// anything) must not panic.
func TestUsageRunTriggeredByCellIgnoresForeignRows(t *testing.T) {
	if got := usageRunTriggeredByCell("not a record"); got != "" {
		t.Fatalf("cell = %q, want empty for a foreign row type", got)
	}
}
