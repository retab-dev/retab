// @oagen-ignore-file
//
// Hand-maintained MimeData ergonomics. The generator emits this on first
// run and then leaves it alone; spec changes do not touch it. Mirrors the
// ergonomic MimeData input handling from the Python (`prepare_mime_document`),
// Node (`coerceMimeData`), and Go (`InferMIMEData`) SDKs.

use std::path::{Path, PathBuf};

use base64::Engine;
use serde::{Deserialize, Serialize};

/// Wire-shape MimeData. Mirrors the spec's `MIMEData` component schema.
///
/// Customers rarely build this directly — they pass a `PathBuf`, byte
/// slice, URL, or string and the [`From`] impls below build the canonical
/// wire shape. Generated resource-method params constructors that accept
/// MimeData fields are bounded `<D: Into<MimeData>>` so any of these
/// inputs work transparently.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MimeData {
    pub filename: String,
    pub url: String,
}

impl MimeData {
    pub fn new(filename: impl Into<String>, url: impl Into<String>) -> Self {
        Self {
            filename: filename.into(),
            url: url.into(),
        }
    }
}

impl From<PathBuf> for MimeData {
    fn from(path: PathBuf) -> Self {
        let filename = path
            .file_name()
            .and_then(|s| s.to_str())
            .unwrap_or("document")
            .to_string();
        let bytes = std::fs::read(&path).unwrap_or_default();
        let mime = mime_guess::from_path(&path)
            .first_or_octet_stream()
            .to_string();
        let data_url = format!(
            "data:{};base64,{}",
            mime,
            base64::engine::general_purpose::STANDARD.encode(&bytes)
        );
        MimeData {
            filename,
            url: data_url,
        }
    }
}

impl From<&Path> for MimeData {
    fn from(p: &Path) -> Self {
        PathBuf::from(p).into()
    }
}

impl From<Vec<u8>> for MimeData {
    fn from(bytes: Vec<u8>) -> Self {
        let mime = infer::get(&bytes)
            .map(|t| t.mime_type().to_string())
            .unwrap_or_else(|| "application/octet-stream".to_string());
        let data_url = format!(
            "data:{};base64,{}",
            mime,
            base64::engine::general_purpose::STANDARD.encode(&bytes)
        );
        MimeData {
            filename: "document".to_string(),
            url: data_url,
        }
    }
}

impl From<&[u8]> for MimeData {
    fn from(bytes: &[u8]) -> Self {
        bytes.to_vec().into()
    }
}

impl From<url::Url> for MimeData {
    fn from(u: url::Url) -> Self {
        let filename = u
            .path_segments()
            .and_then(|s| s.last())
            .filter(|seg| !seg.is_empty())
            .unwrap_or("document")
            .to_string();
        MimeData {
            filename,
            url: u.to_string(),
        }
    }
}

impl From<&str> for MimeData {
    fn from(s: &str) -> Self {
        // Treat anything that parses as a URL as a URL; otherwise build a
        // text/plain data-URL with filename "document".
        match url::Url::parse(s) {
            Ok(u) => u.into(),
            Err(_) => MimeData {
                filename: "document".to_string(),
                url: format!(
                    "data:text/plain;base64,{}",
                    base64::engine::general_purpose::STANDARD.encode(s.as_bytes())
                ),
            },
        }
    }
}

impl From<String> for MimeData {
    fn from(s: String) -> Self {
        s.as_str().into()
    }
}
