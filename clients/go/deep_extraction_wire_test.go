package retab

// Wire-shape tests for the deep_extraction request field.
//
// The generated create methods do NOT marshal their params struct directly —
// they copy selected fields into a private createWireBody. That copy is the
// failure mode worth testing: a field can exist on the params struct, be fully
// documented, compile, and still never reach the server because the generator
// (or a hand overlay) left it out of the wire body. The symptom is a request
// that looks correct in user code and silently runs the default extraction.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func deepExtractionTestDocument() MIMEData {
	return MIMEData{Filename: "ledger.pdf", Content: "data", MIMEType: "application/pdf"}
}

func deepExtractionTestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"rows": map[string]interface{}{"type": "array"},
		},
	}
}

// captureExtractionBody issues one create and returns the raw JSON body the
// client actually put on the wire. Raw bytes, not a decoded struct: the point
// is to observe key presence/absence exactly as the server would.
func captureExtractionBody(t *testing.T, params *ExtractionsCreateParams) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "extr_test",
			"file":        map[string]interface{}{"id": "file_1", "filename": "ledger.pdf", "mime_type": "application/pdf"},
			"model":       "retab-large",
			"json_schema": deepExtractionTestSchema(),
			"output":      map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if _, err := client.Extractions.Create(context.Background(), params); err != nil {
		t.Fatalf("Extractions.Create: %v", err)
	}
	if body == nil {
		t.Fatal("no request body captured")
	}
	return body
}

// TestExtractionsCreateSendsDeepExtraction is the assertion that catches the
// wire-body copy omission: params.DeepExtraction must actually appear in the
// serialized request.
func TestExtractionsCreateSendsDeepExtraction(t *testing.T) {
	body := captureExtractionBody(t, &ExtractionsCreateParams{
		Document:       deepExtractionTestDocument(),
		JSONSchema:     deepExtractionTestSchema(),
		Model:          ptrTo("retab-large"),
		DeepExtraction: ptrTo(true),
	})

	value, present := body["deep_extraction"]
	if !present {
		t.Fatalf("deep_extraction missing from the request body: %#v", body)
	}
	if value != true {
		t.Fatalf("deep_extraction = %#v, want true", value)
	}
}

// TestExtractionsCreateOmitsDeepExtractionWhenUnset pins the pointer semantics.
// A nil DeepExtraction must produce NO key — not false — so every existing
// caller's request body is byte-identical to what it was before the field
// existed, and the server's "absent means default" branch is the one taken.
func TestExtractionsCreateOmitsDeepExtractionWhenUnset(t *testing.T) {
	body := captureExtractionBody(t, &ExtractionsCreateParams{
		Document:   deepExtractionTestDocument(),
		JSONSchema: deepExtractionTestSchema(),
		Model:      ptrTo("retab-large"),
	})

	if value, present := body["deep_extraction"]; present {
		t.Fatalf("deep_extraction must be omitted when unset, got %#v", value)
	}
}

// TestExtractionsCreateSendsExplicitFalseDeepExtraction pins the third state.
// A caller who sets the pointer to false is deliberately opting OUT (for
// example overriding a wrapper's default), which is distinguishable from
// "unset" only if the key is actually transmitted.
func TestExtractionsCreateSendsExplicitFalseDeepExtraction(t *testing.T) {
	body := captureExtractionBody(t, &ExtractionsCreateParams{
		Document:       deepExtractionTestDocument(),
		JSONSchema:     deepExtractionTestSchema(),
		Model:          ptrTo("retab-large"),
		DeepExtraction: ptrTo(false),
	})

	value, present := body["deep_extraction"]
	if !present {
		t.Fatalf("an explicit false must still be sent: %#v", body)
	}
	if value != false {
		t.Fatalf("deep_extraction = %#v, want false", value)
	}
}

// TestExtractionsCreateStreamSendsDeepExtraction covers the streaming create,
// which has its OWN params struct and its own wire-body copy. The two are
// generated independently, so a field wired into only one is a real hole — and
// the server does dispatch the conversational strategy on its streaming path.
func TestExtractionsCreateStreamSendsDeepExtraction(t *testing.T) {
	var body map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/stream+json")
		_, _ = w.Write([]byte("{\"id\":\"extr_test\",\"output\":{}}\n"))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	err := client.Extractions.CreateStream(context.Background(), &ExtractionsCreateStreamParams{
		Document:       deepExtractionTestDocument(),
		JSONSchema:     deepExtractionTestSchema(),
		Model:          ptrTo("retab-large"),
		DeepExtraction: ptrTo(true),
	})
	if err != nil {
		t.Fatalf("Extractions.CreateStream: %v", err)
	}

	if body == nil {
		t.Fatal("no request body captured")
	}
	if body["deep_extraction"] != true {
		t.Fatalf("stream deep_extraction = %#v, want true", body["deep_extraction"])
	}
}

// TestDeepExtractionUsesTheServersFieldName pins the JSON tag itself. Go's
// field name is DeepExtraction and the server reads deep_extraction; a
// camelCase or renamed tag would serialize a key the API ignores, producing a
// silently shallow extraction rather than an error.
func TestDeepExtractionUsesTheServersFieldName(t *testing.T) {
	encoded, err := json.Marshal(&ExtractionsCreateParams{DeepExtraction: ptrTo(true)})
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if decoded["deep_extraction"] != true {
		t.Fatalf("params must serialize deep_extraction (snake_case), got %s", encoded)
	}
}
