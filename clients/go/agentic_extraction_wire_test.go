package retab

// Wire-shape tests for the agentic_extraction request field.
//
// The generated create methods copy selected params fields into a private
// wire body; they do not marshal the params struct. A field can exist on
// ExtractionsCreateParams, compile and be documented, and still never reach
// the server because the copy left it out. The symptom is a request that looks
// right in user code and silently runs an ordinary extraction. So these decode
// the body the client actually sent.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func agenticExtractionTestDocument() MIMEData {
	return MIMEData{Filename: "statement.pdf", Content: "data", MIMEType: "application/pdf"}
}

func agenticExtractionTestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"transactions": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "object"},
			},
		},
	}
}

// captureAgenticCreateBody issues one Create and returns the raw JSON body as
// a map, so key presence and absence are observed exactly as the server sees
// them.
func captureAgenticCreateBody(t *testing.T, params *ExtractionsCreateParams) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "extr_test",
			"file":        map[string]interface{}{"id": "file_1", "filename": "statement.pdf", "mime_type": "application/pdf"},
			"model":       "retab-large",
			"json_schema": agenticExtractionTestSchema(),
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
	if path != "/v1/extractions" {
		t.Fatalf("Create posted to %q, want /v1/extractions", path)
	}
	return body
}

// captureAgenticStreamBody issues one CreateStream and returns the raw body.
func captureAgenticStreamBody(t *testing.T, params *ExtractionsCreateStreamParams) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/stream+json")
		_, _ = w.Write([]byte("{\"id\":\"extr_test\",\"output\":{}}\n"))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.Extractions.CreateStream(context.Background(), params); err != nil {
		t.Fatalf("Extractions.CreateStream: %v", err)
	}
	if body == nil {
		t.Fatal("no request body captured")
	}
	if path != "/v1/extractions/stream" {
		t.Fatalf("CreateStream posted to %q, want /v1/extractions/stream", path)
	}
	return body
}

// TestAgenticExtractionCreateSendsTrue catches the wire-body copy omission:
// params.AgenticExtraction must appear in the serialized request.
func TestAgenticExtractionCreateSendsTrue(t *testing.T) {
	body := captureAgenticCreateBody(t, &ExtractionsCreateParams{
		Document:          agenticExtractionTestDocument(),
		JSONSchema:        agenticExtractionTestSchema(),
		Model:             ptrTo("retab-large"),
		AgenticExtraction: ptrTo(true),
	})

	value, present := body["agentic_extraction"]
	if !present {
		t.Fatalf("agentic_extraction missing from the request body: %#v", body)
	}
	if value != true {
		t.Fatalf("agentic_extraction = %#v, want true", value)
	}
	// The server applies "agentic implies deep"; the client sends only what it
	// was given.
	if value, present := body["deep_extraction"]; present {
		t.Fatalf("deep_extraction must not be added by the client, got %#v", value)
	}
}

// TestAgenticExtractionCreateSendsExplicitFalse pins the third state: a
// pointer set to false is a deliberate opt-out and is transmitted.
func TestAgenticExtractionCreateSendsExplicitFalse(t *testing.T) {
	body := captureAgenticCreateBody(t, &ExtractionsCreateParams{
		Document:          agenticExtractionTestDocument(),
		JSONSchema:        agenticExtractionTestSchema(),
		Model:             ptrTo("retab-large"),
		AgenticExtraction: ptrTo(false),
	})

	value, present := body["agentic_extraction"]
	if !present {
		t.Fatalf("an explicit false must still be sent: %#v", body)
	}
	if value != false {
		t.Fatalf("agentic_extraction = %#v, want false", value)
	}
}

// TestAgenticExtractionCreateOmitsWhenUnset pins the pointer semantics: a nil
// AgenticExtraction produces no key, so every existing caller's body is
// byte-identical to what it was before the field existed.
func TestAgenticExtractionCreateOmitsWhenUnset(t *testing.T) {
	body := captureAgenticCreateBody(t, &ExtractionsCreateParams{
		Document:   agenticExtractionTestDocument(),
		JSONSchema: agenticExtractionTestSchema(),
		Model:      ptrTo("retab-large"),
	})

	if value, present := body["agentic_extraction"]; present {
		t.Fatalf("agentic_extraction must be omitted when unset, got %#v", value)
	}
}

// TestAgenticExtractionCreateStreamSendsTheField covers the streaming create,
// which has its own params struct and its own wire-body copy. The server
// answers 422 for agentic extraction on the streaming route; the client must
// still send the field so the caller gets that error instead of a silently
// ordinary streamed extraction.
func TestAgenticExtractionCreateStreamSendsTheField(t *testing.T) {
	for _, want := range []bool{true, false} {
		body := captureAgenticStreamBody(t, &ExtractionsCreateStreamParams{
			Document:          agenticExtractionTestDocument(),
			JSONSchema:        agenticExtractionTestSchema(),
			Model:             ptrTo("retab-large"),
			AgenticExtraction: ptrTo(want),
		})
		value, present := body["agentic_extraction"]
		if !present {
			t.Fatalf("stream agentic_extraction=%v missing from the body: %#v", want, body)
		}
		if value != want {
			t.Fatalf("stream agentic_extraction = %#v, want %v", value, want)
		}
	}
}

// TestAgenticExtractionCreateStreamOmitsWhenUnset is the stream twin of the
// unset case.
func TestAgenticExtractionCreateStreamOmitsWhenUnset(t *testing.T) {
	body := captureAgenticStreamBody(t, &ExtractionsCreateStreamParams{
		Document:   agenticExtractionTestDocument(),
		JSONSchema: agenticExtractionTestSchema(),
		Model:      ptrTo("retab-large"),
	})

	if value, present := body["agentic_extraction"]; present {
		t.Fatalf("stream agentic_extraction must be omitted when unset, got %#v", value)
	}
}

// TestAgenticExtractionUsesTheServersFieldName pins the JSON tag on both
// params structs: a camelCase or renamed tag would send a key the API ignores.
func TestAgenticExtractionUsesTheServersFieldName(t *testing.T) {
	for name, params := range map[string]interface{}{
		"create":        &ExtractionsCreateParams{AgenticExtraction: ptrTo(true)},
		"create_stream": &ExtractionsCreateStreamParams{AgenticExtraction: ptrTo(true)},
	} {
		encoded, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("%s: marshal params: %v", name, err)
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("%s: unmarshal params: %v", name, err)
		}
		if decoded["agentic_extraction"] != true {
			t.Fatalf("%s params must serialize agentic_extraction (snake_case), got %s", name, encoded)
		}
	}
}
