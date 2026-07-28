package cmd

import "testing"

// The TRIGGERED_BY column has to compress a whole credential into one short
// cell. These tests pin the precedence it uses and, more importantly, that it
// never falls back to something meaningless when a better handle exists — a
// Mongo ObjectID or a raw WorkOS user id in that column is what pushed people to
// ask "who is this?" in the first place.

func triggeredByCell(t *testing.T, trigger usagePrimitiveTriggeredBy) string {
	t.Helper()
	return usagePrimitiveTriggeredByCell(usagePrimitiveRecord{TriggeredBy: &trigger})
}

// A person's email is far more useful than their WorkOS id, so it wins whenever
// the API resolved one.
func TestTriggeredByCellPrefersTheEmailOverTheUserID(t *testing.T) {
	got := triggeredByCell(t, usagePrimitiveTriggeredBy{
		AuthMethod: "session_token",
		UserID:     "user_01H8XYZ",
		UserEmail:  "ada@retab.com",
	})
	if got != "session_token:ada@retab.com" {
		t.Fatalf("cell = %q, want the email", got)
	}
}

// With no resolvable email (WorkOS down, or an unconfigured deployment) the cell
// still identifies the caller rather than going blank.
func TestTriggeredByCellFallsBackToTheUserID(t *testing.T) {
	got := triggeredByCell(t, usagePrimitiveTriggeredBy{
		AuthMethod: "session_token",
		UserID:     "user_01H8XYZ",
	})
	if got != "session_token:user_01H8XYZ" {
		t.Fatalf("cell = %q, want the user id fallback", got)
	}
}

// An API key's own display name is the most meaningful handle for a key caller,
// and outranks everything else — a key has no email of its own.
func TestTriggeredByCellPrefersTheKeyNameForAPIKeys(t *testing.T) {
	got := triggeredByCell(t, usagePrimitiveTriggeredBy{
		AuthMethod: "api_key",
		APIKeyID:   "665f00aa11bb22cc33dd44ee",
		KeyPrefix:  "sk_prod",
		KeyName:    "backend-ingest",
	})
	if got != "api_key:backend-ingest" {
		t.Fatalf("cell = %q, want the key name", got)
	}
}

// A key with no display name falls back to its prefix, never to the raw
// api_keys document id.
func TestTriggeredByCellFallsBackToTheKeyPrefixNotTheObjectID(t *testing.T) {
	got := triggeredByCell(t, usagePrimitiveTriggeredBy{
		AuthMethod: "api_key",
		APIKeyID:   "665f00aa11bb22cc33dd44ee",
		KeyPrefix:  "sk_prod",
	})
	if got != "api_key:sk_prod" {
		t.Fatalf("cell = %q, want the key prefix", got)
	}
	if got == "api_key:665f00aa11bb22cc33dd44ee" {
		t.Fatal("the raw api_keys document id must never be the display handle")
	}
}

// A master-key execution has no person and no key name; the auth method alone is
// the honest answer.
func TestTriggeredByCellShowsTheBareAuthMethodWhenThereIsNoHandle(t *testing.T) {
	if got := triggeredByCell(t, usagePrimitiveTriggeredBy{AuthMethod: "master_key"}); got != "master_key" {
		t.Fatalf("cell = %q, want the bare auth method", got)
	}
}

// An unattributed row (triggered_by null) renders empty rather than inventing a
// label.
func TestTriggeredByCellIsEmptyWithoutProvenance(t *testing.T) {
	if got := usagePrimitiveTriggeredByCell(usagePrimitiveRecord{}); got != "" {
		t.Fatalf("cell = %q, want empty for an unattributed row", got)
	}
}

// The cell is a summary, not the record: --output json must still carry the full
// provenance object, including the email, so scripts are never forced to parse
// the table.
func TestTriggeredByJSONRoundTripKeepsTheEmail(t *testing.T) {
	trigger := usagePrimitiveTriggeredBy{
		AuthMethod: "session_token",
		UserID:     "user_01H8XYZ",
		UserEmail:  "ada@retab.com",
	}
	if trigger.UserEmail != "ada@retab.com" {
		t.Fatalf("user_email lost: %+v", trigger)
	}
	if cell := triggeredByCell(t, trigger); cell == trigger.UserID {
		t.Fatal("the summary cell must not be the bare user id when an email exists")
	}
}
