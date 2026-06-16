// @oagen-ignore-file
//
// Compile-time smoke tests for the MimeData ergonomic input affordances.
// These never make a network call — they exist so `cargo check` proves
// that every From<T> -> MimeData impl threads through the generated
// resource-method constructors.

#![allow(unused_imports, dead_code, unused_variables)]

use std::path::PathBuf;

use retab::models::CreateWorkflowRunRequest;
use retab::{MimeData, Retab};

fn _build_mime_from_pathbuf() {
    let _: MimeData = PathBuf::from("/tmp/example.pdf").into();
}

fn _build_mime_from_bytes() {
    let bytes: Vec<u8> = vec![1, 2, 3];
    let _: MimeData = bytes.into();
}

fn _build_mime_from_url_str() {
    let _: MimeData = "https://storage.retab.com/foo.pdf".into();
}

fn _client_builds() {
    let _client = Retab::new("test-key");
}

#[test]
fn workflow_run_request_inserts_mime_documents() {
    // The canonical POST /v1/workflows/runs route accepts MimeData
    // ({filename, url}) documents only; ergonomic inputs coerce through
    // Into<MimeData> and serialize to the {filename, url} shape.
    let request = CreateWorkflowRunRequest::new("wf_123")
        .with_document("start_doc", "https://example.com/invoice.pdf");

    let json = serde_json::to_value(&request).unwrap();
    assert_eq!(
        json["documents"]["start_doc"]["filename"],
        serde_json::json!("invoice.pdf")
    );
    assert_eq!(
        json["documents"]["start_doc"]["url"],
        serde_json::json!("https://example.com/invoice.pdf")
    );
    assert!(json["documents"]["start_doc"]["id"].is_null());
}
