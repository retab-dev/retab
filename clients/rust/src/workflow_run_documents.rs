// @oagen-ignore-file
//
use std::collections::HashMap;

use crate::mime::MimeData;
use crate::models::{CreateWorkflowRunRequest, CreateWorkflowRunRequestDocumentsOneOf, FileRef};

impl CreateWorkflowRunRequest {
    /// Insert one start-document input, coercing path/string/bytes/URL values
    /// through the same `Into<MimeData>` ergonomics used by document endpoints.
    pub fn insert_document<D>(&mut self, block_id: impl Into<String>, document: D)
    where
        D: Into<MimeData>,
    {
        self.documents
            .get_or_insert_with(HashMap::new)
            .insert(block_id.into(), document.into().into());
    }

    /// Builder-style variant of [`insert_document`](Self::insert_document).
    pub fn with_document<D>(mut self, block_id: impl Into<String>, document: D) -> Self
    where
        D: Into<MimeData>,
    {
        self.insert_document(block_id, document);
        self
    }

    /// Insert one already-uploaded file reference for a start-document block.
    pub fn insert_file_ref(&mut self, block_id: impl Into<String>, file_ref: FileRef) {
        self.documents.get_or_insert_with(HashMap::new).insert(
            block_id.into(),
            CreateWorkflowRunRequestDocumentsOneOf::from(file_ref),
        );
    }

    /// Builder-style variant of [`insert_file_ref`](Self::insert_file_ref).
    pub fn with_file_ref(mut self, block_id: impl Into<String>, file_ref: FileRef) -> Self {
        self.insert_file_ref(block_id, file_ref);
        self
    }
}
