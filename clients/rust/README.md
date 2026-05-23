<!-- @oagen-ignore-file -->

# retab — Rust client for the Retab API

```rust
use retab::Retab;

#[tokio::main]
async fn main() {
    let client = Retab::new(std::env::var("RETAB_API_KEY").unwrap());
    // client.extractions().list(...).await?
}
```

This crate is generated from the Retab OpenAPI spec. The HTTP client
(`src/client.rs`), error type (`src/error.rs`), pagination helper
(`src/pagination.rs`), MimeData ergonomics (`src/mime.rs`), and the
crate manifest (`Cargo.toml`) are hand-maintained — every other file
under `src/` is regenerated when the spec changes.

## MimeData input

Methods that accept a `MimeData` field also accept any type with an
`Into<MimeData>` impl — `PathBuf`, `Vec<u8>`, `url::Url`, `&str`
(URL or literal), or a pre-built `MimeData`.

```rust
use retab::MimeData;
use std::path::PathBuf;

let _: MimeData = PathBuf::from("/path/to/invoice.pdf").into();
let _: MimeData = "https://storage.retab.com/foo.pdf".into();
```
