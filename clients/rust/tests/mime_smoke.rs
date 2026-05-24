// @oagen-ignore-file
//
// Compile-time smoke tests for the MimeData ergonomic input affordances.
// These never make a network call — they exist so `cargo check` proves
// that every From<T> -> MimeData impl threads through the generated
// resource-method constructors.

#![allow(unused_imports, dead_code, unused_variables)]

use std::path::PathBuf;

use retab::models::{CreateWorkflowRunRequest, FileRef};
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
}

#[test]
fn workflow_run_request_preserves_file_refs() {
    let mut request = CreateWorkflowRunRequest::new("wf_123");
    request.insert_file_ref(
        "start_doc",
        FileRef::new("file_123", "stored.pdf", "application/pdf"),
    );

    let json = serde_json::to_value(&request).unwrap();
    assert_eq!(
        json["documents"]["start_doc"]["id"],
        serde_json::json!("file_123")
    );
    assert_eq!(
        json["documents"]["start_doc"]["filename"],
        serde_json::json!("stored.pdf")
    );
    assert_eq!(
        json["documents"]["start_doc"]["mime_type"],
        serde_json::json!("application/pdf")
    );
}
