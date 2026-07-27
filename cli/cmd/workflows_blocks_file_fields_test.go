package cmd

import (
	"strings"
	"testing"
)

// Every top-level --block-file field is read by type assertion, so a misspelled
// key or a wrong-typed value used to be dropped silently: the block was created
// with no label at position (0,0) and the command exited 0. The server already
// rejects unknown keys inside `config`; these pin the same contract on the
// block's own fields.

func validExtractBlockFile() map[string]any {
	return map[string]any{
		"type":       "extract",
		"label":      "Extract",
		"position_x": float64(420),
		"position_y": float64(180),
		"config": map[string]any{
			"model": "retab-small",
		},
	}
}

func TestParseBlockCreateRejectsMisspelledField(t *testing.T) {
	obj := validExtractBlockFile()
	delete(obj, "label")
	obj["labell"] = "typo"

	_, err := parseBlockCreate(obj)
	if err == nil {
		t.Fatal("a misspelled block-file field should not be silently dropped")
	}
	if !strings.Contains(err.Error(), "labell") {
		t.Fatalf("error should name the offending field, got %q", err)
	}
	if !strings.Contains(err.Error(), "Valid fields:") || !strings.Contains(err.Error(), "position_x") {
		t.Fatalf("error should list the valid fields, got %q", err)
	}
}

func TestParseBlockCreateRejectsWrongTypedField(t *testing.T) {
	obj := validExtractBlockFile()
	obj["position_x"] = "700"

	_, err := parseBlockCreate(obj)
	if err == nil {
		t.Fatal("a string position_x should not be silently dropped")
	}
	if !strings.Contains(err.Error(), "position_x") || !strings.Contains(err.Error(), "number") {
		t.Fatalf("error should name the field and expected kind, got %q", err)
	}
}

func TestParseBlockCreateRejectsWrongTypedConfig(t *testing.T) {
	obj := validExtractBlockFile()
	obj["config"] = "model=retab-small"

	_, err := parseBlockCreate(obj)
	if err == nil {
		t.Fatal("a non-object config should be rejected")
	}
	if !strings.Contains(err.Error(), "config") || !strings.Contains(err.Error(), "object") {
		t.Fatalf("error should name config and expected kind, got %q", err)
	}
}

func TestParseBlockCreateAcceptsValidFile(t *testing.T) {
	req, err := parseBlockCreate(validExtractBlockFile())
	if err != nil {
		t.Fatalf("valid block file should parse, got %v", err)
	}
	if req.Label == nil || *req.Label != "Extract" {
		t.Fatalf("label = %#v, want Extract", req.Label)
	}
	if req.PositionX == nil || *req.PositionX != 420 {
		t.Fatalf("position_x = %#v, want 420", req.PositionX)
	}
}

// `workflows blocks get` returns server-computed fields create does not accept;
// feeding its output straight back must keep working.
func TestParseBlockCreateIgnoresServerEchoFields(t *testing.T) {
	obj := validExtractBlockFile()
	obj["id"] = "block_abc"
	obj["workflow_id"] = "wrk_abc"
	obj["updated_at"] = "2026-07-25T10:00:00Z"
	obj["handles"] = map[string]any{"inputs": []any{}}
	obj["declarative_path"] = "loop.child"
	obj["declarative_source_block_id"] = "child"

	req, err := parseBlockCreate(obj)
	if err != nil {
		t.Fatalf("a blocks-get payload should round-trip into create, got %v", err)
	}
	if req.ID == nil || *req.ID != "block_abc" {
		t.Fatalf("id = %#v, want block_abc", req.ID)
	}
}

// An explicit null is the caller leaving a field unset, not a type error.
func TestParseBlockCreateAllowsExplicitNulls(t *testing.T) {
	obj := validExtractBlockFile()
	obj["parent_id"] = nil
	obj["width"] = nil

	if _, err := parseBlockCreate(obj); err != nil {
		t.Fatalf("explicit nulls should be accepted, got %v", err)
	}
}
