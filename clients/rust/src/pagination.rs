// @oagen-ignore-file
//
// Cursor-pagination helper used by generated `*_auto_paging` methods.
// Returns an async [`futures_util::Stream`] that walks every page until
// the cursor is exhausted.
//
// No `'static` bound on the closure / future — the generated wrappers
// capture `&'a self` to call the underlying `*_with_options` method, and
// imposing `'static` would force them to clone the client (or worse, leak
// it). Lifetime inference threads through the returned `impl Stream`.

use std::future::Future;

use futures_util::stream::{self, Stream};

use crate::error::Error;

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
