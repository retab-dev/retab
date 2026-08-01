package cmd

// The /v1/usage list/export DTOs hold a "null when empty, never omit"
// serialization contract: every optional field serializes as an explicit JSON
// `null` when empty rather than being dropped, so a consumer sees the SAME key
// set on every row. The server enforces it with its own reflection drift-guard
// (services/main-api-go/internal/api/routes/v1/usage/usage_null_contract_test.go).
//
// The CLI has to enforce it independently, because `--output json` does NOT pass
// the response body through: RenderList -> renderJSON re-marshals the decoded
// mirror struct. A single `,omitempty` on a mirror field therefore silently
// breaks the contract downstream of an endpoint that carefully preserves it —
// which is exactly what happened: `usage primitives --output json` returned NINE
// different row shapes for one page of staging data (block_id / run_id /
// workflow_id / project_id / duration_ms / metadata vanishing per row) where the
// server had returned one stable 18-key shape, so a script reading
// row["duration_ms"] worked on completed rows and KeyError'd on running ones.
//
// These tests prove it two ways: a drift-guard that rejects `,omitempty` on any
// usage mirror DTO, and marshal tests asserting a zero-valued record emits every
// key with a null value.

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// usageMirrorDTOs is every struct that models a /v1/usage row (or a nested
// object inside one). Add new usage DTOs here.
func usageMirrorDTOs() []any {
	return []any{
		usageRunRecord{},
		usageRunTriggeredBy{},
		usageBlockRecord{},
		usagePrimitiveRecord{},
		usagePrimitiveTriggeredBy{},
		usagePrimitiveDocumentEl{},
	}
}

// TestUsageDTOsNeverUseOmitempty is the drift-guard. `,omitempty` on a usage
// mirror field drops the key from `--output json` whenever the server sent null,
// producing a row-dependent key set.
func TestUsageDTOsNeverUseOmitempty(t *testing.T) {
	for _, dto := range usageMirrorDTOs() {
		rt := reflect.TypeOf(dto)
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			tag := f.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			for _, opt := range strings.Split(tag, ",")[1:] {
				if opt == "omitempty" {
					t.Errorf("%s.%s has `,omitempty`: a null from the server would drop the key "+
						"from --output json, breaking the stable key set the usage API guarantees",
						rt.Name(), f.Name)
				}
			}
		}
	}
}

// TestUsageDTOsMarshalEmptyAsExplicitNull proves the consequence the drift-guard
// protects: a zero-valued record still emits every optional key, as null.
func TestUsageDTOsMarshalEmptyAsExplicitNull(t *testing.T) {
	for _, dto := range usageMirrorDTOs() {
		rt := reflect.TypeOf(dto)
		raw, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("%s: marshal: %v", rt.Name(), err)
		}
		var got map[string]json.RawMessage
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("%s: unmarshal: %v", rt.Name(), err)
		}
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				continue
			}
			value, ok := got[name]
			if !ok {
				t.Errorf("%s: key %q is missing from an empty record; it must be present as null",
					rt.Name(), name)
				continue
			}
			// Only the nullable kinds are asserted to be null — a plain string /
			// int field legitimately marshals as "" / 0.
			switch f.Type.Kind() {
			case reflect.Pointer, reflect.Map, reflect.Slice:
				if string(value) != "null" {
					t.Errorf("%s: key %q on an empty record is %s, want null",
						rt.Name(), name, value)
				}
			}
		}
	}
}

// TestUsagePrimitiveRowShapeIsIndependentOfNullness is the end-to-end statement
// of the bug: a fully-populated row and a row whose optional fields are all null
// must serialize to the SAME key set.
func TestUsagePrimitiveRowShapeIsIndependentOfNullness(t *testing.T) {
	full := usagePrimitiveRecord{
		PrimitiveExecutionID: "pexec_1",
		Operation:            "extraction",
		EnvironmentID:        strPtr("env_1"),
		WorkflowID:           strPtr("wf_1"),
		RunID:                strPtr("run_1"),
		ProjectID:            strPtr("proj_1"),
		BlockID:              strPtr("block_1"),
		Status:               "completed",
		ResourceKind:         strPtr("extraction"),
		Model:                strPtr("retab-micro"),
		CreatedAt:            strPtr("2026-07-31T18:51:48Z"),
		CompletedAt:          strPtr("2026-07-31T18:51:50Z"),
		DurationMs:           ptr(int64(2261)),
		PageCount:            3,
		Credits:              0.6,
		Documents:            []usagePrimitiveDocumentEl{{FileID: strPtr("file_1"), Filename: strPtr("invoice.pdf")}},
		Metadata:             map[string]string{"tenant": "acme"},
		TriggeredBy:          &usagePrimitiveTriggeredBy{AuthMethod: strPtr("api_key")},
	}
	// The still-running / standalone shape: every optional field null.
	sparse := usagePrimitiveRecord{
		PrimitiveExecutionID: "pexec_2",
		Operation:            "extraction",
		Status:               "running",
	}
	keys := func(v usagePrimitiveRecord) []string {
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		out := make([]string, 0, len(m))
		for k := range m {
			out = append(out, k)
		}
		return out
	}
	fullKeys, sparseKeys := keys(full), keys(sparse)
	if len(fullKeys) != len(sparseKeys) {
		t.Fatalf("row shape depends on nullness: populated row has %d keys %v, "+
			"all-null row has %d keys %v", len(fullKeys), fullKeys, len(sparseKeys), sparseKeys)
	}
	set := make(map[string]bool, len(fullKeys))
	for _, k := range fullKeys {
		set[k] = true
	}
	for _, k := range sparseKeys {
		if !set[k] {
			t.Fatalf("all-null row has key %q that the populated row does not", k)
		}
	}
}
