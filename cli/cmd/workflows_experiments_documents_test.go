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
