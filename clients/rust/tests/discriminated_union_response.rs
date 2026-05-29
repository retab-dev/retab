// @oagen-ignore-file
//
// Round-trip tests for discriminated-union RESPONSES.
//
// Two response-position unions used to be typed as the untyped
// `serde_json::Value` instead of named, internally-tagged serde enums:
//   - GET /v1/workflows/artifacts/{artifact_id} -> 11-variant `operation` union
//   - GET /v1/workflows/experiments/metrics     -> 6-variant `kind` union
//
// These tests deserialize a NON-first variant of each union into the generated
// enum, re-serialize it, and assert no fields are dropped — proving the
// response is now losslessly typed rather than collapsing to the first variant
// or to an opaque map.

#![allow(unused_imports, dead_code)]

use retab::models::{
    GetExperimentMetricsForRunResponseOneOf, GetWorkflowArtifactByIdResponseOneOf,
};

/// A `split` artifact is the 2nd variant of the 11-arm `operation` union — a
/// non-first variant, so this exercises the exact bug where the response would
/// collapse to the first variant and silently drop the split-specific fields.
fn split_artifact_json() -> serde_json::Value {
    serde_json::json!({
        "operation": "split",
        "id": "artifact_split_123",
        "file": {
            "id": "file_abc",
            "filename": "invoice.pdf",
            "mime_type": "application/pdf"
        },
        "model": "gpt-4.1-nano",
        "subdocuments": [
            { "name": "page_one", "description": "first page", "allow_multiple_instances": false }
        ],
        "n_consensus": 3,
        "instructions": "split by invoice boundary",
        "output": [
            { "name": "invoice_a", "pages": [1, 2] },
            { "name": "invoice_b", "pages": [3] }
        ],
        "created_at": "2025-01-01T00:00:00Z"
    })
}

#[test]
fn split_artifact_response_round_trips_without_dropping_fields() {
    let input = split_artifact_json();

    // Deserialize into the typed enum (was `serde_json::Value` before the fix).
    let decoded: GetWorkflowArtifactByIdResponseOneOf =
        serde_json::from_value(input.clone()).expect("split artifact should decode into the enum");

    // The internally-tagged enum must pick the SplitWorkflowArtifact arm — not
    // the first arm (extraction). Confirms the discriminator is honored.
    assert!(
        matches!(
            decoded,
            GetWorkflowArtifactByIdResponseOneOf::SplitWorkflowArtifact(_)
        ),
        "split discriminator must select the SplitWorkflowArtifact arm",
    );

    // Re-serialize and assert every input field survives the round-trip.
    let reencoded = serde_json::to_value(&decoded).expect("re-serialize");
    let input_obj = input.as_object().unwrap();
    let out_obj = reencoded.as_object().unwrap();
    for (key, value) in input_obj {
        assert_eq!(
            out_obj.get(key),
            Some(value),
            "field `{key}` was dropped or altered during the union round-trip",
        );
    }
}

/// `by_document` is the 2nd variant of the 6-arm `kind` union for experiment
/// metrics — again a non-first variant.
fn by_document_metrics_json() -> serde_json::Value {
    serde_json::json!({
        "kind": "by_document",
        "run_id": "run_abc123",
        "view": "default",
        "document": { "id": "doc_1", "filename": "invoice.pdf" },
        "score": 0.87
    })
}

#[test]
fn experiment_metrics_response_decodes_a_non_first_variant() {
    let input = by_document_metrics_json();

    let decoded: GetExperimentMetricsForRunResponseOneOf =
        serde_json::from_value(input.clone()).expect("metrics should decode into the enum");

    assert!(
        matches!(
            decoded,
            GetExperimentMetricsForRunResponseOneOf::ExperimentByDocumentMetricsResponse(_)
        ),
        "by_document discriminator must select the ExperimentByDocumentMetricsResponse arm",
    );

    // The `kind` discriminator survives re-serialization.
    let reencoded = serde_json::to_value(&decoded).expect("re-serialize");
    assert_eq!(
        reencoded.get("kind").and_then(|v| v.as_str()),
        Some("by_document"),
    );
}
