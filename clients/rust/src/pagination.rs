// @oagen-ignore-file
//
// Cursor-pagination primitives shared by every generated `list` method.

use std::future::Future;
use std::pin::Pin;
use std::task::{Context, Poll};

use futures_util::{stream, Stream};
use serde::Deserialize;

use crate::error::Error;
use crate::models::ListMetadata;

#[derive(Debug, Deserialize)]
pub(crate) struct PageEnvelope<T> {
    pub data: Vec<T>,
    pub list_metadata: ListMetadata,
}

type FetchNext<T> = Box<
    dyn FnMut(String) -> Pin<Box<dyn Future<Output = Result<PageEnvelope<T>, Error>> + Send>>
        + Send,
>;

/// Cursor-paginated list returned by every generated `list` method.
///
/// The first page stays available through `data` and `list_metadata`.
/// Polling the value as a [`Stream`] yields those items, then lazily fetches
/// subsequent pages until `list_metadata.after` is exhausted.
pub struct PaginatedList<T> {
    pub data: Vec<T>,
    pub list_metadata: ListMetadata,
    fetch_next: Option<FetchNext<T>>,
    pending: Option<Pin<Box<dyn Future<Output = Result<PageEnvelope<T>, Error>> + Send>>>,
}

impl<T> std::fmt::Debug for PaginatedList<T>
where
    T: std::fmt::Debug,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("PaginatedList")
            .field("data", &self.data)
            .field("list_metadata", &self.list_metadata)
            .finish_non_exhaustive()
    }
}

impl<T> PaginatedList<T> {
    pub(crate) fn new(
        data: Vec<T>,
        list_metadata: ListMetadata,
        fetch_next: Option<FetchNext<T>>,
    ) -> Self {
        Self {
            data,
            list_metadata,
            fetch_next,
            pending: None,
        }
    }

    pub fn has_next_page(&self) -> bool {
        self.list_metadata.after.is_some()
    }
}

impl<T> Unpin for PaginatedList<T> {}

impl<T> Stream for PaginatedList<T>
where
    T: Unpin,
{
    type Item = Result<T, Error>;

    fn poll_next(self: Pin<&mut Self>, cx: &mut Context<'_>) -> Poll<Option<Self::Item>> {
        let this = self.get_mut();

        if !this.data.is_empty() {
            return Poll::Ready(Some(Ok(this.data.remove(0))));
        }

        loop {
            if let Some(pending) = this.pending.as_mut() {
                match pending.as_mut().poll(cx) {
                    Poll::Pending => return Poll::Pending,
                    Poll::Ready(Ok(page)) => {
                        this.pending = None;
                        this.data = page.data;
                        this.list_metadata = page.list_metadata;
                        if !this.data.is_empty() {
                            return Poll::Ready(Some(Ok(this.data.remove(0))));
                        }
                        continue;
                    }
                    Poll::Ready(Err(err)) => {
                        this.pending = None;
                        this.fetch_next = None;
                        return Poll::Ready(Some(Err(err)));
                    }
                }
            }

            let Some(after) = this.list_metadata.after.clone() else {
                return Poll::Ready(None);
            };
            let Some(fetch_next) = this.fetch_next.as_mut() else {
                return Poll::Ready(None);
            };
            this.pending = Some(fetch_next(after));
        }
    }
}

pub fn auto_paginate_pages<T, F, Fut>(load: F) -> impl Stream<Item = Result<T, Error>>
where
    F: FnMut(Option<String>) -> Fut,
    Fut: Future<Output = Result<(Vec<T>, Option<String>), Error>>,
{
    enum State<T, F> {
        Running {
            load: F,
            cursor: Option<String>,
            buffer: std::vec::IntoIter<T>,
            first: bool,
        },
        Done,
    }

    let state: State<T, F> = State::Running {
        load,
        cursor: None,
        buffer: Vec::new().into_iter(),
        first: true,
    };

    stream::unfold(state, |st| async move {
        match st {
            State::Done => None,
            State::Running {
                mut load,
                mut cursor,
                mut buffer,
                first,
            } => {
                if let Some(item) = buffer.next() {
                    return Some((
                        Ok(item),
                        State::Running {
                            load,
                            cursor,
                            buffer,
                            first: false,
                        },
                    ));
                }
                if !first && cursor.is_none() {
                    return None;
                }
                match load(cursor.take()).await {
                    Ok((page, next)) => {
                        cursor = next;
                        buffer = page.into_iter();
                        match buffer.next() {
                            Some(item) => Some((
                                Ok(item),
                                State::Running {
                                    load,
                                    cursor,
                                    buffer,
                                    first: false,
                                },
                            )),
                            None => None,
                        }
                    }
                    Err(e) => Some((Err(e), State::Done)),
                }
            }
        }
    })
}
