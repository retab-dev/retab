package retab

import (
	"encoding/json"
	"testing"
)

// PaginatedList.UnmarshalJSON pins the live wire shape: every list endpoint
// returns {"data": [...], "list_metadata": {...}}. A nil `data` field is
// normalised to an empty slice so callers can range over `.Data` without a
// nil check.

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

// Null and missing `data` must yield a non-nil empty slice so callers can
// range over `.Data` unconditionally.
func TestPaginatedListEmptyFormsYieldNonNilSlice(t *testing.T) {
	for _, body := range []string{`{"data":null}`, `{}`} {
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
