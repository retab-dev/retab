//go:build go1.23

package retab

import (
	"context"
	"errors"
	"iter"
)

// errStopAutoPagingSeq is returned from AutoPaging's yield callback when the
// range-over-func caller breaks out of the for-loop. It tells AutoPaging to
// halt without exposing the sentinel value at the call site.
var errStopAutoPagingSeq = errors.New("retab: AutoPagingSeq stopped by caller")

// AutoPagingSeq returns a Go 1.23+ iter.Seq2 over every item across every
// page, with errors surfaced through the (T, error) pair.
//
// Usage:
//
//	for item, err := range page.AutoPagingSeq(ctx) {
//	    if err != nil {
//	        return err
//	    }
//	    // ...
//	}
//
// Yields (item, nil) for each item; on a page-fetch failure mid-iteration,
// yields a single (zero T, err) pair and then stops. Caller `break`s
// short-circuit cleanly without surfacing an error.
//
// Wraps the callback-based AutoPaging so both APIs share the same cursor /
// fetch semantics. AutoPaging stays the supported API for callers that
// prefer explicit per-item error handling or that haven't bumped to Go 1.23.
func (p *PaginatedList[T]) AutoPagingSeq(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		err := p.AutoPaging(ctx, func(item T) error {
			if !yield(item, nil) {
				return errStopAutoPagingSeq
			}
			return nil
		})
		if err != nil && !errors.Is(err, errStopAutoPagingSeq) {
			var zero T
			yield(zero, err)
		}
	}
}
