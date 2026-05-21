package retab

import "testing"

func TestReviewWireConstantsMatchHardCutover(t *testing.T) {
	cases := []struct {
		wire     string
		constant string
	}{
		{"accepted", SubmissionStatusAccepted},
		{"already_applied", SubmissionStatusAlreadyApplied},
		{"accepted", AppendStatusAccepted},
		{"already_exists", AppendStatusAlreadyExists},
		{"pending", ResumeStatusPending},
		{"resumed", ResumeStatusResumed},
		{"skipped", ResumeStatusSkipped},
	}
	for _, tc := range cases {
		if tc.constant != tc.wire {
			t.Errorf("constant %q does not match wire value %q", tc.constant, tc.wire)
		}
	}
}
