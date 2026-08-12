package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

// Round-10 bug hunt. Each test pins a fix corroborated by reading the code and
// verified by independent adversarial review agents.

// Bug: the OAuth `auth login` path cleared APIKey / AccessToken /
// DefaultEnvironment / BaseURL for the incoming account but left the previous
// account's EnvironmentID / EnvironmentType in place, and passed the stale id
// in as the selection preference. When post-login environment resolution fails
// (a handled path — /v1/environments is a live network call), the new OAuth
// tokens were saved paired with the OLD account's environment. Beyond routing
// requests at an environment outside the new org, the offline production gate
// reads EnvironmentType: a stale "non_production" would silently disengage the
// confirmation prompt for what is now a production session. The fix resets both
// fields before resolution so a failure fails safe (empty type => gated), while
// still offering the old id as a preference so a re-login to the SAME account
// re-picks its environment. Mirrors `org switch`.
func TestOAuthLoginResetsStaleEnvironmentOnResolveFailure(t *testing.T) {
	isolateHome(t)
	t.Setenv("RETAB_API_BASE_URL", "")

	// One server plays discovery, the WorkOS device flow, and /v1/environments.
	// The environments call fails (500), exercising the handled resolve-failure
	// path that used to leak the previous account's environment.
	var envCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/auth/cli/config":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"client_id":           "client_test",
				"workos_api_base_url": serverURLFromRequest(r),
			})
		case r.Method == http.MethodPost && r.URL.Path == "/user_management/authorize/device":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"device_code":"dc","user_code":"WXYZ-1234","verification_uri":"https://x/device","verification_uri_complete":"https://x/device?code=WXYZ-1234","expires_in":300,"interval":1}`)
		case r.Method == http.MethodPost && r.URL.Path == "/user_management/authenticate":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"access_token":"at_new","refresh_token":"rt_new","token_type":"Bearer","expires_in":3600,"organization_id":"org_new"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/environments":
			envCalls++
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// The device/token poll runs through tokenHTTPClient; point it at the
	// plain test server. Discovery and /v1/environments use their own clients
	// (http.DefaultClient) and reach the server directly.
	withTrustingTokenClient(t, server.Client())

	// A prior account is on file, with an environment from THAT account. A
	// "non_production" type is the dangerous case: leaked into the new
	// production-bound session it would disengage the safety gate.
	if err := saveConfig(retabConfig{
		BaseURL:         server.URL,
		EnvironmentID:   "env_old_account",
		EnvironmentType: "non_production",
		OAuth: &oauthTokens{
			AccessToken:      "at_old",
			RefreshToken:     "rt_old",
			ExpiresAt:        time.Now().Add(time.Hour),
			WorkosAPIBaseURL: server.URL,
			ClientID:         "client_old",
		},
	}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	// The device flow prints instructions and "opens" a browser; stub the
	// opener so nothing shells out.
	prevOpener := browserOpener
	browserOpener = func(string) error { return nil }
	t.Cleanup(func() {
		browserOpener = prevOpener
		_ = authLoginCmd.Flags().Set("base-url", "")
		if f := authLoginCmd.Flags().Lookup("base-url"); f != nil {
			f.Changed = false
		}
	})

	_, _ = captureStd(t, func() {
		if err := runRootForTest(t, "auth", "login", "--base-url", server.URL); err != nil {
			// The login itself succeeds; the environment warning is non-fatal.
			t.Fatalf("auth login: %v", err)
		}
	})

	if envCalls == 0 {
		t.Fatal("test did not exercise the /v1/environments resolve-failure path")
	}

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.OAuth == nil || cfg.OAuth.AccessToken != "at_new" {
		t.Fatalf("new OAuth tokens were not persisted: %+v", cfg.OAuth)
	}
	if cfg.EnvironmentID != "" {
		t.Errorf("EnvironmentID = %q, want empty — the previous account's environment leaked into the new session", cfg.EnvironmentID)
	}
	if cfg.EnvironmentType != "" {
		t.Errorf("EnvironmentType = %q, want empty — a stale non_production type would disengage the production gate", cfg.EnvironmentType)
	}
}

// serverURLFromRequest reconstructs the base URL the client used to reach the
// test server, so discovery can echo it back as the WorkOS base URL.
func serverURLFromRequest(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// Bug: `--where "col in a,b,c"` / `not_in` forwarded raw strings, while every
// other operator (eq/lt/between/...) ran its value(s) through
// coerceTableScalarValue. So an `in` filter on a numeric or boolean column sent
// JSON strings that match nothing server-side — a silently-empty page. The fix
// coerces each list element, matching `between`.
func TestTableWhereInOperatorCoercesScalars(t *testing.T) {
	cases := []struct {
		name  string
		where string
		op    string
		want  []any
	}{
		{"int in", "priority in 1,2,3", "in", []any{json.Number("1"), json.Number("2"), json.Number("3")}},
		{"bool not_in", "active not-in true,false", "not_in", []any{true, false}},
		{"string in preserved", "countrycode in GB,FR", "in", []any{"GB", "FR"}},
		{"null coerced", "note in null,x", "in", []any{nil, "x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTableWhereFlag(tc.where)
			if err != nil {
				t.Fatalf("parseTableWhereFlag(%q): %v", tc.where, err)
			}
			if got["operator"] != tc.op {
				t.Fatalf("operator = %v, want %v", got["operator"], tc.op)
			}
			values, ok := got["value"].([]any)
			if !ok {
				t.Fatalf("value is %T, want []any", got["value"])
			}
			if len(values) != len(tc.want) {
				t.Fatalf("value = %#v, want %#v", values, tc.want)
			}
			for i := range values {
				if fmt.Sprintf("%v (%T)", values[i], values[i]) != fmt.Sprintf("%v (%T)", tc.want[i], tc.want[i]) {
					t.Errorf("value[%d] = %v (%T), want %v (%T)", i, values[i], values[i], tc.want[i], tc.want[i])
				}
			}
		})
	}
}

// Bug: a bare `--category invoice` (no `=`) serialized {"name":"invoice",
// "description":""} because Category.Description is `*string,omitempty` and a
// non-nil pointer to "" is NOT omitted, while the --categories-file form omits
// the key entirely for the equivalent {"name":"invoice"}. Two equivalent inputs
// must produce the same wire bytes. The fix leaves Description nil unless the
// user actually wrote `name=desc`.
func TestClassificationInlineCategoryOmitsEmptyDescription(t *testing.T) {
	t.Setenv("RETAB_API_KEY", "rt_test_key")
	t.Setenv("HOME", t.TempDir())

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/classifications" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"cls_1","status":"completed"}`)
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	mustSet := func(name, value string) {
		if err := classificationsCreateCmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	mustSet("url", "https://example.com/x.pdf")
	mustSet("model", "gpt-4o")
	mustSet("category", "invoice")        // bare: no description
	mustSet("category", "receipt=a bill") // explicit description survives
	t.Cleanup(func() {
		for _, n := range []string{"url", "model"} {
			_ = classificationsCreateCmd.Flags().Set(n, "")
			if f := classificationsCreateCmd.Flags().Lookup(n); f != nil {
				f.Changed = false
			}
		}
		if f := classificationsCreateCmd.Flags().Lookup("category"); f != nil {
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				_ = sv.Replace(nil)
			}
			f.Changed = false
		}
	})

	if _, err := captureStdAndRun(t, func() error {
		return classificationsCreateCmd.RunE(classificationsCreateCmd, nil)
	}); err != nil {
		t.Fatalf("classifications create: %v", err)
	}

	cats, ok := body["categories"].([]any)
	if !ok || len(cats) != 2 {
		t.Fatalf("categories = %#v, want 2", body["categories"])
	}
	first := cats[0].(map[string]any)
	if first["name"] != "invoice" {
		t.Fatalf("first category = %#v", first)
	}
	if _, present := first["description"]; present {
		t.Errorf("bare --category serialized description=%#v; it must be omitted like the file form", first["description"])
	}
	second := cats[1].(map[string]any)
	if second["description"] != "a bill" {
		t.Errorf("explicit description was dropped: %#v", second)
	}
}

// Bug: `files upload` derived its Content-Type from the host MIME registry
// (mime.TypeByExtension), while the parse path uses a hardcoded, host-
// independent table (extMIME). On a host lacking a registry entry (e.g. `.tif`
// on a stripped Linux/container box), an upload declared application/octet-
// stream even though the CLI's own parse table knows it is image/tiff. The fix
// consults extMIME first, so both paths agree regardless of host.
func TestUploadMIMEIsHostIndependentForKnownKinds(t *testing.T) {
	cases := map[string]string{
		"scan.tif":  "image/tiff",
		"scan.tiff": "image/tiff",
		"logo.bmp":  "image/bmp",
		"page.webp": "image/webp",
		"doc.pdf":   "application/pdf",
		"pic.png":   "image/png",
	}
	for name, want := range cases {
		if got := mimeTypeFromExtension(name); got != want {
			t.Errorf("mimeTypeFromExtension(%q) = %q, want %q (must match the parse path's extMIME table)", name, got, want)
		}
	}
	// Extensions the CLI does not itself understand still defer to the host
	// registry — a nil/empty result there is fine, we only pin the known kinds.
	if got := mimeTypeFromExtension("archive.tar.zst"); got == "image/tiff" {
		t.Errorf("unexpected coercion of unknown extension: %q", got)
	}
}
