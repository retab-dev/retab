// @oagen-ignore-file
//
// Compile-time smoke tests for the MimeData ergonomic input affordances.
// These never make a network call — they exist so `cargo check` proves
// that every From<T> -> MimeData impl threads through the generated
// resource-method constructors.

#![allow(unused_imports, dead_code, unused_variables)]

use std::path::PathBuf;

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
