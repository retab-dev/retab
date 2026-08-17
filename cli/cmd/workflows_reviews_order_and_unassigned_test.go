package cmd

import "testing"

// Two contract gaps found driving `retab workflows reviews` against staging.
//
// 1. GET /v1/workflows/reviews takes `order` (asc|desc over created_at) and both the
//    MCP tool and the human_review prompt tell an agent to pass desc for the newest
//    reviews — but the CLI registered no --order flag, so collectListParams could
//    never populate it and only the asc default was reachable.
// 2. `reviews schema` advertised the for_each snapshot as partitions-only with
//    additionalProperties:false, while the split_by_key seed version carries
//    `unassigned_pages` and the server accepts it. Following the advertised schema
//    meant DELETING the list a reviewer had just read back — silently rewriting
//    "page 8 is unassigned" into "nothing is unassigned", which is exactly the
//    signal the any_pages_unassigned predicate reads.

func TestReviewsListExposesOrderFlag(t *testing.T) {
	flag := workflowsReviewsListCmd.Flags().Lookup("order")
	if flag == nil {
		t.Fatal("reviews list must expose --order; the route takes order=asc|desc and the docs tell agents to use it")
	}
	if err := flag.Value.Set("desc"); err != nil {
		t.Fatalf("--order desc must be accepted: %v", err)
	}
	if err := flag.Value.Set("asc"); err != nil {
		t.Fatalf("--order asc must be accepted: %v", err)
	}
	if err := flag.Value.Set("sideways"); err == nil {
		t.Fatal("--order must reject a value the route's enum does not allow")
	}
}

func TestForEachReviewSchemaAllowsUnassignedPages(t *testing.T) {
	schema, err := reviewSchemaForBlockType("for_each")
	if err != nil {
		t.Fatalf("for_each snapshot schema must resolve: %v", err)
	}
	properties, ok := schema.Snapshot["properties"].(map[string]any)
	if !ok {
		t.Fatalf("for_each snapshot schema must declare properties; got %#v", schema.Snapshot)
	}
	if _, ok := properties["unassigned_pages"]; !ok {
		t.Fatalf("for_each snapshot must allow unassigned_pages (the seed version carries it); properties=%v", keysOfForEachSchema(properties))
	}
	// Optional: a reviewer whose split assigned every page must not be forced to
	// invent the key, and the empty list must stay legal.
	required, _ := schema.Snapshot["required"].([]string)
	for _, name := range required {
		if name == "unassigned_pages" {
			t.Fatal("unassigned_pages must stay optional")
		}
	}
	items, ok := properties["unassigned_pages"].(map[string]any)
	if !ok || items["type"] != "array" {
		t.Fatalf("unassigned_pages must be an array of page numbers; got %#v", properties["unassigned_pages"])
	}
}

func keysOfForEachSchema(properties map[string]any) []string {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	return names
}
