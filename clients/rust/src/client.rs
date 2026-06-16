// @oagen-ignore-file
//
// Retab HTTP client. Hand-maintained — the generator emits this on first
// run and then leaves it alone. The spec-derived resource modules
// (`src/resources/*.rs`) reference the methods defined here (e.g.
// `request_with_body_opts`); changing a signature here means rerunning
// generation.

use std::sync::Arc;
use std::time::Duration;

use http::{HeaderName, HeaderValue, Method, StatusCode};
use reqwest::Client as HttpClient;
use serde::{de::DeserializeOwned, Serialize};
use serde_json::Value;

use crate::error::Error;
use crate::pagination::{PageEnvelope, PaginatedList};

const DEFAULT_BASE_URL: &str = "https://api.retab.com";
const DEFAULT_TIMEOUT_SECS: u64 = 60;

/// Trim a trailing `/v<digits>` (and any trailing slashes) from a base URL.
/// The SDK's spec-derived paths already include `/v1`, so accepting bases
/// like `https://api.retab.com/v1` or `http://localhost:4000/v1` produces
/// `/v1/v1/...` 404s otherwise. Mirrors the Go SDK's WithBaseURL behaviour.
fn normalize_base_url(s: &str) -> String {
    let trimmed = s.trim_end_matches('/');
    if let Some(idx) = trimmed.rfind("/v") {
        let suffix = &trimmed[idx + 2..];
        if !suffix.is_empty() && suffix.chars().all(|c| c.is_ascii_digit()) {
            return trimmed[..idx].to_string();
        }
    }
    trimmed.to_string()
}

/// Public Retab API client. Construct via [`Retab::new`] or
/// [`Retab::builder`].
#[derive(Clone)]
pub struct Retab {
    inner: Arc<Inner>,
}

struct Inner {
    http: HttpClient,
    api_key: String,
    #[allow(dead_code)]
    base_url: String,
    #[allow(dead_code)]
    client_id: String,
}

/// Per-request overrides (extra headers, idempotency keys, …) layered on
/// top of the client defaults. Threaded through every `_with_options`
/// variant emitted by the generator.
#[derive(Default, Clone, Debug)]
pub struct RequestOptions {
    pub extra_headers: Vec<(HeaderName, HeaderValue)>,
    pub idempotency_key: Option<String>,
}

/// Builder for [`Retab`]. Use when you need to override the
/// default base URL, timeout, or HTTP client.
pub struct RetabBuilder {
    api_key: Option<String>,
    base_url: Option<String>,
    client_id: Option<String>,
    timeout: Option<Duration>,
    http: Option<HttpClient>,
}

impl RetabBuilder {
    pub fn api_key(mut self, key: impl Into<String>) -> Self {
        self.api_key = Some(key.into());
        self
    }
    pub fn base_url(mut self, url: impl Into<String>) -> Self {
        self.base_url = Some(normalize_base_url(&url.into()));
        self
    }
    pub fn client_id(mut self, id: impl Into<String>) -> Self {
        self.client_id = Some(id.into());
        self
    }
    pub fn timeout(mut self, t: Duration) -> Self {
        self.timeout = Some(t);
        self
    }
    pub fn http_client(mut self, http: HttpClient) -> Self {
        self.http = Some(http);
        self
    }
    pub fn build(self) -> Retab {
        let timeout = self
            .timeout
            .unwrap_or(Duration::from_secs(DEFAULT_TIMEOUT_SECS));
        let http = self.http.unwrap_or_else(|| {
            HttpClient::builder()
                .timeout(timeout)
                .build()
                .expect("retab: reqwest client build failed")
        });
        Retab {
            inner: Arc::new(Inner {
                http,
                api_key: self.api_key.unwrap_or_default(),
                base_url: self
                    .base_url
                    .unwrap_or_else(|| DEFAULT_BASE_URL.to_string()),
                client_id: self.client_id.unwrap_or_default(),
            }),
        }
    }
}

impl Retab {
    /// Construct a Retab client with the given API key, default base URL,
    /// and a fresh `reqwest::Client`. For more control use
    /// [`Retab::builder`].
    pub fn new(api_key: impl Into<String>) -> Self {
        Self::builder().api_key(api_key).build()
    }

    pub fn builder() -> RetabBuilder {
        RetabBuilder {
            api_key: None,
            base_url: None,
            client_id: None,
            timeout: None,
            http: None,
        }
    }

    #[allow(dead_code)]
    pub(crate) fn api_key(&self) -> &str {
        &self.inner.api_key
    }

    #[allow(dead_code)]
    pub(crate) fn client_id(&self) -> &str {
        &self.inner.client_id
    }

    #[allow(dead_code)]
    pub fn base_url(&self) -> &str {
        &self.inner.base_url
    }

    /// JSON-bodied request with an optional body. Used by every generated
    /// resource method that has a request body.
    pub(crate) async fn request_with_body_opts<Q, B, R>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        body: Option<&B>,
        options: Option<&RequestOptions>,
    ) -> Result<R, Error>
    where
        Q: Serialize + ?Sized,
        B: Serialize + ?Sized,
        R: DeserializeOwned,
    {
        let bytes = self.send(method, path, query, body, options).await?;
        serde_json::from_slice(&bytes).map_err(|e| Error::Deserialize(e.to_string()))
    }

    /// Variant returning `()` for endpoints with no response body.
    #[allow(dead_code)]
    pub(crate) async fn request_with_body_opts_empty<Q, B>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        body: Option<&B>,
        options: Option<&RequestOptions>,
    ) -> Result<(), Error>
    where
        Q: Serialize + ?Sized,
        B: Serialize + ?Sized,
    {
        self.send(method, path, query, body, options).await?;
        Ok(())
    }

    /// `multipart/form-data` request — used by upload endpoints whose request
    /// body the gateway expects as a multipart form (e.g. table create/replace).
    /// Each field is appended to a [`reqwest::multipart::Form`]; binary fields
    /// become file parts and the rest become text parts. `reqwest` sets the
    /// `Content-Type: multipart/form-data; boundary=...` header itself, so we
    /// must NOT set it manually.
    #[allow(dead_code)]
    pub(crate) async fn request_multipart_opts<Q, R>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        fields: Vec<MultipartField>,
        options: Option<&RequestOptions>,
    ) -> Result<R, Error>
    where
        Q: Serialize + ?Sized,
        R: DeserializeOwned,
    {
        let bytes = self
            .send_multipart(method, path, query, fields, options)
            .await?;
        serde_json::from_slice(&bytes).map_err(|e| Error::Deserialize(e.to_string()))
    }

    /// Body-less request — used by every GET / DELETE-style endpoint.
    pub(crate) async fn request_with_query_opts<Q, R>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        options: Option<&RequestOptions>,
    ) -> Result<R, Error>
    where
        Q: Serialize + ?Sized,
        R: DeserializeOwned,
    {
        let bytes = self
            .send::<Q, ()>(method, path, query, None, options)
            .await?;
        serde_json::from_slice(&bytes).map_err(|e| Error::Deserialize(e.to_string()))
    }

    /// Execute a cursor-paginated request and return a stream-capable page.
    pub(crate) async fn request_page<Q, T>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        cursor_param: &'static str,
        options: Option<&RequestOptions>,
    ) -> Result<PaginatedList<T>, Error>
    where
        Q: Serialize + ?Sized,
        T: DeserializeOwned + Send + Unpin + 'static,
    {
        let page: PageEnvelope<T> = self
            .request_with_query_opts(method.clone(), path, query, options)
            .await?;
        let client = self.clone();
        let path = path.to_string();
        let options = options.cloned();
        let query_template = normalized_query_value(query)?;

        Ok(PaginatedList::new(
            page.data,
            page.list_metadata,
            Some(Box::new(move |cursor| {
                let client = client.clone();
                let path = path.clone();
                let options = options.clone();
                let method = method.clone();
                let query = query_with_cursor(query_template.clone(), cursor_param, cursor);
                Box::pin(async move {
                    client
                        .request_with_query_opts(method, &path, &query, options.as_ref())
                        .await
                })
            })),
        ))
    }

    pub(crate) async fn request_with_query_opts_empty<Q>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        options: Option<&RequestOptions>,
    ) -> Result<(), Error>
    where
        Q: Serialize + ?Sized,
    {
        self.send::<Q, ()>(method, path, query, None, options)
            .await?;
        Ok(())
    }

    async fn send<Q, B>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        body: Option<&B>,
        options: Option<&RequestOptions>,
    ) -> Result<Vec<u8>, Error>
    where
        Q: Serialize + ?Sized,
        B: Serialize + ?Sized,
    {
        let qs = crate::query::encode_query(query)?;
        let url = if qs.is_empty() {
            format!("{}{}", self.inner.base_url, path)
        } else {
            format!("{}{}?{}", self.inner.base_url, path, qs)
        };

        let mut req = self.inner.http.request(method, &url);
        if !self.inner.api_key.is_empty() {
            req = req.header("Authorization", format!("Bearer {}", self.inner.api_key));
        }
        if let Some(opts) = options {
            for (k, v) in &opts.extra_headers {
                req = req.header(k.clone(), v.clone());
            }
            if let Some(key) = &opts.idempotency_key {
                req = req.header("Idempotency-Key", key);
            }
        }
        if let Some(b) = body {
            req = req.json(b);
        }

        let resp = req.send().await?;
        let status = resp.status();
        let bytes = resp.bytes().await?.to_vec();
        if !status.is_success() {
            return Err(Error::Api {
                status,
                message: status
                    .canonical_reason()
                    .unwrap_or("HTTP error")
                    .to_string(),
                body: String::from_utf8(bytes).ok(),
            });
        }
        let _: StatusCode = status;
        Ok(bytes)
    }

    async fn send_multipart<Q>(
        &self,
        method: Method,
        path: &str,
        query: &Q,
        fields: Vec<MultipartField>,
        options: Option<&RequestOptions>,
    ) -> Result<Vec<u8>, Error>
    where
        Q: Serialize + ?Sized,
    {
        let qs = crate::query::encode_query(query)?;
        let url = if qs.is_empty() {
            format!("{}{}", self.inner.base_url, path)
        } else {
            format!("{}{}?{}", self.inner.base_url, path, qs)
        };

        // Build the multipart form. Binary fields become file parts (with the
        // field name reused as the filename — the gateway keys off the part
        // name, not the filename); every other field is a plain text part.
        let mut form = reqwest::multipart::Form::new();
        for field in fields {
            match field {
                MultipartField::Text { name, value } => {
                    form = form.text(name, value);
                }
                MultipartField::File { name, bytes } => {
                    let part = reqwest::multipart::Part::bytes(bytes).file_name(name.clone());
                    form = form.part(name, part);
                }
            }
        }

        let mut req = self.inner.http.request(method, &url);
        if !self.inner.api_key.is_empty() {
            req = req.header("Authorization", format!("Bearer {}", self.inner.api_key));
        }
        if let Some(opts) = options {
            for (k, v) in &opts.extra_headers {
                req = req.header(k.clone(), v.clone());
            }
            if let Some(key) = &opts.idempotency_key {
                req = req.header("Idempotency-Key", key);
            }
        }
        // `.multipart(form)` sets the `Content-Type: multipart/form-data;
        // boundary=...` header itself — do NOT set it manually.
        req = req.multipart(form);

        let resp = req.send().await?;
        let status = resp.status();
        let bytes = resp.bytes().await?.to_vec();
        if !status.is_success() {
            return Err(Error::Api {
                status,
                message: status
                    .canonical_reason()
                    .unwrap_or("HTTP error")
                    .to_string(),
                body: String::from_utf8(bytes).ok(),
            });
        }
        let _: StatusCode = status;
        Ok(bytes)
    }
}

/// One field of a `multipart/form-data` request body. Generated upload
/// methods build a `Vec<MultipartField>` and hand it to
/// [`Retab::request_multipart_opts`]. Binary fields (the file
/// bytes) map to [`MultipartField::File`]; all other scalar fields map to
/// [`MultipartField::Text`].
#[allow(dead_code)]
pub(crate) enum MultipartField {
    /// A plain text form field (`name=value`).
    Text { name: String, value: String },
    /// A file part carrying raw bytes; the field name is reused as the filename.
    File { name: String, bytes: Vec<u8> },
}

/// Percent-encode a path segment for safe interpolation into a URL.
/// Generated resource methods call this on every `{path_param}` slot.
pub fn path_segment(s: &str) -> String {
    use percent_encoding::{utf8_percent_encode, NON_ALPHANUMERIC};
    utf8_percent_encode(s, NON_ALPHANUMERIC).to_string()
}

fn normalized_query_value<Q>(query: &Q) -> Result<Value, Error>
where
    Q: Serialize + ?Sized,
{
    let mut value = serde_json::to_value(query)
        .map_err(|e| Error::Builder(format!("query encode failed: {e}")))?;
    strip_nulls(&mut value);
    Ok(value)
}

fn strip_nulls(value: &mut Value) {
    match value {
        Value::Object(map) => {
            map.retain(|_, v| !v.is_null());
            for v in map.values_mut() {
                strip_nulls(v);
            }
        }
        Value::Array(items) => {
            for item in items {
                strip_nulls(item);
            }
        }
        _ => {}
    }
}

fn query_with_cursor(mut query: Value, cursor_param: &str, cursor: String) -> Value {
    if let Value::Object(map) = &mut query {
        map.insert(cursor_param.to_string(), Value::String(cursor));
        if cursor_param == "after" {
            map.remove("before");
        } else if cursor_param == "before" {
            map.remove("after");
        }
    }
    query
}
