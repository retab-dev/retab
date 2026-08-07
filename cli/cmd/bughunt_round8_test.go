package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestSafeDownloadNameRejectsWindowsReservedNames pins the round-8 fix: a
// server-recorded filename that is a Windows reserved device name (CON, NUL,
// COM1, ... — reserved even with an extension) must not become the download
// destination. On Windows the OS resolves such a base name to a device, so
// streamDownloadToFile's rename would silently discard the downloaded bytes.
// safeDownloadName returns "" for these so the caller falls back to the file id.
func TestSafeDownloadNameRejectsWindowsReservedNames(t *testing.T) {
	reserved := []string{
		"NUL", "nul", "Nul",
		"CON", "con",
		"PRN", "AUX",
		"COM1", "com9", "LPT1", "lpt9",
		"nul.pdf", "CON.txt", "com1.tar.gz", "LpT3.json",
		"NUL ", "nul .pdf", // Windows strips the trailing space -> NUL device
	}
	for _, name := range reserved {
		if got := safeDownloadName(name); got != "" {
			t.Errorf("safeDownloadName(%q) = %q, want \"\" (reserved device name)", name, got)
		}
	}
	// Names that merely CONTAIN a device token but aren't the stem stay valid.
	safe := map[string]string{
		"NULL.pdf":       "NULL.pdf",       // stem "NULL" != "NUL"
		"CONTRACT.pdf":   "CONTRACT.pdf",   // stem "CONTRACT" != "CON"
		"COM10.log":      "COM10.log",      // only COM1..COM9 are reserved
		"report-nul.pdf": "report-nul.pdf", // "nul" is not the stem
		"invoice.pdf":    "invoice.pdf",
	}
	for in, want := range safe {
		if got := safeDownloadName(in); got != want {
			t.Errorf("safeDownloadName(%q) = %q, want %q (not reserved)", in, got, want)
		}
	}
}

// TestRefreshAccessTokenPreservesOrganizationID pins the round-8 fix: when the
// WorkOS refresh response omits organization_id / token_type (both omitempty and
// not guaranteed on a refresh_token grant), the refreshed token bag must keep the
// caller's existing values rather than blanking them — otherwise every
// transparent refresh silently wipes the org id `auth status` reports. Mirrors
// the long-standing refresh_token preservation.
func TestRefreshAccessTokenPreservesOrganizationID(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Deliberately omit organization_id AND token_type from the response.
		_, _ = w.Write([]byte(`{
			"access_token": "new_access",
			"refresh_token": "new_refresh",
			"expires_in": 600
		}`))
	}))
	defer srv.Close()
	withTrustingTokenClient(t, srv.Client())

	tok := &oauthTokens{
		AccessToken:      "old_access",
		RefreshToken:     "old_refresh",
		TokenType:        "Bearer",
		OrganizationID:   "org_keep_me",
		ClientID:         "c",
		WorkosAPIBaseURL: srv.URL,
		ExpiresAt:        time.Now(),
	}
	got, err := refreshAccessToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got.OrganizationID != "org_keep_me" {
		t.Errorf("OrganizationID = %q, want it preserved as %q", got.OrganizationID, "org_keep_me")
	}
	if got.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want it preserved as %q", got.TokenType, "Bearer")
	}
	// A response that DOES carry an org id still wins over the preserved one.
	srv2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "a2",
			"refresh_token": "r2",
			"token_type": "Bearer",
			"organization_id": "org_from_server",
			"expires_in": 600
		}`))
	}))
	defer srv2.Close()
	withTrustingTokenClient(t, srv2.Client())
	tok.WorkosAPIBaseURL = srv2.URL
	got2, err := refreshAccessToken(context.Background(), tok)
	if err != nil {
		t.Fatalf("refresh 2: %v", err)
	}
	if got2.OrganizationID != "org_from_server" {
		t.Errorf("OrganizationID = %q, want server value %q", got2.OrganizationID, "org_from_server")
	}
}

// TestCSVTimestampColumnsStayFaithful pins the round-8 fix: an IsTimestamp
// column renders canonical second-precision UTC in the TABLE grid but keeps the
// raw, full-precision value in CSV (which is a re-importable data export). The
// table renderer applies normalizeTimestampCell via IsTimestamp; the CSV path
// (writeCSV) never does.
func TestCSVTimestampColumnsStayFaithful(t *testing.T) {
	col := TableColumn{
		Header:      "CREATED_AT",
		Extract:     func(row any) string { return row.(map[string]any)["created_at"].(string) },
		IsTimestamp: true,
	}
	rows := []any{
		map[string]any{"created_at": "2026-05-15T14:21:10.389000Z"},
		map[string]any{"created_at": "2026-05-15T14:21:10+02:00"},
	}

	var csvBuf bytes.Buffer
	if err := writeCSV(&csvBuf, rows, []TableColumn{col}); err != nil {
		t.Fatalf("writeCSV: %v", err)
	}
	csv := csvBuf.String()
	// CSV must keep sub-second precision and the original offset verbatim.
	if !strings.Contains(csv, "2026-05-15T14:21:10.389000Z") {
		t.Errorf("CSV dropped sub-second precision:\n%s", csv)
	}
	if !strings.Contains(csv, "2026-05-15T14:21:10+02:00") {
		t.Errorf("CSV rewrote the timezone offset:\n%s", csv)
	}

	var tblBuf bytes.Buffer
	if err := renderAutoTableWithEmptyHint(&tblBuf, rows, []TableColumn{col}, nil); err != nil {
		t.Fatalf("renderAutoTable: %v", err)
	}
	tbl := tblBuf.String()
	// Table grid canonicalizes: microseconds dropped, offset shifted to UTC.
	if strings.Contains(tbl, ".389000") {
		t.Errorf("table did not canonicalize the timestamp:\n%s", tbl)
	}
	if !strings.Contains(tbl, "2026-05-15T14:21:10Z") {
		t.Errorf("table missing canonical created_at:\n%s", tbl)
	}
	// The +02:00 instant canonicalizes to 12:21:10Z in the grid.
	if !strings.Contains(tbl, "2026-05-15T12:21:10Z") {
		t.Errorf("table did not shift offset to UTC:\n%s", tbl)
	}
}
