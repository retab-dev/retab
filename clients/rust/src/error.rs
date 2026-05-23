// @oagen-ignore-file
//
// Retab error type. Hand-maintained — generated resource methods
// surface `Result<T, Error>`. Surface area is intentionally small;
// extend with structured API error parsing as needs arise.

use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("HTTP transport error: {0}")]
    Transport(#[from] reqwest::Error),

    #[error("API error {status}: {message}")]
    Api {
        status: http::StatusCode,
        message: String,
        body: Option<String>,
    },

    #[error("deserialization error: {0}")]
    Deserialize(String),

    #[error("request builder error: {0}")]
    Builder(String),
}
