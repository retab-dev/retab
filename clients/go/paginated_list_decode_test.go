package retab

import (
	"encoding/json"
	"testing"
)

// PaginatedList.UnmarshalJSON must accept BOTH list response shapes the
// Retab API emits in the wild:
//
//   - the {data, list_metadata} envelope — most list endpoints
//   - a bare JSON array — e.g. GET /workflows/{id}/edges, which has no
//     pagination wrapper
//
// `retab workflows edges list` crashed against the bare-array form with
//   json: cannot unmarshal array into Go value of type retab.alias[...]
// because the decoder only handled the envelope. These tests pin both
// paths so a future refactor can't silently regress one of them.

func TestPaginatedListDecodesEnvelopeShape(t *testing.T) {
	var list PaginatedList[WorkflowEdgeDoc]
	body := `{"data":[{"id":"edge-1","source_block":"a","target_block":"b"}],
	          "list_metadata":{"before":"x","after":"y"}}`
	if err := json.Unmarshal([]byte(body), &list); err != nil {
		t.Fatalf("envelope shape failed to decode: %v", err)
	}
	if len(list.Data) != 1 || list.Data[0].ID != "edge-1" {
		t.Fatalf("data not decoded: %#v", list.Data)
	}
	if list.ListMetadata.Before != "x" || list.ListMetadata.After != "y" {
		t.Fatalf("list_metadata not decoded: %#v", list.ListMetadata)
	}
}

func TestPaginatedListDecodesBareArrayShape(t *testing.T) {
	var list PaginatedList[WorkflowEdgeDoc]
	// The exact shape GET /workflows/{id}/edges returns: a bare array,
	// no envelope. Pre-fix this was an immediate decode error.
	body := `[{"id":"edge-1","source_block":"a","target_block":"b"},
	          {"id":"edge-2","source_block":"b","target_block":"c"}]`
	if err := json.Unmarshal([]byte(body), &list); err != nil {
		t.Fatalf("bare array shape failed to decode: %v", err)
	}
	if len(list.Data) != 2 {
		t.Fatalf("expected 2 edges, got %d: %#v", len(list.Data), list.Data)
	}
	if list.Data[0].ID != "edge-1" || list.Data[1].ID != "edge-2" {
		t.Fatalf("edges decoded wrong: %#v", list.Data)
	}
	// Pagination fields stay zero-valued — a bare array carries none.
	if list.ListMetadata.Before != "" || list.ListMetadata.After != "" {
		t.Fatalf("bare array must leave list_metadata zero, got %#v", list.ListMetadata)
	}
}

// Both empty forms — `[]` and `{"data":null}` — must yield a non-nil
// empty slice so callers can range over `.Data` without a nil check.
func TestPaginatedListEmptyFormsYieldNonNilSlice(t *testing.T) {
	for _, body := range []string{`[]`, `{"data":null}`, `{}`} {
		var list PaginatedList[WorkflowEdgeDoc]
		if err := json.Unmarshal([]byte(body), &list); err != nil {
			t.Fatalf("body %q failed to decode: %v", body, err)
		}
		if list.Data == nil {
			t.Errorf("body %q: Data should be non-nil empty slice, got nil", body)
		}
		if len(list.Data) != 0 {
			t.Errorf("body %q: expected empty slice, got %#v", body, list.Data)
		}
	}
}

// Leading whitespace before the array must not defeat the shape sniff.
func TestPaginatedListDecodesBareArrayWithLeadingWhitespace(t *testing.T) {
	var list PaginatedList[WorkflowEdgeDoc]
	if err := json.Unmarshal([]byte("  \n\t [{\"id\":\"edge-1\"}]"), &list); err != nil {
		t.Fatalf("whitespace-prefixed array failed to decode: %v", err)
	}
	if len(list.Data) != 1 || list.Data[0].ID != "edge-1" {
		t.Fatalf("data not decoded: %#v", list.Data)
	}
}
