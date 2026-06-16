// @oagen-ignore-file
//
use std::collections::HashMap;

use crate::mime::MimeData;
use crate::models::CreateWorkflowRunRequest;

// The canonical POST /v1/workflows/runs route accepts MimeData ({filename, url})
// documents only. Persisted FileRef inputs were removed from this surface; they
// are only accepted by the legacy compatibility route
// POST /v1/workflows/{workflow_id}/run.
impl CreateWorkflowRunRequest {
    /// Insert one start-document input, coercing path/string/bytes/URL values
    /// through the same `Into<MimeData>` ergonomics used by document endpoints.
    pub fn insert_document<D>(&mut self, block_id: impl Into<String>, document: D)
    where
        D: Into<MimeData>,
    {
        self.documents
            .get_or_insert_with(HashMap::new)
            .insert(block_id.into(), document.into());
    }

    /// Builder-style variant of [`insert_document`](Self::insert_document).
    pub fn with_document<D>(mut self, block_id: impl Into<String>, document: D) -> Self
    where
        D: Into<MimeData>,
    {
        self.insert_document(block_id, document);
        self
    }
}
