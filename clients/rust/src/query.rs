// @oagen-ignore-file
//
// Query-string encoder shared by every generated resource method.
// Custom comma-joined serializers mirror the `style: form, explode: false`
// shape some endpoints declare in the spec.

use serde::Serialize;

use crate::error::Error;

pub fn encode_query<T: Serialize + ?Sized>(value: &T) -> Result<String, Error> {
    serde_urlencoded::to_string(value)
        .map_err(|e| Error::Builder(format!("query encode failed: {e}")))
}

pub fn serialize_comma_separated<S, T>(items: &[T], serializer: S) -> Result<S::Ok, S::Error>
where
    S: serde::Serializer,
    T: std::fmt::Display,
{
    let joined: String = items
        .iter()
        .map(|x| x.to_string())
        .collect::<Vec<_>>()
        .join(",");
    serializer.serialize_str(&joined)
}

pub fn serialize_comma_separated_opt<S, T>(
    items: &Option<Vec<T>>,
    serializer: S,
) -> Result<S::Ok, S::Error>
where
    S: serde::Serializer,
    T: std::fmt::Display,
{
    match items {
        Some(v) => serialize_comma_separated(v, serializer),
        None => serializer.serialize_none(),
    }
}
