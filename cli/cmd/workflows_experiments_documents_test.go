//go:build !retab_oagen_cli_workflows_experiments

package cmd

import "testing"

func TestExperimentHandleInputsFromMapPreservesFileDescriptor(t *testing.T) {
	inputs, err := experimentHandleInputsFromMap(map[string]any{
		"input-file-document": map[string]any{
			"type": "file",
			"document": map[string]any{
				"id":        "file_123",
				"filename":  "invoice.jpeg",
				"mime_type": "image/jpeg",
			},
		},
	})
	if err != nil {
		t.Fatalf("experimentHandleInputsFromMap: %v", err)
	}
	got := inputs["input-file-document"]
	if got.Type() != "file" {
		t.Fatalf("input type = %q, want file", got.Type())
	}
	fileInput, err := got.AsFileHandleInput()
	if err != nil {
		t.Fatalf("AsFileHandleInput: %v", err)
	}
	if fileInput.Document.ID != "file_123" ||
		fileInput.Document.Filename != "invoice.jpeg" ||
		fileInput.Document.MIMEType != "image/jpeg" {
		t.Fatalf("document = %#v", fileInput.Document)
	}
}

func TestExperimentHandleInputsFromMapKeepsJSONDescriptors(t *testing.T) {
	inputs, err := experimentHandleInputsFromMap(map[string]any{
		"input-json-0": map[string]any{
			"type": "json",
			"data": map[string]any{"vendor": "AMNOSH"},
		},
		"legacy-json": map[string]any{"total": 12113.67},
	})
	if err != nil {
		t.Fatalf("experimentHandleInputsFromMap: %v", err)
	}
	for _, key := range []string{"input-json-0", "legacy-json"} {
		if inputs[key].Type() != "json" {
			t.Fatalf("%s type = %q, want json", key, inputs[key].Type())
		}
		jsonInput, err := inputs[key].AsJSONHandleInput()
		if err != nil {
			t.Fatalf("%s AsJSONHandleInput: %v", key, err)
		}
		if jsonInput.Data == nil {
			t.Fatalf("%s data is nil", key)
		}
	}
}

// A raw JSON document is allowed to carry a top-level "type" field whose value
// is neither "file" nor "json" (e.g. a document classified as {"type":"invoice"}).
// It must pass through as raw JSON data, not be rejected as an unsupported
// handle input type.
func TestExperimentHandleInputsFromMapAllowsRawTypeField(t *testing.T) {
	inputs, err := experimentHandleInputsFromMap(map[string]any{
		"input-json-0": map[string]any{
			"type":   "invoice",
			"amount": 100,
		},
	})
	if err != nil {
		t.Fatalf("experimentHandleInputsFromMap: %v", err)
	}
	got := inputs["input-json-0"]
	if got.Type() != "json" {
		t.Fatalf("input type = %q, want json", got.Type())
	}
	jsonInput, err := got.AsJSONHandleInput()
	if err != nil {
		t.Fatalf("AsJSONHandleInput: %v", err)
	}
	if jsonInput.Data == nil {
		t.Fatal("data is nil; raw document with a type field was dropped")
	}
	// Data round-trips through JSON, so numbers decode back as float64.
	obj, ok := (*jsonInput.Data).(map[string]any)
	if !ok || obj["type"] != "invoice" || obj["amount"] != float64(100) {
		t.Fatalf("raw data not preserved: %#v", jsonInput.Data)
	}
}

func TestExperimentHandleInputsFromMapRejectsMalformedFileDescriptor(t *testing.T) {
	_, err := experimentHandleInputsFromMap(map[string]any{
		"input-file-document": map[string]any{
			"type": "file",
			"document": map[string]any{
				"filename":  "invoice.jpeg",
				"mime_type": "image/jpeg",
			},
		},
	})
	if err == nil {
		t.Fatal("expected malformed file descriptor to fail")
	}
}
