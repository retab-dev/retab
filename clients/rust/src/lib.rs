// @oagen-ignore-file
//
// Crate root for the Retab Rust SDK. Hand-maintained scaffold — the
// generator wires up the spec-derived modules (`models`, `enums`,
// `resources`, `resources_api`) but does not regenerate this file.

#![allow(clippy::too_many_arguments)]
#![allow(clippy::useless_conversion)]
#![allow(clippy::doc_lazy_continuation)]

pub mod client;
pub mod enums;
pub mod error;
pub mod mime;
pub mod models;
pub mod pagination;
pub mod query;
pub mod resources;
pub mod resources_api;
pub mod secret;
pub mod workflow_run_documents;

pub use crate::client::{RequestOptions, Retab};
pub use crate::error::Error;
pub use crate::mime::MimeData;
pub use crate::secret::SecretString;
