package cmd

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
)

func apiErr(status int, msg string) error {
	return &retab.APIError{StatusCode: status, Message: msg}
}

func TestDeleteWithRetryOn404(t *testing.T) {
	orig := deleteRetryBaseDelay
	deleteRetryBaseDelay = time.Millisecond
	defer func() { deleteRetryBaseDelay = orig }()

	t.Run("succeeds immediately, no retry", func(t *testing.T) {
		calls := 0
		err := deleteWithRetryOn404(context.Background(), func() error { calls++; return nil })
		if err != nil || calls != 1 {
			t.Fatalf("err=%v calls=%d", err, calls)
		}
	})

	t.Run("retries on 404 then succeeds", func(t *testing.T) {
		calls := 0
		err := deleteWithRetryOn404(context.Background(), func() error {
			calls++
			if calls < 2 {
				return apiErr(http.StatusNotFound, "not found")
			}
			return nil
		})
		if err != nil || calls != 2 {
			t.Fatalf("err=%v calls=%d", err, calls)
		}
	})

	t.Run("persistent 404 surfaces after max attempts", func(t *testing.T) {
		calls := 0
		err := deleteWithRetryOn404(context.Background(), func() error {
			calls++
			return apiErr(http.StatusNotFound, "not found")
		})
		if apiErrorWithStatus(err, http.StatusNotFound) == nil {
			t.Fatalf("want 404, got %v", err)
		}
		if calls != 3 {
			t.Fatalf("want 3 attempts, got %d", calls)
		}
	})

	t.Run("non-404 fails fast without retry", func(t *testing.T) {
		calls := 0
		err := deleteWithRetryOn404(context.Background(), func() error {
			calls++
			return apiErr(http.StatusInternalServerError, "boom")
		})
		if apiErrorWithStatus(err, http.StatusInternalServerError) == nil {
			t.Fatalf("want 500, got %v", err)
		}
		if calls != 1 {
			t.Fatalf("want 1 attempt (no retry), got %d", calls)
		}
	})
}

func TestHintInvalidConfigFields(t *testing.T) {
	t.Run("nil passes through", func(t *testing.T) {
		if hintInvalidConfigFields(nil, true) != nil {
			t.Fatal("nil should stay nil")
		}
	})

	t.Run("non-APIError passes through unchanged", func(t *testing.T) {
		in := errors.New("plain")
		if got := hintInvalidConfigFields(in, true); got != in {
			t.Fatalf("want same error, got %v", got)
		}
	})

	t.Run("422 without the marker is unchanged", func(t *testing.T) {
		in := apiErr(http.StatusUnprocessableEntity, "some other validation error")
		got := hintInvalidConfigFields(in, true)
		if strings.Contains(got.Error(), "Tip:") {
			t.Fatalf("should not annotate: %v", got)
		}
	})

	t.Run("422 invalid config fields gets a merge tip", func(t *testing.T) {
		in := apiErr(http.StatusUnprocessableEntity, "Invalid config fields for 'extract' block: image_resolution_dpi")
		got := hintInvalidConfigFields(in, true).Error()
		if !strings.Contains(got, "Tip:") || !strings.Contains(got, "null") {
			t.Fatalf("merge tip missing: %q", got)
		}
	})

	t.Run("422 invalid config fields gets a replace tip", func(t *testing.T) {
		in := apiErr(http.StatusUnprocessableEntity, "Invalid config fields for 'extract' block: image_resolution_dpi")
		got := hintInvalidConfigFields(in, false).Error()
		if !strings.Contains(got, "--config-file") {
			t.Fatalf("replace tip missing: %q", got)
		}
	})
}
