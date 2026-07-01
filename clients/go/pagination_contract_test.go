package retab

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

// TestPaginationContract is a runtime-introspection regression test that
// catches future drift away from the centralized doPaginated[T] helper. See
// docs/blueprints/sdk-pagination-contract.md for the contract this enforces.
//
// It walks every *Service exposed on *Client via reflection, finds all
// `.List(...)` methods returning `*PaginatedList[T]`, and asserts each one
// wires the fetchNext closure when called against a stubbed transport.
//
// Two assertions per method:
//  1. The unexported fetchNext field on the returned page is non-nil
//     (direct evidence; reflection reads unexported fields via unsafe).
//  2. Calling AutoPaging issues a second HTTP request (indirect evidence
//     that confirms the closure actually re-issues the request).
//
// knownBypass below is the explicit allowlist of List methods that
// legitimately bypass the helper (dual-shape envelope-or-array decoders).
// Any new entry requires updating the blueprint's "Acceptable exceptions"
// section in lockstep.
func TestPaginationContract(t *testing.T) {
	// Allowlist: List methods that intentionally bypass doPaginated[T].
	// Any new entry here MUST also be documented in
	// docs/blueprints/sdk-pagination-contract.md under "Acceptable exceptions".
	knownBypass := map[string]string{
		"Workflows.Artifacts.List": "dual-shape (envelope OR bare array) — uses decodeArtifactListResponse",
		"Workflows.Blocks.List":    "dual-shape envelope-or-array decoder",
	}

	// Two-page server stub: the first request gets a page with after="cursor-2",
	// the second gets a terminal page. Counter lets us assert that AutoPaging
	// triggered a second HTTP fetch via the closure.
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":          []map[string]any{{"id": "sentinel_id_1"}},
				"list_metadata": map[string]any{"before": "", "after": "cursor-2"},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []map[string]any{{"id": "sentinel_id_2"}},
			"list_metadata": map[string]any{"before": "", "after": ""},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	listMethods := discoverListMethods(t, reflect.ValueOf(client), "")
	if len(listMethods) == 0 {
		t.Fatal("discovered 0 List methods on *Client; reflection walk is broken")
	}

	for _, lm := range listMethods {
		lm := lm
		t.Run(lm.path, func(t *testing.T) {
			atomic.StoreInt32(&calls, 0)

			page, err := callListMethod(t, lm)
			if err != nil {
				t.Fatalf("call %s: %v", lm.path, err)
			}
			if page.IsNil() {
				t.Fatalf("%s returned nil *PaginatedList", lm.path)
			}

			// HasNextPage proves the envelope was parsed correctly off the
			// stub server — the response we encoded sets after="cursor-2".
			if !page.MethodByName("HasNextPage").Call(nil)[0].Bool() {
				t.Fatalf("%s: HasNextPage()=false; expected true (envelope decode broken?)", lm.path)
			}

			// Direct evidence: the unexported fetchNext field on the
			// returned page. Reflection reads it via unsafe so we can
			// detect a missing closure even before AutoPaging runs.
			fetchNextField := page.Elem().FieldByName("fetchNext")
			if !fetchNextField.IsValid() {
				t.Fatalf("%s: fetchNext field not found on returned PaginatedList[T]", lm.path)
			}
			// FieldByName on an unexported field returns a non-settable
			// value whose .IsNil() still works for func types.
			fetchNextIsNil := fetchNextField.IsNil()

			if reason, allowed := knownBypass[lm.path]; allowed {
				// For known bypass methods, fetchNext is expected to be
				// nil (no auto-pagination) — AutoPaging silently stops
				// after the first page. We don't assert second-fetch.
				if !fetchNextIsNil {
					t.Logf("%s is on the bypass allowlist (%s) but fetchNext is now wired — "+
						"that's an improvement; consider dropping it from knownBypass and the blueprint.",
						lm.path, reason)
				}
				return
			}

			// Closure invariant: every page returned from a live .List call
			// MUST have its fetchNext wired. Both pieces of evidence below.
			if fetchNextIsNil {
				t.Fatalf("%s: fetchNext is nil on returned page — bypass of doPaginated[T] "+
					"(if intentional, add to knownBypass and document in the blueprint)", lm.path)
			}

			// Indirect evidence: AutoPaging must trigger a second HTTP call.
			// The stub already returned page 1 (calls=1); driving AutoPaging
			// to completion should fetch page 2 (calls=2).
			autoPagingArgs, ok := buildAutoPagingArgs(page)
			if !ok {
				t.Fatalf("%s: could not build AutoPaging args for page type %s", lm.path, page.Type())
			}
			autoPagingResult := page.MethodByName("AutoPaging").Call(autoPagingArgs)
			if !autoPagingResult[0].IsNil() {
				t.Fatalf("%s: AutoPaging error: %v", lm.path, autoPagingResult[0].Interface())
			}
			got := atomic.LoadInt32(&calls)
			if got != 2 {
				t.Fatalf("%s: expected 2 HTTP calls (initial + 1 auto-page), got %d — "+
					"closure didn't re-issue the request", lm.path, got)
			}
		})
	}
}

// discoveredListMethod is one List method discovered via reflection.
type discoveredListMethod struct {
	// path is the dotted accessor from *Client, e.g. "Workflows.Evals.Runs.List".
	path string
	// receiver is the *Service value the method is on.
	receiver reflect.Value
	// method is the List Method on receiver.
	method reflect.Value
	// methodType is method.Type() (a func type with the receiver already bound).
	methodType reflect.Type
}

// discoverListMethods walks every field of v recursively, finding all *Service
// structs and collecting their `List` methods that return (*PaginatedList[T], error).
func discoverListMethods(t *testing.T, v reflect.Value, prefix string) []discoveredListMethod {
	t.Helper()
	var out []discoveredListMethod

	// Deref pointer.
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)
		if !fieldType.IsExported() {
			continue
		}
		// Only care about pointer-to-struct service fields.
		if field.Kind() != reflect.Pointer || field.IsNil() {
			continue
		}
		elemType := field.Type().Elem()
		if elemType.Kind() != reflect.Struct {
			continue
		}
		typeName := elemType.Name()
		if !strings.HasSuffix(typeName, "Service") {
			continue
		}

		servicePath := fieldType.Name
		if prefix != "" {
			servicePath = prefix + "." + servicePath
		}

		// Look for a `List` method that returns (*PaginatedList[T], error).
		listMethod := field.MethodByName("List")
		if listMethod.IsValid() && isPaginatedListMethod(listMethod.Type()) {
			out = append(out, discoveredListMethod{
				path:       servicePath + ".List",
				receiver:   field,
				method:     listMethod,
				methodType: listMethod.Type(),
			})
		}

		// Recurse into nested services.
		out = append(out, discoverListMethods(t, field, servicePath)...)
	}
	return out
}

// isPaginatedListMethod reports whether mt has the shape
// `func(ctx, ...args) (*PaginatedList[T], error)`.
func isPaginatedListMethod(mt reflect.Type) bool {
	if mt.Kind() != reflect.Func {
		return false
	}
	if mt.NumOut() != 2 {
		return false
	}
	ret0 := mt.Out(0)
	if ret0.Kind() != reflect.Pointer {
		return false
	}
	if !strings.HasPrefix(ret0.Elem().Name(), "PaginatedList[") {
		return false
	}
	// Out(1) should be error.
	errInterface := reflect.TypeOf((*error)(nil)).Elem()
	return mt.Out(1) == errInterface
}

// callListMethod constructs minimal args for the discovered List method and
// invokes it. The shape of args varies — some take *Params, some take
// `string workflowID, *Params`, some take just a `string` or `string, int`.
// Variadic opts are skipped.
func callListMethod(t *testing.T, lm discoveredListMethod) (reflect.Value, error) {
	t.Helper()
	mt := lm.methodType

	nFixed := mt.NumIn()
	if mt.IsVariadic() {
		nFixed-- // drop the variadic slice
	}

	args := make([]reflect.Value, nFixed)
	for i := 0; i < nFixed; i++ {
		inType := mt.In(i)
		args[i] = buildArg(inType)
	}

	results := lm.method.Call(args)
	if !results[1].IsNil() {
		return reflect.Value{}, results[1].Interface().(error)
	}
	return results[0], nil
}

// buildArg returns a non-zero argument value for the given parameter type.
// Special cases:
//   - context.Context → context.Background()
//   - string → "sentinel"
//   - int → 0 (default limit branch in WorkflowExperimentRunResultsService.List)
//   - *T where T is a struct → reflect.New(T) with required ID fields populated
//   - T where T is a struct → zero value with required ID fields populated
func buildArg(t reflect.Type) reflect.Value {
	ctxInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	if t == ctxInterface {
		return reflect.ValueOf(context.Background())
	}
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf("sentinel")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.Zero(t)
	case reflect.Pointer:
		v := reflect.New(t.Elem())
		populateRequiredFields(v.Elem())
		return v
	case reflect.Struct:
		v := reflect.New(t).Elem()
		populateRequiredFields(v)
		return v
	default:
		return reflect.Zero(t)
	}
}

// populateRequiredFields fills in the string fields that a handful of List
// methods sanity-check before issuing the request (`RunID`, `BlockID`,
// `ReviewID`, `WorkflowID`). Without this, those methods return an error
// before the stub server is ever called.
func populateRequiredFields(v reflect.Value) {
	if v.Kind() != reflect.Struct {
		return
	}
	for _, name := range []string{"RunID", "BlockID", "ReviewID", "WorkflowID"} {
		f := v.FieldByName(name)
		if f.IsValid() && f.Kind() == reflect.String && f.CanSet() {
			f.SetString("sentinel")
		}
	}
}

// buildAutoPagingArgs returns the (ctx, yield) args for AutoPaging on the
// given *PaginatedList[T] reflect.Value. yield is constructed dynamically as
// a func(T) error that always returns nil, so we walk every item without
// short-circuiting.
func buildAutoPagingArgs(page reflect.Value) ([]reflect.Value, bool) {
	autoPaging := page.MethodByName("AutoPaging")
	if !autoPaging.IsValid() {
		return nil, false
	}
	mt := autoPaging.Type()
	// Signature: func(ctx context.Context, yield func(item T) error) error
	if mt.NumIn() != 2 {
		return nil, false
	}
	yieldType := mt.In(1)
	if yieldType.Kind() != reflect.Func {
		return nil, false
	}
	yield := reflect.MakeFunc(yieldType, func(args []reflect.Value) []reflect.Value {
		nilErr := reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
		return []reflect.Value{nilErr}
	})
	return []reflect.Value{reflect.ValueOf(context.Background()), yield}, true
}

// TestPaginationContractReflectionCanReadFetchNext is a sanity check that
// confirms reflection-on-unexported-fields is wired correctly. If this
// test ever starts failing, the main contract test's fetchNext-nil
// assertion is silently a no-op and the indirect HTTP-count assertion
// is the only real guard.
func TestPaginationContractReflectionCanReadFetchNext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":          []map[string]any{{"id": "wf_1"}},
			"list_metadata": map[string]any{"before": "", "after": "next"},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.Workflows.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if page == nil {
		t.Fatal("nil page")
	}

	v := reflect.ValueOf(page).Elem().FieldByName("fetchNext")
	if !v.IsValid() {
		t.Fatal("fetchNext field missing")
	}
	if v.IsNil() {
		t.Fatal("fetchNext is nil on a page produced by Workflows.List — the canonical helper isn't wiring the closure")
	}
	// Compare against a hand-built page; fetchNext should be nil there.
	bare := &PaginatedList[Workflow]{}
	bareField := reflect.ValueOf(bare).Elem().FieldByName("fetchNext")
	if !bareField.IsNil() {
		t.Fatal("hand-built PaginatedList[Workflow] should have nil fetchNext")
	}
}
