package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestEndpointCLIExposesCRUDCommands(t *testing.T) {
	for _, name := range []string{"list", "get", "create", "update", "delete"} {
		found := false
		for _, command := range endpointsCmd.Commands() {
			if command.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("retab endpoints is missing %q", name)
		}
	}
	if endpointsCmd.Name() != "endpoints" || len(endpointsCmd.Aliases) != 1 || endpointsCmd.Aliases[0] != "endpoint" {
		t.Fatalf("endpoint command aliases = %#v", endpointsCmd.Aliases)
	}
}

func TestEndpointCLICreateAndListUseEndpointAutomationRoutes(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "rt_test_key")

	var createBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == endpointAutomationPath:
			if err := json.NewDecoder(r.Body).Decode(&createBody); err != nil {
				t.Errorf("decode create body: %v", err)
			}
			_, _ = w.Write([]byte(`{"id":"endp_1","object":"automation.endpoint","name":"Invoices","processor_id":"proc_1","webhook_url":"https://example.test/hook","webhook_headers":{"Authorization":"Bearer must-not-print"},"need_validation":true}`))
		case r.Method == http.MethodGet && r.URL.Path == endpointAutomationPath:
			if r.URL.Query().Get("processor_id") != "proc_1" || r.URL.Query().Get("limit") != "25" {
				t.Errorf("list query = %q", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"endp_1","object":"automation.endpoint","name":"Invoices","processor_id":"proc_1","webhook_url":"https://example.test/hook","webhook_headers":{"Authorization":"Bearer must-not-print"}}],"list_metadata":{"before":null,"after":null}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	for flag, value := range map[string]string{
		"name": "Invoices", "processor-id": "proc_1", "webhook-url": "https://example.test/hook",
		"need-validation": "true",
	} {
		if err := endpointsCreateCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("set create --%s: %v", flag, err)
		}
	}
	t.Cleanup(func() {
		for _, flag := range []string{"name", "processor-id", "webhook-url"} {
			_ = endpointsCreateCmd.Flags().Set(flag, "")
			if value := endpointsCreateCmd.Flags().Lookup(flag); value != nil {
				value.Changed = false
			}
		}
		_ = endpointsCreateCmd.Flags().Set("need-validation", "false")
		endpointsCreateCmd.Flags().Lookup("need-validation").Changed = false
	})
	stdout, err := captureStdAndRun(t, func() error { return endpointsCreateCmd.RunE(endpointsCreateCmd, nil) })
	if err != nil {
		t.Fatalf("endpoints create: %v", err)
	}
	if !strings.Contains(stdout, `"id": "endp_1"`) {
		t.Fatalf("create output = %s", stdout)
	}
	if strings.Contains(stdout, "must-not-print") || strings.Contains(stdout, "webhook_headers") {
		t.Fatalf("create output exposed webhook credentials: %s", stdout)
	}
	if createBody["name"] != "Invoices" || createBody["processor_id"] != "proc_1" || createBody["need_validation"] != true {
		t.Fatalf("create body = %#v", createBody)
	}
	if err := endpointsListCmd.Flags().Set("processor-id", "proc_1"); err != nil {
		t.Fatal(err)
	}
	if err := endpointsListCmd.Flags().Set("limit", "25"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = endpointsListCmd.Flags().Set("processor-id", "")
		_ = endpointsListCmd.Flags().Set("limit", "0")
	})
	stdout, err = captureStdAndRun(t, func() error { return endpointsListCmd.RunE(endpointsListCmd, nil) })
	if err != nil {
		t.Fatalf("endpoints list: %v", err)
	}
	if !strings.Contains(stdout, "endp_1") || !strings.Contains(stdout, "Invoices") {
		t.Fatalf("list output = %s", stdout)
	}
	if strings.Contains(stdout, "must-not-print") || strings.Contains(stdout, "webhook_headers") {
		t.Fatalf("list output exposed webhook credentials: %s", stdout)
	}
}

func TestEndpointCLIGetUpdateAndDeleteUseEndpointAutomationRoutes(t *testing.T) {
	resetEnvironmentCommandPersistentFlags(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("RETAB_API_KEY", "rt_test_key")

	var methods []string
	var updateBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != endpointAutomationPath+"/endp_1" {
			http.NotFound(w, r)
			return
		}
		methods = append(methods, r.Method)
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":"endp_1","object":"automation.endpoint","name":"Invoices","processor_id":"proc_1","webhook_url":"https://example.test/hook"}`))
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&updateBody); err != nil {
				t.Errorf("decode update body: %v", err)
			}
			_, _ = w.Write([]byte(`{"id":"endp_1","object":"automation.endpoint","name":"Receipts","processor_id":"proc_1","webhook_url":"https://example.test/hook"}`))
		case http.MethodDelete:
			_, _ = w.Write([]byte(`{"message":"Endpoint deleted successfully"}`))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()
	t.Setenv("RETAB_API_BASE_URL", server.URL)

	if _, err := captureStdAndRun(t, func() error { return endpointsGetCmd.RunE(endpointsGetCmd, []string{"endp_1"}) }); err != nil {
		t.Fatalf("endpoints get: %v", err)
	}
	if err := endpointsUpdateCmd.Flags().Set("name", "Receipts"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = endpointsUpdateCmd.Flags().Set("name", "")
		endpointsUpdateCmd.Flags().Lookup("name").Changed = false
	})
	if _, err := captureStdAndRun(t, func() error { return endpointsUpdateCmd.RunE(endpointsUpdateCmd, []string{"endp_1"}) }); err != nil {
		t.Fatalf("endpoints update: %v", err)
	}
	if updateBody["name"] != "Receipts" || len(updateBody) != 1 {
		t.Fatalf("update body = %#v", updateBody)
	}
	if err := endpointsDeleteCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = endpointsDeleteCmd.Flags().Set("yes", "false")
		endpointsDeleteCmd.Flags().Lookup("yes").Changed = false
	})
	if _, err := captureStdAndRun(t, func() error { return endpointsDeleteCmd.RunE(endpointsDeleteCmd, []string{"endp_1"}) }); err != nil {
		t.Fatalf("endpoints delete: %v", err)
	}

	wantMethods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	if len(methods) != len(wantMethods) {
		t.Fatalf("methods = %v", methods)
	}
	for index := range wantMethods {
		if methods[index] != wantMethods[index] {
			t.Fatalf("methods = %v, want %v", methods, wantMethods)
		}
	}
}

func TestEndpointHeadersPreserveValuesAfterFirstEquals(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	command.Flags().StringArray("webhook-header", nil, "")
	if err := command.Flags().Set("webhook-header", "Authorization=Bearer value=with=equals"); err != nil {
		t.Fatal(err)
	}
	headers, err := endpointHeaders(command)
	if err != nil {
		t.Fatal(err)
	}
	if headers["Authorization"] != "Bearer value=with=equals" {
		t.Fatalf("headers = %#v", headers)
	}
}
